package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/network"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

func GetRequestCost(c *gin.Context) {
	reqId := c.Param("request_id")
	if reqId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "request_id 不能为空",
		})
		return
	}

	docu, err := model.GetCostByRequestId(reqId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, docu)
}

func GetAllTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.Query("order")
	tokens, err := model.GetAllUserTokens(userId, p*config.ItemsPerPage, config.ItemsPerPage, order)

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
		"data":    tokens,
	})
	return
}

func SearchTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	keyword := c.Query("keyword")
	tokens, err := model.SearchUserTokens(userId, keyword)
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
		"data":    tokens,
	})
	return
}

func GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	token, err := model.GetTokenByIds(id, userId)
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
		"data":    token,
	})
	return
}

func GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt(ctxkey.TokenId)
	userId := c.GetInt(ctxkey.Id)
	token, err := model.GetTokenByIds(tokenId, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

func validateToken(c *gin.Context, token *model.Token) error {
	if len(token.Name) > 30 {
		return fmt.Errorf("Token name is too long")
	}

	if token.Subnet != nil && *token.Subnet != "" {
		err := network.IsValidSubnets(*token.Subnet)
		if err != nil {
			return fmt.Errorf("None效的网段：%s", err.Error())
		}
	}

	return nil
}

func AddToken(c *gin.Context) {
	token := new(model.Token)
	err := c.ShouldBindJSON(token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数Error:%s", err.Error()),
		})
		return
	}

	cleanToken := model.Token{
		UserId:         c.GetInt(ctxkey.Id),
		Name:           token.Name,
		Key:            random.GenerateKey(),
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    token.ExpiredTime,
		RemainQuota:    token.RemainQuota,
		UnlimitedQuota: token.UnlimitedQuota,
		Models:         token.Models,
		Subnet:         token.Subnet,
	}
	err = cleanToken.Insert()
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
		"data":    cleanToken,
	})
	return
}

func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	err := model.DeleteTokenById(id, userId)
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

type consumeTokenRequest struct {
	// AddUsedQuota add or subtract used quota from another source
	AddUsedQuota uint `json:"add_used_quota" gorm:"-"`
	// AddReason is the reason for adding or subtracting used quota
	AddReason string `json:"add_reason" gorm:"-"`
}

// ConsumeToken consume token from another source,
// let one-api to gather billing from different sources.
func ConsumeToken(c *gin.Context) {
	tokenPatch := new(consumeTokenRequest)
	err := c.ShouldBindJSON(tokenPatch)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	userID := c.GetInt(ctxkey.Id)
	tokenID := c.GetInt(ctxkey.TokenId)
	cleanToken, err := model.GetTokenByIds(tokenID, userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if cleanToken.Status != model.TokenStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "API KeysNot enabled",
		})
		return
	}

	if cleanToken.Status == model.TokenStatusExpired &&
		cleanToken.ExpiredTime <= helper.GetTimestamp() && cleanToken.ExpiredTime != -1 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "The token has expired and cannot be enabled. Please modify the expiration time of the token, or set it to never expire.",
		})
		return
	}
	if cleanToken.Status == model.TokenStatusExhausted &&
		cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "The available quota of the token has been used up and cannot be enabled. Please modify the remaining quota of the token, or set it to unlimited quota",
		})
		return
	}

	// let admin to add or subtract used quota,
	// make it possible to aggregate billings from different sources.
	if cleanToken.RemainQuota < int64(tokenPatch.AddUsedQuota) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Remaining quota不足",
		})
		return
	}

	if err = model.DecreaseTokenQuota(cleanToken.Id, int64(tokenPatch.AddUsedQuota)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	model.RecordConsumeLog(c.Request.Context(),
		userID, 0, 0, 0, tokenPatch.AddReason, cleanToken.Name,
		int64(tokenPatch.AddUsedQuota),
		fmt.Sprintf("外部(%s)消耗 %s",
			tokenPatch.AddReason, common.LogQuota(int64(tokenPatch.AddUsedQuota))))

	err = cleanToken.Update()
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
		"data":    cleanToken,
	})

	return
}

func UpdateToken(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	statusOnly := c.Query("status_only")
	tokenPatch := new(model.Token)
	err := c.ShouldBindJSON(tokenPatch)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	token := new(model.Token)
	if err = copier.Copy(token, tokenPatch); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数Error:%s", err.Error()),
		})
		return
	}

	cleanToken, err := model.GetTokenByIds(token.Id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	switch token.Status {
	case model.TokenStatusEnabled:
		if cleanToken.Status == model.TokenStatusExpired &&
			cleanToken.ExpiredTime <= helper.GetTimestamp() && cleanToken.ExpiredTime != -1 &&
			token.ExpiredTime != -1 && token.ExpiredTime < helper.GetTimestamp() {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "The token has expired and cannot be enabled. Please modify the expiration time of the token, or set it to never expire.",
			})
			return
		}
		if cleanToken.Status == model.TokenStatusExhausted &&
			cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota &&
			token.RemainQuota <= 0 && !token.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "The available quota of the token has been used up and cannot be enabled. Please modify the remaining quota of the token, or set it to unlimited quota",
			})
			return
		}
	case model.TokenStatusExhausted:
		if token.RemainQuota > 0 || token.UnlimitedQuota {
			token.Status = model.TokenStatusEnabled
		}
	case model.TokenStatusExpired:
		if token.ExpiredTime == -1 || token.ExpiredTime > helper.GetTimestamp() {
			token.Status = model.TokenStatusEnabled
		}
	}

	if statusOnly != "" {
		cleanToken.Status = token.Status
	} else {
		// If you add more fields, please also update token.Update()
		cleanToken.Name = token.Name
		cleanToken.ExpiredTime = token.ExpiredTime
		cleanToken.UnlimitedQuota = token.UnlimitedQuota
		cleanToken.Models = token.Models
		cleanToken.Subnet = token.Subnet
		cleanToken.RemainQuota = token.RemainQuota
		cleanToken.Status = token.Status
	}

	err = cleanToken.Update()
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
		"data":    cleanToken,
	})
	return
}
