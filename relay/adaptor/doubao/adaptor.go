package doubao

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
	return "doubao"
}

// GetDefaultModelPricing returns the pricing information for Doubao models
// Based on Doubao pricing: https://www.volcengine.com/docs/82379/1099320
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensRmb = 3.5 // 0.000007 * 500000 = 3.5 quota per milli-token

	return map[string]adaptor.ModelPrice{
		// Doubao Models - Based on https://www.volcengine.com/docs/82379/1099320
		"doubao-lite-4k":        {Ratio: 0.0003 * MilliTokensRmb, CompletionRatio: 1},
		"doubao-lite-32k":       {Ratio: 0.0006 * MilliTokensRmb, CompletionRatio: 1},
		"doubao-lite-128k":      {Ratio: 0.0008 * MilliTokensRmb, CompletionRatio: 1},
		"doubao-pro-4k":         {Ratio: 0.0008 * MilliTokensRmb, CompletionRatio: 1},
		"doubao-pro-32k":        {Ratio: 0.002 * MilliTokensRmb, CompletionRatio: 1},
		"doubao-pro-128k":       {Ratio: 0.005 * MilliTokensRmb, CompletionRatio: 1},
		"doubao-pro-256k":       {Ratio: 0.008 * MilliTokensRmb, CompletionRatio: 1},
		"doubao-character-8k":   {Ratio: 0.008 * MilliTokensRmb, CompletionRatio: 1},
		"doubao-vision-pro-32k": {Ratio: 0.002 * MilliTokensRmb, CompletionRatio: 1},
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
