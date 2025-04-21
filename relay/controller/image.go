package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func getImageRequest(c *gin.Context, _ int) (*relaymodel.ImageRequest, error) {
	imageRequest := &relaymodel.ImageRequest{}
	err := common.UnmarshalBodyReusable(c, imageRequest)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// if imageRequest.N == 0 {
	// 	imageRequest.N = 1
	// }
	// if imageRequest.Size == "" {
	// 	imageRequest.Size = "1024x1024"
	// }
	// if imageRequest.Model == "" {
	// 	imageRequest.Model = "dall-e-2"
	// }
	return imageRequest, nil
}

// func isValidImageSize(model string, size string) bool {
// 	if model == "cogview-3" || billingratio.ImageSizeRatios[model] == nil {
// 		return true
// 	}
// 	_, ok := billingratio.ImageSizeRatios[model][size]
// 	return ok
// }

// func isValidImagePromptLength(model string, promptLength int) bool {
// 	maxPromptLength, ok := billingratio.ImagePromptLengthLimitations[model]
// 	return !ok || promptLength <= maxPromptLength
// }

// func isWithinRange(element string, value int) bool {
// 	amounts, ok := billingratio.ImageGenerationAmounts[element]
// 	return !ok || (value >= amounts[0] && value <= amounts[1])
// }

// func getImageSizeRatio(model string, size string) float64 {
// 	if ratio, ok := billingratio.ImageSizeRatios[model][size]; ok {
// 		return ratio
// 	}
// 	return 1
// }

// func validateImageRequest(imageRequest *relaymodel.ImageRequest, _ *metalib.Meta) *relaymodel.ErrorWithStatusCode {
// 	// check prompt length
// 	if imageRequest.Prompt == "" {
// 		return openai.ErrorWrapper(errors.New("prompt is required"), "prompt_missing", http.StatusBadRequest)
// 	}

// 	// model validation
// 	if !isValidImageSize(imageRequest.Model, imageRequest.Size) {
// 		return openai.ErrorWrapper(errors.New("size not supported for this image model"), "size_not_supported", http.StatusBadRequest)
// 	}

// 	if !isValidImagePromptLength(imageRequest.Model, len(imageRequest.Prompt)) {
// 		return openai.ErrorWrapper(errors.New("prompt is too long"), "prompt_too_long", http.StatusBadRequest)
// 	}

// 	// Number of generated images validation
// 	if !isWithinRange(imageRequest.Model, imageRequest.N) {
// 		return openai.ErrorWrapper(errors.New("invalid value of n"), "n_not_within_range", http.StatusBadRequest)
// 	}
// 	return nil
// }

