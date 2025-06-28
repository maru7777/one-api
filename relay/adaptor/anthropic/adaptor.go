package anthropic

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *meta.Meta) {

}

// https://docs.anthropic.com/claude/reference/messages_post
// anthopic migrate to Message API
func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return fmt.Sprintf("%s/v1/messages", meta.BaseURL), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("x-api-key", meta.APIKey)
	anthropicVersion := c.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = "2023-06-01"
	}
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("anthropic-beta", "messages-2023-12-15")

	// https://x.com/alexalbert__/status/1812921642143900036
	// claude-3-5-sonnet can support 8k context
	if strings.HasPrefix(meta.ActualModelName, "claude-3-7-sonnet") {
		req.Header.Set("anthropic-beta", "output-128k-2025-02-19")
	}

	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	c.Set("claude_model", request.Model)
	return ConvertRequest(c, *request)
}

func (a *Adaptor) ConvertImageRequest(_ *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
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
	return "anthropic"
}

// Pricing methods - Anthropic adapter manages its own model pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.000001

	// Direct map definition - much easier to maintain and edit
	return map[string]adaptor.ModelPrice{
		// Claude Instant Models
		"claude-instant-1.2": {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 3.0},

		// Claude 2 Models
		"claude-2.0": {Ratio: 8 * MilliTokensUsd, CompletionRatio: 3.0},
		"claude-2.1": {Ratio: 8 * MilliTokensUsd, CompletionRatio: 3.0},

		// Claude 3 Haiku Models
		"claude-3-haiku-20240307":   {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 5.0},
		"claude-3-5-haiku-latest":   {Ratio: 1 * MilliTokensUsd, CompletionRatio: 5.0},
		"claude-3-5-haiku-20241022": {Ratio: 1 * MilliTokensUsd, CompletionRatio: 5.0},

		// Claude 3 Sonnet Models
		"claude-3-sonnet-20240229":   {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5.0},
		"claude-3-5-sonnet-latest":   {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5.0},
		"claude-3-5-sonnet-20240620": {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5.0},
		"claude-3-5-sonnet-20241022": {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5.0},
		"claude-3-7-sonnet-latest":   {Ratio: 15 * MilliTokensUsd, CompletionRatio: 5.0},
		"claude-3-7-sonnet-20250219": {Ratio: 15 * MilliTokensUsd, CompletionRatio: 5.0},

		// Claude 3 Opus Models
		"claude-3-opus-20240229": {Ratio: 15 * MilliTokensUsd, CompletionRatio: 5.0},
		"claude-opus-4-20250514": {Ratio: 60 * MilliTokensUsd, CompletionRatio: 5.0},

		// Claude 4 Sonnet Models
		"claude-sonnet-4-20250514": {Ratio: 15 * MilliTokensUsd, CompletionRatio: 5.0},
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Fallback to global pricing for unknown models
	return 3 * 0.000001 // Default Anthropic pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for Anthropic
	return 5.0
}
