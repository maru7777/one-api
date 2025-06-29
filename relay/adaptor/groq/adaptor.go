package groq

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct {
	adaptor.DefaultPricingMethods
}

func (a *Adaptor) GetChannelName() string {
	return "groq"
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

// GetDefaultModelPricing returns the pricing information for Groq models
// Based on Groq pricing: https://groq.com/pricing/
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	return map[string]adaptor.ModelPrice{
		// Regular Models
		"distil-whisper-large-v3-en": {Ratio: 0.111 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"gemma2-9b-it":               {Ratio: 0.20 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"llama-3.3-70b-versatile":    {Ratio: 0.59 * ratio.MilliTokensUsd, CompletionRatio: 0.79 / 0.59},
		"llama-3.1-8b-instant":       {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 0.08 / 0.05},
		"llama-guard-3-8b":           {Ratio: 0.20 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"llama3-70b-8192":            {Ratio: 0.59 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"llama3-8b-8192":             {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"mixtral-8x7b-32768":         {Ratio: 0.24 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"whisper-large-v3":           {Ratio: 0.111 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"whisper-large-v3-turbo":     {Ratio: 0.04 * ratio.MilliTokensUsd, CompletionRatio: 1},

		// Preview Models
		"qwen-qwq-32b":                          {Ratio: 0.29 * ratio.MilliTokensUsd, CompletionRatio: 0.39 / 0.29},
		"mistral-saba-24b":                      {Ratio: 0.79 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"qwen-2.5-coder-32b":                    {Ratio: 0.79 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"qwen-2.5-32b":                          {Ratio: 0.79 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"deepseek-r1-distill-qwen-32b":          {Ratio: 0.29 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"deepseek-r1-distill-llama-70b-specdec": {Ratio: 0.75 * ratio.MilliTokensUsd, CompletionRatio: 0.99 / 0.75},
		"deepseek-r1-distill-llama-70b":         {Ratio: 0.75 * ratio.MilliTokensUsd, CompletionRatio: 0.99 / 0.75},
		"llama-3.2-1b-preview":                  {Ratio: 0.04 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"llama-3.2-3b-preview":                  {Ratio: 0.06 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"llama-3.2-11b-vision-preview":          {Ratio: 0.18 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"llama-3.2-90b-vision-preview":          {Ratio: 0.90 * ratio.MilliTokensUsd, CompletionRatio: 1},
		"llama-3.3-70b-specdec":                 {Ratio: 0.59 * ratio.MilliTokensUsd, CompletionRatio: 0.99 / 0.59},
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

// Implement required adaptor interface methods (Groq uses OpenAI-compatible API)
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
