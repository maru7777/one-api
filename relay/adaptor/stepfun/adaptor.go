package stepfun

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
	return "stepfun"
}

// GetDefaultModelPricing returns the pricing information for StepFun models
// Based on StepFun pricing: https://platform.stepfun.com/docs/pricing/details
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliRmb = 3.5 // 0.000007 * 500000 = 3.5 quota per milli-token

	return map[string]adaptor.ModelPrice{
		// StepFun Models - Based on https://platform.stepfun.com/docs/pricing/details
		"step-1-8k":      {Ratio: 0.005 * MilliRmb, CompletionRatio: 1},
		"step-1-32k":     {Ratio: 0.015 * MilliRmb, CompletionRatio: 1},
		"step-1-128k":    {Ratio: 0.040 * MilliRmb, CompletionRatio: 1},
		"step-1-256k":    {Ratio: 0.095 * MilliRmb, CompletionRatio: 1},
		"step-1-flash":   {Ratio: 0.001 * MilliRmb, CompletionRatio: 1},
		"step-2-16k":     {Ratio: 0.038 * MilliRmb, CompletionRatio: 1},
		"step-1v-8k":     {Ratio: 0.005 * MilliRmb, CompletionRatio: 1},
		"step-1v-32k":    {Ratio: 0.015 * MilliRmb, CompletionRatio: 1},
		"step-1x-medium": {Ratio: 0.020 * MilliRmb, CompletionRatio: 1}, // Estimated pricing
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
