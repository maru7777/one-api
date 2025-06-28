package cohere

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct {
}

// ConvertImageRequest implements adaptor.Adaptor.
func (*Adaptor) ConvertImageRequest(_ *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, errors.New("not implemented")
}

// ConvertImageRequest implements adaptor.Adaptor.

func (a *Adaptor) Init(meta *meta.Meta) {

}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return fmt.Sprintf("%s/v1/chat", meta.BaseURL), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return ConvertRequest(*request), nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, resp)
	} else {
		err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "Cohere"
}

// Pricing methods - Cohere adapter manages its own model pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.000001

	// Direct map definition - much easier to maintain and edit
	// Pricing from https://cohere.com/pricing
	return map[string]adaptor.ModelPrice{
		// Command Models
		"command":         {Ratio: 15 * MilliTokensUsd, CompletionRatio: 2}, // $15/$30 per 1M tokens
		"command-nightly": {Ratio: 15 * MilliTokensUsd, CompletionRatio: 2}, // $15/$30 per 1M tokens

		// Command Light Models
		"command-light":         {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 2}, // $0.3/$0.6 per 1M tokens
		"command-light-nightly": {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 2}, // $0.3/$0.6 per 1M tokens

		// Command R Models
		"command-r":      {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 3}, // $0.5/$1.5 per 1M tokens
		"command-r-plus": {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},   // $3/$15 per 1M tokens

		// Internet-enabled variants (same pricing as base models)
		"command-internet":               {Ratio: 15 * MilliTokensUsd, CompletionRatio: 2},
		"command-nightly-internet":       {Ratio: 15 * MilliTokensUsd, CompletionRatio: 2},
		"command-light-internet":         {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 2},
		"command-light-nightly-internet": {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 2},
		"command-r-internet":             {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 3},
		"command-r-plus-internet":        {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default Cohere pricing
	return 0.5 * 0.000001 // Default USD pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for Cohere
	return 3.0
}
