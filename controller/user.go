package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/i18n"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	if !config.PasswordLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "The administrator has turned off password login",
			"success": false,
		})
		return
	}
	var loginRequest LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&loginRequest)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": i18n.Translate(c, "invalid_parameter"),
			"success": false,
		})
		return
	}
	username := loginRequest.Username
	password := loginRequest.Password
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"message": i18n.Translate(c, "invalid_parameter"),
			"success": false,
		})
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}
	err = user.ValidateAndFill()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	SetupLogin(&user, c)
}

// setup session & cookies and then return user info
func SetupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	session.Options(sessions.Options{
		Path:     "/",  // 设置 cookie 在整个网站下都有效
		MaxAge:   0,    // 86400 * 14 有效期 14 天（单位：秒）
		HttpOnly: true, // 只允许后端访问，提高安全性
	})
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	err := session.Save()
	if err != nil {
		logger.Errorf(c.Request.Context(), "Unable to save login session information: %+v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "Unable to save login session information, please try again",
			"success": false,
		})
		return
	}

	// set auth header
	// c.Set("id", user.Id)
	// GenerateAccessToken(c)
	// c.Header("Authorization", user.AccessToken)

	cleanUser := model.User{
		Id:          user.Id,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
	}

	// 从Redis DB1中获取用户的apikey
	var apikey string = ""
	if common.RedisEnabled && common.RDB1 != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var exist_user model.User
		_ = model.DB.Where("id = ?", user.Id).First(&exist_user).Error
		// 使用用户邮箱作为键从Redis DB1中获取token
		apikey_redis, err := common.RDB1.HGet(ctx, user.Email, "apikey").Result()
		if err == nil && apikey_redis != "" {
			apikey = apikey_redis
		}
	}
	// 如果Redis中没有找到token，则从数据库中查询
	if apikey == "" {
		// 查询用户的第一个可用token
		var token model.Token
		err := model.DB.Where("user_id = ? AND status = ?", user.Id, model.TokenStatusEnabled).
			Order("id desc").
			Limit(1).
			First(&token).Error
		if err == nil {
			apikey = token.Key
			// 将查询到的token缓存到Redis中
			if common.RedisEnabled && common.RDB1 != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var exist_user model.User
				_ = model.DB.Where("id = ?", user.Id).First(&exist_user).Error
				err := common.RDB1.HSet(ctx, exist_user.Email, "apikey", token.Key).Err()
				if err != nil {
					logger.SysError("缓存 apikey 到 redis 失败,请检查: " + err.Error())
				}
			}
		} else {
			// 数据库中也没有找到token，创建一个新token
			key := random.GenerateKey()
			if key != "" {
				// 使用已有的模式创建新token，参考Register函数中的token创建逻辑
				token := model.Token{
					UserId:         user.Id,
					Name:           user.Username + " automaticlly generated token",
					Key:            key,
					CreatedTime:    time.Now().Unix(),
					AccessedTime:   time.Now().Unix(),
					ExpiredTime:    -1,        // 永不过期
					RemainQuota:    500000000, // 示例额度
					UnlimitedQuota: true,
				}
				if err := token.Insert(); err == nil {
					apikey = key
					// 如果Redis可用且用户有邮箱，将token保存到Redis中
					if common.RedisEnabled && common.RDB1 != nil {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						var exist_user model.User
						_ = model.DB.Where("id = ?", user.Id).First(&exist_user).Error
						err := common.RDB1.HSet(ctx, exist_user.Email, "apikey", key).Err()
						if err != nil {
							logger.SysError("缓存 apikey 到 redis 失败,请检查: " + err.Error())
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data":    cleanUser,
		"apikey":  apikey,
	})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
	})
}

func Register(c *gin.Context) {
	ctx := c.Request.Context()
	if !config.RegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "The administrator has turned off new user registration",
			"success": false,
		})
		return
	}
	if !config.PasswordRegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "The administrator has turned off registration via password. Please use the form of third-party account verification to register",
			"success": false,
		})
		return
	}
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}
	if config.EmailVerificationEnabled {
		if user.Email == "" || user.VerificationCode == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "The administrator has turned on email verification, please enter the email address and verification code",
			})
			return
		}
		if !common.VerifyCodeWithKey(user.Email, user.VerificationCode, common.EmailVerificationPurpose) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Verification code error or expired",
			})
			return
		}
	}
	affCode := user.AffCode // this code is the inviter's code, not the user's own code
	inviterId, _ := model.GetUserIdByAffCode(affCode)
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.Username,
		InviterId:   inviterId,
	}
	if config.EmailVerificationEnabled {
		cleanUser.Email = user.Email
	}
	if err := cleanUser.Insert(ctx, inviterId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func GetAllUsers(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.DefaultQuery("order", "")
	users, err := model.GetAllUsers(p*config.ItemsPerPage, config.ItemsPerPage, order)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    users,
	})
}

func SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	users, err := model.SearchUsers(keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    users,
	})
	return
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole <= user.Role && myRole != model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to get information of users at the same level or higher",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
	return
}

func GetUserDashboard(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	now := time.Now()
	startOfDay := now.Truncate(24*time.Hour).AddDate(0, 0, -6).Unix()
	endOfDay := now.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second).Unix()

	dashboards, err := model.SearchLogsByDayAndModel(id, int(startOfDay), int(endOfDay))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to get user dashboard data: " + err.Error(),
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    dashboards,
	})
	return
}

func GenerateAccessToken(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.AccessToken = random.GetUUID()

	if model.DB.Where("access_token = ?", user.AccessToken).First(user).RowsAffected != 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Please try again, the system-generated UUID is actually duplicated!",
		})
		return
	}

	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AccessToken,
	})
	return
}

