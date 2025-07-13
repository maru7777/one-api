package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/pricing"
)

func GetAllChannels(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	channels, err := model.GetAllChannels(p*config.MaxItemsPerPage, config.MaxItemsPerPage, "limited")
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
		"data":    channels,
	})
	return
}

func SearchChannels(c *gin.Context) {
	keyword := c.Query("keyword")
	channels, err := model.SearchChannels(keyword)
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
		"data":    channels,
	})
	return
}

func GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, false)
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
		"data":    channel,
	})
	return
}

func AddChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validate inference profile ARN map if provided
	if channel.InferenceProfileArnMap != nil && *channel.InferenceProfileArnMap != "" {
		err = model.ValidateInferenceProfileArnMapJSON(*channel.InferenceProfileArnMap)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Invalid inference profile ARN map: " + err.Error(),
			})
			return
		}
	}

	channel.CreatedTime = helper.GetTimestamp()
	keys := strings.Split(channel.Key, "\n")
	channels := make([]model.Channel, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		localChannel := channel
		localChannel.Key = key
		channels = append(channels, localChannel)
	}
	err = model.BatchInsertChannels(channels)
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

func DeleteChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	channel := model.Channel{Id: id}
	err := channel.Delete()
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

func DeleteDisabledChannel(c *gin.Context) {
	rows, err := model.DeleteDisabledChannel()
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
		"data":    rows,
	})
	return
}

func UpdateChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validate inference profile ARN map if provided
	if channel.InferenceProfileArnMap != nil && *channel.InferenceProfileArnMap != "" {
		err = model.ValidateInferenceProfileArnMapJSON(*channel.InferenceProfileArnMap)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Invalid inference profile ARN map: " + err.Error(),
			})
			return
		}
	}

	err = channel.Update()
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
		"data":    channel,
	})
	return
}

func GetChannelPricing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	modelRatio := channel.GetModelRatio()
	completionRatio := channel.GetCompletionRatio()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"model_ratio":      modelRatio,
			"completion_ratio": completionRatio,
		},
	})
	return
}

func UpdateChannelPricing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var request struct {
		ModelRatio      map[string]float64 `json:"model_ratio"`
		CompletionRatio map[string]float64 `json:"completion_ratio"`
	}

	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	channel, err := model.GetChannelById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	err = channel.SetModelRatio(request.ModelRatio)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to set model ratio: " + err.Error(),
		})
		return
	}

	err = channel.SetCompletionRatio(request.CompletionRatio)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to set completion ratio: " + err.Error(),
		})
		return
	}

	err = channel.Update()
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

func GetChannelDefaultPricing(c *gin.Context) {
	channelType, err := strconv.Atoi(c.Query("type"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid channel type: " + err.Error(),
		})
		return
	}

	var defaultPricing map[string]adaptor.ModelPrice

	// For Custom channels and OpenAI-compatible channels, use global pricing from all adapters
	// This gives users access to pricing for all supported models
	if channelType == channeltype.Custom || channelType == channeltype.OpenAICompatible {
		// Use global pricing manager to get pricing from all adapters
		defaultPricing = pricing.GetGlobalModelPricing()
	} else {
		// For specific channel types, use their adapter's default pricing
		// Convert channel type to API type first
		apiType := channeltype.ToAPIType(channelType)
		adaptor := relay.GetAdaptor(apiType)
		if adaptor == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unsupported channel type",
			})
			return
		}
		defaultPricing = adaptor.GetDefaultModelPricing()
	}

	// Separate model ratios and completion ratios for UI compatibility
	modelRatios := make(map[string]float64)
	completionRatios := make(map[string]float64)

	for model, price := range defaultPricing {
		modelRatios[model] = price.Ratio
		// Include all completion ratios, including 0 (which is valid pricing info)
		completionRatios[model] = price.CompletionRatio
	}

	// Convert to JSON
	modelRatioJSON, err := json.Marshal(modelRatios)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to serialize model ratios: " + err.Error(),
		})
		return
	}

	completionRatioJSON, err := json.Marshal(completionRatios)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to serialize completion ratios: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"model_ratio":      string(modelRatioJSON),
			"completion_ratio": string(completionRatioJSON),
		},
	})
}