//	func getImageCostRatio(imageRequest *relaymodel.ImageRequest) (float64, error) {
//		if imageRequest == nil {
//			return 0, errors.New("imageRequest is nil")
//		}
//		imageCostRatio := getImageSizeRatio(imageRequest.Model, imageRequest.Size)
//		if imageRequest.Quality == "hd" && imageRequest.Model == "dall-e-3" {
//			if imageRequest.Size == "1024x1024" {
//				imageCostRatio *= 2
//			} else {
//				imageCostRatio *= 1.5
//			}
//		}
//		return imageCostRatio, nil
//	}
func getFromContext[T any](c *gin.Context, key string, defaultValue T) T {
	val, exists := c.Get(key)
	if !exists {
		return defaultValue
	}
	if typedVal, ok := val.(T); ok {
		return typedVal
	}
	return defaultValue
}
func RelayImageHelper(c *gin.Context, relayMode int) *relaymodel.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := metalib.GetByContext(c)
	imageRequest, err := getImageRequest(c, meta.Mode)
	if err != nil {
		logger.Errorf(ctx, "getImageRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_image_request", http.StatusBadRequest)
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)
	finalRequest, err := adaptor.ConvertImageRequest(c, imageRequest)
	if err != nil {
		return openai.ErrorWrapper(err, "convert_image_request_failed", http.StatusInternalServerError)
	}
	// // map model name
	// var isModelMapped bool
	// meta.OriginModelName = finalRequest.Model
	// imageRequest.Model = meta.ActualModelName
	// isModelMapped = meta.OriginModelName != meta.ActualModelName
	// meta.ActualModelName = imageRequest.Model
	// metalib.Set2Context(c, meta)

	// // model validation
	// bizErr := validateImageRequest(imageRequest, meta)
	// if bizErr != nil {
	// 	return bizErr
	// }

	// c.Set("response_format", imageRequest.ResponseFormat)

	var requestBody io.Reader
	// if strings.ToLower(c.GetString(ctxkey.ContentType)) == "application/json" &&
	// 	isModelMapped || meta.ChannelType == channeltype.Azure { // make Azure channel request body
	// 	jsonStr, err := json.Marshal(imageRequest)
	// 	if err != nil {
	// 		return openai.ErrorWrapper(err, "marshal_image_request_failed", http.StatusInternalServerError)
	// 	}
	// 	requestBody = bytes.NewBuffer(jsonStr)
	// } else {
	// 	requestBody = c.Request.Body
	// }
	jsonStr, err := json.Marshal(finalRequest)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_image_request_failed", http.StatusInternalServerError)
	}
	requestBody = bytes.NewBuffer(jsonStr)

	// // these adaptors need to convert the request
	// switch meta.ChannelType {
	// case channeltype.Zhipu,
	// 	channeltype.Ali,
	// 	channeltype.VertextAI,
	// 	channeltype.Baidu:
	// 	finalRequest, err := adaptor.ConvertImageRequest(c, imageRequest)
	// 	if err != nil {
	// 		return openai.ErrorWrapper(err, "convert_image_request_failed", http.StatusInternalServerError)
	// 	}
	// 	jsonStr, err := json.Marshal(finalRequest)
	// 	if err != nil {
	// 		return openai.ErrorWrapper(err, "marshal_image_request_failed", http.StatusInternalServerError)
	// 	}
	// 	requestBody = bytes.NewBuffer(jsonStr)
	// case channeltype.Replicate:
	// 	finalRequest, err := replicate.ConvertImageRequest(c, imageRequest)
	// 	if err != nil {
	// 		return openai.ErrorWrapper(err, "convert_image_request_failed", http.StatusInternalServerError)
	// 	}
	// 	jsonStr, err := json.Marshal(finalRequest)
	// 	if err != nil {
	// 		return openai.ErrorWrapper(err, "marshal_image_request_failed", http.StatusInternalServerError)
	// 	}
	// 	requestBody = bytes.NewBuffer(jsonStr)
	// }

	// 适配所有图像请求的计费方法
	n := meta.VendorContext["PicNumber"].(int)
	originModelName := meta.VendorContext["Model"].(string)
	size := meta.VendorContext["PicSize"].(string)
	quality := meta.VendorContext["Quality"].(string)
	var imageCostRatio float64 = 1.0
	if billingratio.ImageSizeRatios[originModelName] != nil {
		if ratio, ok := billingratio.ImageSizeRatios[originModelName][size]; ok {
			imageCostRatio = ratio
		}
	}
	if quality == "hd" && originModelName == "dall-e-3" {
		if size == "1024x1024" {
			imageCostRatio *= 2
		} else {
			imageCostRatio *= 1.5
		}
	}
	modelRatio := billingratio.GetModelRatio(meta.ActualModelName, meta.ChannelType)
	groupRatio := c.GetFloat64(ctxkey.ChannelRatio)
	var quota = int64(modelRatio*groupRatio*imageCostRatio) * int64(n) * 1000
	userQuota, _ := model.CacheGetUserQuota(ctx, meta.UserId)
	if userQuota < quota {
		return openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	defer func(ctx context.Context) {
		if resp != nil &&
			resp.StatusCode != http.StatusCreated && // replicate returns 201
			resp.StatusCode != http.StatusOK {
			return
		}

		err := model.PostConsumeTokenQuota(meta.TokenId, quota)
		if err != nil {
			logger.SysError("error consuming token remain quota: " + err.Error())
		}
		err = model.CacheUpdateUserQuota(ctx, meta.UserId)
		if err != nil {
			logger.SysError("error update user quota cache: " + err.Error())
		}
		if quota >= 0 {
			tokenName := c.GetString(ctxkey.TokenName)
			logContent := fmt.Sprintf("model rate %.2f, group rate %.2f, num %d",
				modelRatio, groupRatio, n)
			model.RecordConsumeLog(ctx, &model.Log{
				UserId:           meta.UserId,
				ChannelId:        meta.ChannelId,
				PromptTokens:     0,
				CompletionTokens: 0,
				ModelName:        originModelName,
				TokenName:        tokenName,
				Quota:            int(quota),
				Content:          logContent,
			})
			model.UpdateUserUsedQuotaAndRequestCount(meta.UserId, quota)
			channelId := c.GetInt(ctxkey.ChannelId)
			model.UpdateChannelUsedQuota(channelId, quota)

			// also update user request cost
			docu := model.NewUserRequestCost(
				c.GetInt(ctxkey.Id),
				c.GetString(ctxkey.RequestId),
				quota,
			)
			if err = docu.Insert(); err != nil {
				logger.Errorf(c, "insert user request cost failed: %+v", err)
			}
		}
	}(c.Request.Context())

	// do response
	_, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		return respErr
	}

	return nil
}
