package lingyiwanwu

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
	return "lingyiwanwu"
}

// GetDefaultModelPricing returns the pricing information for LingYiWanWu models
// Based on LingYiWanWu pricing: https://platform.lingyiwanwu.com/docs#-计费单元
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensRmb = 3.5 // 0.000007 * 500000 = 3.5 quota per milli-token

	return map[string]adaptor.ModelPrice{
		// LingYiWanWu Models - Based on https://platform.lingyiwanwu.com/docs#-计费单元
		"yi-34b-chat-0205": {Ratio: 2.5 * MilliTokensRmb, CompletionRatio: 1},
		"yi-34b-chat-200k": {Ratio: 12.0 * MilliTokensRmb, CompletionRatio: 1},
		"yi-vl-plus":       {Ratio: 6.0 * MilliTokensRmb, CompletionRatio: 1},
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
