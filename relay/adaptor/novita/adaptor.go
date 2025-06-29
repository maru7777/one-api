package novita

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
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "novita"
}

// GetDefaultModelPricing returns the pricing information for Novita models
// Based on Novita pricing: https://novita.ai/pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.5 // 0.000001 * 500000 = 0.5 quota per milli-token

	return map[string]adaptor.ModelPrice{
		// Novita Models - Based on https://novita.ai/pricing
		"meta-llama/llama-3.1-8b-instruct":        {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.1-70b-instruct":       {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.1-405b-instruct":      {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3-8b-instruct":          {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3-70b-instruct":         {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-7b-instruct":           {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mixtral-8x7b-instruct":         {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mixtral-8x22b-instruct":        {Ratio: 1.2 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2-72b-instruct":                {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2-7b-instruct":                 {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek-ai/deepseek-coder-33b-instruct": {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 1},
		"01-ai/yi-34b-chat":                       {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 1},
		"01-ai/yi-6b-chat":                        {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemma-2-9b-it":                    {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemma-2-27b-it":                   {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 1},
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
