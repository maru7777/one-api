package model

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/message"
	"gorm.io/gorm"
)

const (
	TokenStatusEnabled   = 1 // don't use 0, 0 is the default value!
	TokenStatusDisabled  = 2 // also don't use 0
	TokenStatusExpired   = 3
	TokenStatusExhausted = 4
)

type Token struct {
	Id             int     `json:"id"`
	UserId         int     `json:"user_id"`
	Key            string  `json:"key" gorm:"type:char(48);uniqueIndex"`
	Status         int     `json:"status" gorm:"default:1"`
	Name           string  `json:"name" gorm:"index" `
	CreatedTime    int64   `json:"created_time" gorm:"bigint"`
	AccessedTime   int64   `json:"accessed_time" gorm:"bigint"`
	ExpiredTime    int64   `json:"expired_time" gorm:"bigint;default:-1"` // -1 means never expired
	RemainQuota    int64   `json:"remain_quota" gorm:"bigint;default:0"`
	UnlimitedQuota bool    `json:"unlimited_quota" gorm:"default:false"`
	UsedQuota      int64   `json:"used_quota" gorm:"bigint;default:0"` // used quota
	Models         *string `json:"models" gorm:"type:text"`            // allowed models
	Subnet         *string `json:"subnet" gorm:"default:''"`           // allowed subnet
}

func GetAllUserTokens(userId int, startIdx int, num int, order string) ([]*Token, error) {
	var tokens []*Token
	var err error
	query := DB.Where("user_id = ?", userId)

	switch order {
	case "remain_quota":
		query = query.Order("unlimited_quota desc, remain_quota desc")
	case "used_quota":
		query = query.Order("used_quota desc")
	default:
		query = query.Order("id desc")
	}

	err = query.Limit(num).Offset(startIdx).Find(&tokens).Error
	return tokens, err
}

func SearchUserTokens(userId int, keyword string) (tokens []*Token, err error) {
	err = DB.Where("user_id = ?", userId).Where("name LIKE ?", keyword+"%").Find(&tokens).Error
	return tokens, err
}

func ValidateUserToken(key string) (token *Token, err error) {
	if key == "" {
		return nil, errors.New("No token provided")
	}
	token, err = CacheGetTokenByKey(key)
	if err != nil {
		logger.SysError("CacheGetTokenByKey failed: " + err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(err, "token not found")
		}

		return nil, errors.Wrap(err, "failed to get token by key")
	}
	if token.Status == TokenStatusExhausted {
		return nil, fmt.Errorf("API Key %s (#%d) quota has been exhausted", token.Name, token.Id)
	} else if token.Status == TokenStatusExpired {
		return nil, errors.New("The token has expired")
	}
	if token.Status != TokenStatusEnabled {
		return nil, errors.New("The token status is not available")
	}
	if token.ExpiredTime != -1 && token.ExpiredTime < helper.GetTimestamp() {
		if !common.RedisEnabled {
			token.Status = TokenStatusExpired
			err := token.SelectUpdate()
			if err != nil {
				logger.SysError("failed to update token status" + err.Error())
			}
		}
		return nil, errors.New("The token has expired")
	}
	if !token.UnlimitedQuota && token.RemainQuota <= 0 {
		if !common.RedisEnabled {
			// in this case, we can make sure the token is exhausted
			token.Status = TokenStatusExhausted
			err := token.SelectUpdate()
			if err != nil {
				logger.SysError("failed to update token status" + err.Error())
			}
		}
		return nil, errors.New("The token quota has been used up")
	}
	return token, nil
}

func GetTokenByIds(id int, userId int) (*Token, error) {
	if id == 0 || userId == 0 {
		return nil, errors.New("id or userId is empty!")
	}
	token := Token{Id: id, UserId: userId}
	var err error = nil
	err = DB.First(&token, "id = ? and user_id = ?", id, userId).Error
	return &token, err
}

func GetTokenById(id int) (*Token, error) {
	if id == 0 {
		return nil, errors.New("id is empty!")
	}
	token := Token{Id: id}
	var err error = nil
	err = DB.First(&token, "id = ?", id).Error
	return &token, err
}

