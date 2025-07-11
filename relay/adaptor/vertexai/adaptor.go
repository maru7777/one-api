package vertexai

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor"
	channelhelper "github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/vertexai/imagen"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	relayModel "github.com/songquanpeng/one-api/relay/model"
)

var _ adaptor.Adaptor = new(Adaptor)

const channelName = "vertexai"

// IsRequireGlobalEndpoint determines if the given model requires a global endpoint
//
//   - https://cloud.google.com/vertex-ai/generative-ai/docs/models/gemini/2-5-pro
//   - https://cloud.google.com/vertex-ai/generative-ai/docs/learn/locations#global-preview
func IsRequireGlobalEndpoint(model string) bool {
	// gemini-2.5-pro-preview models use global endpoint
	return strings.HasPrefix(model, "gemini-2.5-pro-preview")
}

type Adaptor struct {
}

func (a *Adaptor) Init(meta *meta.Meta) {
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	meta := meta.GetByContext(c)

	if request.ResponseFormat == nil || *request.ResponseFormat != "b64_json" {
		return nil, errors.New("only support b64_json response format")
	}

	adaptor := GetAdaptor(meta.ActualModelName)
	if adaptor == nil {
		return nil, errors.Errorf("cannot found vertex image adaptor for model %s", meta.ActualModelName)
	}

	return adaptor.ConvertImageRequest(c, request)
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	meta := meta.GetByContext(c)

	adaptor := GetAdaptor(meta.ActualModelName)
	if adaptor == nil {
		return nil, errors.Errorf("cannot found vertex chat adaptor for model %s", meta.ActualModelName)
	}

	return adaptor.ConvertRequest(c, relayMode, request)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	adaptor := GetAdaptor(meta.ActualModelName)
	if adaptor == nil {
		return nil, &relayModel.ErrorWithStatusCode{
			StatusCode: http.StatusInternalServerError,
			Error: relayModel.Error{
				Message: "adaptor not found",
			},
		}
	}
	return adaptor.DoResponse(c, resp, meta)
}

func (a *Adaptor) GetModelList() (models []string) {
	models = modelList
	return
}

func (a *Adaptor) GetChannelName() string {
	return channelName
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	var suffix string
	var location string
	var baseHost string

	switch {
	case strings.HasPrefix(meta.ActualModelName, "gemini"):
		if meta.IsStream {
			suffix = "streamGenerateContent?alt=sse"
		} else {
			suffix = "generateContent"
		}

		// Use global endpoint for models that require it
		if IsRequireGlobalEndpoint(meta.ActualModelName) {
			location = "global"
			baseHost = "aiplatform.googleapis.com"
		} else {
			location = meta.Config.Region
			baseHost = fmt.Sprintf("%s-aiplatform.googleapis.com", meta.Config.Region)
		}
	case slices.Contains(imagen.ModelList, meta.ActualModelName):
		return fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/imagen-3.0-generate-001:predict",
			meta.Config.Region, meta.Config.VertexAIProjectID, meta.Config.Region,
		), nil
	default:
		if meta.IsStream {
			suffix = "streamRawPredict?alt=sse"
		} else {
			suffix = "rawPredict"
		}
		location = meta.Config.Region
		baseHost = fmt.Sprintf("%s-aiplatform.googleapis.com", meta.Config.Region)
	}

	if meta.BaseURL != "" {
		return fmt.Sprintf(
			"%s/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
			meta.BaseURL,
			meta.Config.VertexAIProjectID,
			location,
			meta.ActualModelName,
			suffix,
		), nil
	}

	return fmt.Sprintf(
		"https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
		baseHost,
		meta.Config.VertexAIProjectID,
		location,
		meta.ActualModelName,
		suffix,
	), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	token, err := getToken(c, meta.ChannelId, meta.Config.VertexAIADC)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return channelhelper.DoRequestHelper(a, c, meta, requestBody)
}

// Pricing methods - VertexAI adapter manages its own model pricing
// VertexAI uses the same Gemini models but with Google Cloud pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.000001

	// Direct map definition - much easier to maintain and edit
	// Pricing from https://cloud.google.com/vertex-ai/generative-ai/pricing
	// VertexAI shares models with Gemini but may have different pricing tiers
	return map[string]adaptor.ModelPrice{
		// Gemini Pro Models (VertexAI)
		"gemini-pro":     {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 3},
		"gemini-1.0-pro": {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 3},

		// Gemma Models (VertexAI)
		"gemma-2-2b-it":  {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1.4},
		"gemma-2-9b-it":  {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1.4},
		"gemma-2-27b-it": {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1.4},
		"gemma-3-27b-it": {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1.4},

		// Gemini 1.5 Flash Models (VertexAI)
		"gemini-1.5-flash":    {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-1.5-flash-8b": {Ratio: 0.0375 * MilliTokensUsd, CompletionRatio: 4},

		// Gemini 1.5 Pro Models (VertexAI)
		"gemini-1.5-pro":              {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-1.5-pro-experimental": {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},

		// Embedding Models (VertexAI)
		"text-embedding-004": {Ratio: 0.00001 * MilliTokensUsd, CompletionRatio: 1},
		"aqa":                {Ratio: 1, CompletionRatio: 1},

		// Gemini 2.0 Flash Models (VertexAI)
		"gemini-2.0-flash":                      {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.0-flash-exp":                  {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.0-flash-lite":                 {Ratio: 0.0375 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.0-flash-thinking-exp-01-21":   {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.0-flash-exp-image-generation": {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},

		// Gemini 2.0 Pro Models (VertexAI)
		"gemini-2.0-pro-exp-02-05": {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},

		// Gemini 2.5 Flash Models (VertexAI)
		"gemini-2.5-flash-lite-preview-06-17": {Ratio: 0.0375 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-flash":                    {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-flash-preview-04-17":      {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-flash-preview-05-20":      {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 4},

		// Gemini 2.5 Pro Models (VertexAI)
		"gemini-2.5-pro":               {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-pro-exp-03-25":     {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-pro-preview-05-06": {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},
		"gemini-2.5-pro-preview-06-05": {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 4},

		// VertexAI Claude Models (if supported)
		"claude-3-5-sonnet-20241022": {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},
		"claude-3-5-sonnet-20240620": {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},
		"claude-3-5-haiku-20241022":  {Ratio: 1 * MilliTokensUsd, CompletionRatio: 5},
		"claude-3-opus-20240229":     {Ratio: 15 * MilliTokensUsd, CompletionRatio: 5},
		"claude-3-sonnet-20240229":   {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},
		"claude-3-haiku-20240307":    {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 5},

		// VertexAI Imagen Models (image generation)
		"imagen-3.0-generate-001": {Ratio: 0.04 * MilliTokensUsd, CompletionRatio: 1}, // Per image pricing

		// VertexAI Veo Models (video generation)
		"veo-001": {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1}, // Per video pricing
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default VertexAI pricing (similar to Gemini)
	return 0.5 * 0.000001 // Default USD pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for VertexAI
	return 3.0
}
