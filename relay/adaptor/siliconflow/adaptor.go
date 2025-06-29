package siliconflow

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
	return "siliconflow"
}

// GetDefaultModelPricing returns the pricing information for SiliconFlow models
// Based on SiliconFlow pricing: https://siliconflow.cn/pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.5 // 0.000001 * 500000 = 0.5 quota per milli-token

	return map[string]adaptor.ModelPrice{
		// SiliconFlow Models - Based on https://siliconflow.cn/pricing
		"deepseek-chat":                           {Ratio: 0.14 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek-coder":                          {Ratio: 0.14 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen2-72B-Instruct":                 {Ratio: 0.56 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen2-7B-Instruct":                  {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen2-1.5B-Instruct":                {Ratio: 0.14 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen2-0.5B-Instruct":                {Ratio: 0.14 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3-8B-Instruct":     {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3-70B-Instruct":    {Ratio: 0.56 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3.1-8B-Instruct":   {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3.1-70B-Instruct":  {Ratio: 0.56 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3.1-405B-Instruct": {Ratio: 2.8 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/Mistral-7B-Instruct-v0.2":      {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/Mixtral-8x7B-Instruct-v0.1":    {Ratio: 0.56 * MilliTokensUsd, CompletionRatio: 1},
		"01-ai/Yi-1.5-9B-Chat-16K":                {Ratio: 0.14 * MilliTokensUsd, CompletionRatio: 1},
		"01-ai/Yi-1.5-6B-Chat":                    {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"THUDM/glm-4-9b-chat":                     {Ratio: 0.14 * MilliTokensUsd, CompletionRatio: 1},
		"THUDM/chatglm3-6b":                       {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"internlm/internlm2_5-7b-chat":            {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemma-2-9b-it":                    {Ratio: 0.14 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemma-2-27b-it":                   {Ratio: 0.28 * MilliTokensUsd, CompletionRatio: 1},
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
