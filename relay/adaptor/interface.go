package adaptor

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// ModelPrice represents pricing information for a model
type ModelPrice struct {
	Ratio           float64 `json:"ratio"`
	CompletionRatio float64 `json:"completion_ratio,omitempty"`
}

type Adaptor interface {
	Init(meta *meta.Meta)
	GetRequestURL(meta *meta.Meta) (string, error)
	SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error
	ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error)
	ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error)
	DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error)
	DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode)
	GetModelList() []string
	GetChannelName() string

	// Pricing methods - each adapter manages its own model pricing
	GetDefaultModelPricing() map[string]ModelPrice
	GetModelRatio(modelName string) float64
	GetCompletionRatio(modelName string) float64
}

// DefaultPricingMethods provides default implementations for adapters without specific pricing
type DefaultPricingMethods struct{}

func (d *DefaultPricingMethods) GetDefaultModelPricing() map[string]ModelPrice {
	return make(map[string]ModelPrice) // Empty pricing map
}

func (d *DefaultPricingMethods) GetModelRatio(modelName string) float64 {
	// Fallback to a reasonable default
	return 2.5 * 0.000001 // 2.5 USD per million tokens
}

func (d *DefaultPricingMethods) GetCompletionRatio(modelName string) float64 {
	return 1.0 // Default completion ratio
}

// GetModelListFromPricing derives model list from pricing map keys
// This eliminates the need for separate ModelList variables
func GetModelListFromPricing(pricing map[string]ModelPrice) []string {
	models := make([]string, 0, len(pricing))
	for model := range pricing {
		models = append(models, model)
	}
	return models
}