func GetAffCode(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.AffCode == "" {
		user.AffCode = random.GetRandomString(4)
		if err := user.Update(false); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AffCode,
	})
	return
}

// GetSelfByToken get user by openai api token
func GetSelfByToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"uid":      c.GetInt("id"),
		"token_id": c.GetInt("token_id"),
		"username": c.GetString("username"),
	})
	return
}

func GetSelf(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
	return
}

func UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var updatedUser model.User
	err := json.NewDecoder(c.Request.Body).Decode(&updatedUser)
	if err != nil || updatedUser.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if updatedUser.Password == "" {
		updatedUser.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&updatedUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}
	originUser, err := model.GetUserById(updatedUser.Id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole <= originUser.Role && myRole != model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to update user information with the same permission level or higher permission level",
		})
		return
	}
	if myRole <= updatedUser.Role && myRole != model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to promote other users to a permission level greater than or equal to your own",
		})
		return
	}
	if updatedUser.Password == "$I_LOVE_U" {
		updatedUser.Password = "" // rollback to what it should be
	}
	updatePassword := updatedUser.Password != ""
	var existed_user model.User
	_ = model.DB.Where("id = ?", updatedUser.Id).First(&existed_user).Error
	if hashedNewPassword, err := common.Password2Hash(updatedUser.Password); err == nil && hashedNewPassword == existed_user.Password {
		updatePassword = false // 如果新密码与旧密码相同，则不更新密码
	}
	if err := updatedUser.Update(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if originUser.Quota != updatedUser.Quota {
		model.RecordLog(ctx, originUser.Id, model.LogTypeManage, fmt.Sprintf("Admin changed user quota from %s to %s", common.LogQuota(originUser.Quota), common.LogQuota(updatedUser.Quota)))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func UpdateSelf(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if user.Password == "" {
		user.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Input is illegal " + err.Error(),
		})
		return
	}

	cleanUser := model.User{
		Id:          c.GetInt(ctxkey.Id),
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if user.Password == "$I_LOVE_U" {
		user.Password = "" // rollback to what it should be
		cleanUser.Password = ""
	}
	updatePassword := user.Password != ""
	var existed_user model.User
	_ = model.DB.Where("id = ?", cleanUser.Id).First(&existed_user).Error
	if hashedNewPassword, err := common.Password2Hash(cleanUser.Password); err == nil && hashedNewPassword == existed_user.Password {
		updatePassword = false // 如果新密码与旧密码相同，则不更新密码
	}
	if err := cleanUser.Update(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	originUser, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to delete users with the same permission level or higher permission level",
		})
		return
	}
	err = model.DeleteUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
		})
		return
	}
}

func DeleteSelf(c *gin.Context) {
	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)

	if user.Role == model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Cannot delete super administrator account",
		})
		return
	}

	err := model.DeleteUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	myRole := c.GetInt("role")
	if user.Role >= myRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Unable to create users with permissions greater than or equal to your own",
		})
		return
	}
	// Even for admin users, we cannot fully trust them!
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if err := cleanUser.Insert(ctx, 0); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type ManageRequest struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

// ManageUser Only admin user can do this
func ManageUser(c *gin.Context) {
	var req ManageRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	user := model.User{
		Username: req.Username,
	}
	// Fill attributes
	model.DB.Where(&user).First(&user)
	if user.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "User does not exist",
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to update user information with the same permission level or higher permission level",
		})
		return
	}
	switch req.Action {
	case "disable":
		user.Status = model.UserStatusDisabled
		if user.Role == model.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unable to disable super administrator user",
			})
			return
		}
	case "enable":
		user.Status = model.UserStatusEnabled
	case "delete":
		if user.Role == model.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unable to delete super administrator user",
			})
			return
		}
		if err := user.Delete(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "promote":
		if myRole != model.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Ordinary administrator users cannot promote other users to administrators",
			})
			return
		}
		if user.Role >= model.RoleAdminUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "The user is already an administrator",
			})
			return
		}
		user.Role = model.RoleAdminUser
	case "demote":
		if user.Role == model.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unable to downgrade super administrator user",
			})
			return
		}
		if user.Role == model.RoleCommonUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "The user is already an ordinary user",
			})
			return
		}
		user.Role = model.RoleCommonUser
	}

	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	clearUser := model.User{
		Role:   user.Role,
		Status: user.Status,
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    clearUser,
	})
	return
}

func EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")
	if !common.VerifyCodeWithKey(email, code, common.EmailVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Verification code error or expired",
		})
		return
	}
	id := c.GetInt("id")
	user := model.User{
		Id: id,
	}
	err := user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.Email = email
	// no need to check if this email already taken, because we have used verification code to check it
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.Role == model.RoleRootUser {
		config.RootUserEmail = email
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type topUpRequest struct {
	Key string `json:"key"`
}

func TopUp(c *gin.Context) {
	ctx := c.Request.Context()
	req := topUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	id := c.GetInt("id")
	quota, err := model.Redeem(ctx, req.Key, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    quota,
	})
	return
}

type adminTopUpRequest struct {
	UserId int    `json:"user_id"`
	Quota  int    `json:"quota"`
	Remark string `json:"remark"`
}

func AdminTopUp(c *gin.Context) {
	ctx := c.Request.Context()
	req := adminTopUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = model.IncreaseUserQuota(req.UserId, int64(req.Quota))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if req.Remark == "" {
		req.Remark = fmt.Sprintf("Recharged via API %s", common.LogQuota(int64(req.Quota)))
	}
	model.RecordTopupLog(ctx, req.UserId, req.Remark, req.Quota)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}
