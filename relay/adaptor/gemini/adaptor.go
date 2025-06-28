package gemini

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	channelhelper "github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *meta.Meta) {
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	defaultVersion := config.GeminiVersion
	if strings.Contains(meta.ActualModelName, "gemini-2") ||
		strings.Contains(meta.ActualModelName, "gemini-1.5") ||
		strings.Contains(meta.ActualModelName, "gemma-3") {
		defaultVersion = "v1beta"
	}

	version := helper.AssignOrDefault(meta.Config.APIVersion, defaultVersion)
	action := ""
	switch meta.Mode {
	case relaymode.Embeddings:
		action = "batchEmbedContents"
	default:
		action = "generateContent"
	}

	if meta.IsStream {
		action = "streamGenerateContent?alt=sse"
	}

	return fmt.Sprintf("%s/%s/models/%s:%s", meta.BaseURL, version, meta.ActualModelName, action), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	channelhelper.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("x-goog-api-key", meta.APIKey)
	req.URL.Query().Add("key", meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case relaymode.Embeddings:
		geminiEmbeddingRequest := ConvertEmbeddingRequest(*request)
		return geminiEmbeddingRequest, nil
	default:
		geminiRequest := ConvertRequest(*request)
		return geminiRequest, nil
	}
}

func (a *Adaptor) ConvertImageRequest(_ *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return channelhelper.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		var responseText string
		err, responseText = StreamHandler(c, resp)
		usage = openai.ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
	} else {
		switch meta.Mode {
		case relaymode.Embeddings:
			err, usage = EmbeddingHandler(c, resp)
		default:
			err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "google gemini"
}

// Pricing methods - Gemini adapter manages its own model pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]channelhelper.ModelPrice {
	const MilliTokensUsd = 0.000001

	// Direct map definition - much easier to maintain and edit
	// Pricing from https://ai.google.dev/pricing
	return map[string]channelhelper.ModelPrice{
		// Gemini Pro Models
		"gemini-pro":     {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 3},
		"gemini-1.0-pro": {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 3},

		// Gemma Models
		"gemma-2-2b-it":  {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1.4},
		"gemma-2-9b-it":  {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1.4},
		"gemma-2-27b-it": {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1.4},
		"gemma-3-27b-it": {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1.4},

		// Gemini 1.5 Flash Models
		"gemini-1.5-flash":    {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-1.5-flash-8b": {Ratio: 0.0375 * MilliTokensUsd, CompletionRatio: 4},

		// Gemini 1.5 Pro Models
		"gemini-1.5-pro":              {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-1.5-pro-experimental": {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},

		// Embedding Models
		"text-embedding-004": {Ratio: 0.00001 * MilliTokensUsd, CompletionRatio: 1},
		"aqa":                {Ratio: 1, CompletionRatio: 1},

		// Gemini 2.0 Flash Models
		"gemini-2.0-flash":                      {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.0-flash-exp":                  {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.0-flash-lite":                 {Ratio: 0.0375 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.0-flash-thinking-exp-01-21":   {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.0-flash-exp-image-generation": {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},

		// Gemini 2.0 Pro Models
		"gemini-2.0-pro-exp-02-05": {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},

		// Gemini 2.5 Flash Models
		"gemini-2.5-flash-lite-preview-06-17": {Ratio: 0.0375 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-flash":                    {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-flash-preview-04-17":      {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-flash-preview-05-20":      {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},

		// Gemini 2.5 Pro Models
		"gemini-2.5-pro":               {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-pro-exp-03-25":     {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-pro-preview-05-06": {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-pro-preview-06-05": {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default Gemini pricing
	return 0.5 * 0.000001 // Default USD pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for Gemini
	return 3.0
}
