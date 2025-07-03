package xai

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct {
	adaptor.DefaultPricingMethods
}

func (a *Adaptor) Init(meta *meta.Meta) {}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return "", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	return nil, nil
}

func (a *Adaptor) GetModelList() []string {
	return adaptor.GetModelListFromPricing(ModelRatios)
}

func (a *Adaptor) GetChannelName() string {
	return "xai"
}

// GetDefaultModelPricing returns the pricing information for XAI models
// Based on XAI pricing: https://console.x.ai/
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.5 // 0.000001 * 500000 = 0.5 quota per milli-token

	return map[string]adaptor.ModelPrice{
		// XAI Models - Based on https://console.x.ai/
		"grok-2":               {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"grok-vision-beta":     {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"grok-2-vision-1212":   {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"grok-2-vision":        {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"grok-2-vision-latest": {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"grok-2-1212":          {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"grok-2-latest":        {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"grok-beta":            {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Use default fallback from DefaultPricingMethods
	return a.DefaultPricingMethods.GetModelRatio(modelName)
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Use default fallback from DefaultPricingMethods
	return a.DefaultPricingMethods.GetCompletionRatio(modelName)
}
