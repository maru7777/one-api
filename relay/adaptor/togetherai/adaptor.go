package togetherai

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
	return "togetherai"
}

// GetDefaultModelPricing returns the pricing information for TogetherAI models
// Based on TogetherAI pricing: https://www.together.ai/pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.5 // 0.000001 * 500000 = 0.5 quota per milli-token

	return map[string]adaptor.ModelPrice{
		// TogetherAI Models - Based on https://www.together.ai/pricing
		"meta-llama/Llama-2-7b-chat-hf":                {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Llama-2-13b-chat-hf":               {Ratio: 0.225 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Llama-2-70b-chat-hf":               {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3-8B-Instruct":          {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3-70B-Instruct":         {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3.1-8B-Instruct":        {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3.1-70B-Instruct":       {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/Meta-Llama-3.1-405B-Instruct":      {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/Mistral-7B-Instruct-v0.1":           {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/Mixtral-8x7B-Instruct-v0.1":         {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/Mixtral-8x22B-Instruct-v0.1":        {Ratio: 1.2 * MilliTokensUsd, CompletionRatio: 1},
		"NousResearch/Nous-Hermes-2-Mixtral-8x7B-DPO":  {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen1.5-7B-Chat":                         {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen1.5-14B-Chat":                        {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen1.5-32B-Chat":                        {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen1.5-72B-Chat":                        {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen1.5-110B-Chat":                       {Ratio: 1.8 * MilliTokensUsd, CompletionRatio: 1},
		"Qwen/Qwen2-72B-Instruct":                      {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek-ai/deepseek-coder-33b-instruct":      {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek-ai/deepseek-llm-67b-chat":            {Ratio: 0.9 * MilliTokensUsd, CompletionRatio: 1},
		"togethercomputer/RedPajama-INCITE-Chat-3B-v1": {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1},
		"togethercomputer/RedPajama-INCITE-7B-Chat":    {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"zero-one-ai/Yi-34B-Chat":                      {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 1},
		"01-ai/Yi-6B-Chat":                             {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"01-ai/Yi-34B-Chat":                            {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 1},
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
