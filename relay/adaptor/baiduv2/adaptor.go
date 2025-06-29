package baiduv2

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
	return "baiduv2"
}

// GetDefaultModelPricing returns the pricing information for BaiduV2 models
// Based on Baidu pricing: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9do2
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensRmb = 3.5 // 0.000007 * 500000 = 3.5 quota per milli-token

	return map[string]adaptor.ModelPrice{
		// BaiduV2 Models - Based on https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9do2
		"ernie-4.0-8k":         {Ratio: 0.12 * MilliTokensRmb, CompletionRatio: 1},  // ¥0.12 / 1k tokens
		"ernie-4.0-8k-preview": {Ratio: 0.12 * MilliTokensRmb, CompletionRatio: 1},  // ¥0.12 / 1k tokens
		"ernie-4.0-turbo-8k":   {Ratio: 0.02 * MilliTokensRmb, CompletionRatio: 1},  // ¥0.02 / 1k tokens
		"ernie-3.5-8k":         {Ratio: 0.012 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.012 / 1k tokens
		"ernie-3.5-8k-0205":    {Ratio: 0.012 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.012 / 1k tokens
		"ernie-3.5-8k-1222":    {Ratio: 0.012 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.012 / 1k tokens
		"ernie-3.5-4k-0205":    {Ratio: 0.012 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.012 / 1k tokens
		"ernie-speed-8k":       {Ratio: 0.004 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.004 / 1k tokens
		"ernie-speed-128k":     {Ratio: 0.004 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.004 / 1k tokens
		"ernie-lite-8k":        {Ratio: 0.008 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.008 / 1k tokens
		"ernie-lite-8k-0922":   {Ratio: 0.008 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.008 / 1k tokens
		"ernie-lite-8k-0308":   {Ratio: 0.008 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.008 / 1k tokens
		"ernie-tiny-8k":        {Ratio: 0.001 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.001 / 1k tokens
		"ernie-char-8k":        {Ratio: 0.04 * MilliTokensRmb, CompletionRatio: 1},  // ¥0.04 / 1k tokens
		"ernie-text-embedding": {Ratio: 0.002 * MilliTokensRmb, CompletionRatio: 1}, // ¥0.002 / 1k tokens
		"ernie-vilg-v2":        {Ratio: 0.06 * MilliTokensRmb, CompletionRatio: 1},  // ¥0.06 / 1k tokens
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