func (t *Token) Insert() error {
	var err error
	err = DB.Create(t).Error
	return err
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (t *Token) Update() error {
	var err error
	err = DB.Model(t).Select("name", "status", "expired_time", "remain_quota", "unlimited_quota", "models", "subnet").Updates(t).Error
	return err
}

func (t *Token) SelectUpdate() error {
	// This can update zero values
	return DB.Model(t).Select("accessed_time", "status").Updates(t).Error
}

func (t *Token) Delete() error {
	var err error
	err = DB.Delete(t).Error
	return err
}

func (t *Token) GetModels() string {
	if t == nil {
		return ""
	}
	if t.Models == nil {
		return ""
	}
	return *t.Models
}

func DeleteTokenById(id int, userId int) (err error) {
	// Why we need userId here? In case user want to delete other's token.
	if id == 0 || userId == 0 {
		return errors.New("id or userId is empty!")
	}
	token := Token{Id: id, UserId: userId}
	err = DB.Where(token).First(&token).Error
	if err != nil {
		return err
	}
	// 检查Redis中是否有该用户邮箱关联的apikey
	if common.RedisEnabled && common.RDB1 != nil {
		// 被删除的token不见得和redis里的相同,相同时才将redis里的apikey置空
		var user User
		userErr := DB.Where("id = ?", userId).First(&user).Error
		if userErr != nil {
			// 用户可能已被删除，这是正常情况，不需要处理Redis
			logger.SysLog("查找用户失败，可能用户已被删除")
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			// 先检查key是否存在
			exists, redisErr := common.RDB1.Exists(ctx, user.Email).Result()
			if redisErr != nil || exists == 0 {
				logger.SysLog("Redis中未找到该邮箱对应的记录，可能已被删除")
			} else {
				// 获取存储在Redis中的apikey
				apikey, redisErr := common.RDB1.HGet(ctx, user.Email, "apikey").Result()
				if redisErr == nil && apikey == token.Key {
					// 如果Redis中存储的apikey与要删除的token.Key相同，则将其置空
					err := common.RDB1.HSet(ctx, user.Email, "apikey", "").Err()
					if err != nil {
						logger.SysError("清除Redis中的apikey失败: " + err.Error())
					} else {
						logger.SysLog("已成功清除Redis中用户的apikey: " + user.Email)
					}
				}
			}
		}
	}
	return token.Delete()
}

func IncreaseTokenQuota(id int, quota int64) (err error) {
	if quota < 0 {
		return errors.New("quota cannot be negative!")
	}
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeTokenQuota, id, quota)
		return nil
	}
	return increaseTokenQuota(id, quota)
}

func increaseTokenQuota(id int, quota int64) (err error) {
	err = DB.Model(&Token{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota + ?", quota),
			"used_quota":    gorm.Expr("used_quota - ?", quota),
			"accessed_time": helper.GetTimestamp(),
		},
	).Error
	return err
}

func DecreaseTokenQuota(id int, quota int64) (err error) {
	if quota < 0 {
		return errors.New("quota cannot be negative!")
	}
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeTokenQuota, id, -quota)
		return nil
	}
	return decreaseTokenQuota(id, quota)
}

func decreaseTokenQuota(id int, quota int64) (err error) {
	err = DB.Model(&Token{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota - ?", quota),
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"accessed_time": helper.GetTimestamp(),
		},
	).Error
	return err
}

func PreConsumeTokenQuota(tokenId int, quota int64) (err error) {
	if quota < 0 {
		return errors.New("quota cannot be negative!")
	}
	token, err := GetTokenById(tokenId)
	if err != nil {
		return err
	}
	if !token.UnlimitedQuota && token.RemainQuota < quota {
		return errors.New("Insufficient token quota")
	}
	userQuota, err := GetUserQuota(token.UserId)
	if err != nil {
		return err
	}
	if userQuota < quota {
		return errors.New("Insufficient user quota")
	}
	quotaTooLow := userQuota >= config.QuotaRemindThreshold && userQuota-quota < config.QuotaRemindThreshold
	noMoreQuota := userQuota-quota <= 0
	if quotaTooLow || noMoreQuota {
		go func() {
			email, err := GetUserEmail(token.UserId)
			if err != nil {
				logger.SysError("failed to fetch user email: " + err.Error())
			}
			prompt := "Quota Reminder"
			var contentText string
			if noMoreQuota {
				contentText = "Your quota has been exhausted"
			} else {
				contentText = "Your quota is about to be exhausted"
			}
			if email != "" {
				topUpLink := fmt.Sprintf("%s/topup", config.ServerAddress)
				content := message.EmailTemplate(
					prompt,
					fmt.Sprintf(`
								<p>Hello!</p>
								<p>%s, your current remaining quota is <strong>%d</strong>.</p>
								<p>To avoid any disruption to your service, please top up in a timely manner.</p>
								<p style="text-align: center; margin: 30px 0;">
									<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Top Up Now</a>
								</p>
								<p style="color: #666;">If the button does not work, please copy the following link and paste it into your browser:</p>
								<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px; word-break: break-all;">%s</p>
							`, contentText, userQuota, topUpLink, topUpLink),
				)
				err = message.SendEmail(prompt, email, content)
				if err != nil {
					logger.SysError("failed to send email: " + err.Error())
				}
			}
		}()
	}
	if !token.UnlimitedQuota {
		err = DecreaseTokenQuota(tokenId, quota)
		if err != nil {
			return err
		}
	}
	err = DecreaseUserQuota(token.UserId, quota)
	return err
}

func PostConsumeTokenQuota(tokenId int, quota int64) (err error) {
	token, err := GetTokenById(tokenId)
	if err != nil {
		return err
	}
	if quota > 0 {
		err = DecreaseUserQuota(token.UserId, quota)
	} else {
		err = IncreaseUserQuota(token.UserId, -quota)
	}
	if !token.UnlimitedQuota {
		if quota > 0 {
			err = DecreaseTokenQuota(tokenId, quota)
		} else {
			err = IncreaseTokenQuota(tokenId, -quota)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
