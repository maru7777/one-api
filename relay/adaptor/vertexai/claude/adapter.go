package vertexai

import (
	"net/http"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on VertexAI Claude pricing: https://cloud.google.com/vertex-ai/generative-ai/pricing
var ModelRatios = map[string]adaptor.ModelPrice{
	// Claude Models on VertexAI
	"claude-3-haiku@20240307":       {Ratio: 0.25 * ratio.MilliTokensUsd, CompletionRatio: 5.0}, // $0.25/$1.25 per 1M tokens
	"claude-3-opus@20240229":        {Ratio: 15.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0}, // $15/$75 per 1M tokens
	"claude-opus-4@20250514":        {Ratio: 15.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0}, // $15/$75 per 1M tokens
	"claude-3-sonnet@20240229":      {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $3/$15 per 1M tokens
	"claude-3-5-sonnet@20240620":    {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $3/$15 per 1M tokens
	"claude-3-5-sonnet-v2@20241022": {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $3/$15 per 1M tokens
	"claude-3-5-haiku@20241022":     {Ratio: 1.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $1/$5 per 1M tokens
	"claude-3-7-sonnet@20250219":    {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $3/$15 per 1M tokens
	"claude-sonnet-4@20250514":      {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $3/$15 per 1M tokens
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

const anthropicVersion = "vertex-2023-10-16"

type Adaptor struct {
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	claudeReq, err := anthropic.ConvertRequest(c, *request)
	if err != nil {
		return nil, errors.Wrap(err, "convert request")
	}

	req := Request{
		AnthropicVersion: anthropicVersion,
		// Model:            claudeReq.Model,
		Messages:    claudeReq.Messages,
		System:      claudeReq.System,
		MaxTokens:   claudeReq.MaxTokens,
		Temperature: claudeReq.Temperature,
		TopP:        claudeReq.TopP,
		TopK:        claudeReq.TopK,
		Stream:      claudeReq.Stream,
		Tools:       claudeReq.Tools,
	}

	c.Set(ctxkey.RequestModel, request.Model)
	c.Set(ctxkey.ConvertedRequest, req)
	return req, nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, errors.New("not support image request")
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = anthropic.StreamHandler(c, resp)
	} else {
		err, usage = anthropic.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
	return
}
