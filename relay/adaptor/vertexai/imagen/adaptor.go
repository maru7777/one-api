package imagen

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on VertexAI Imagen pricing: https://cloud.google.com/vertex-ai/generative-ai/pricing
var ModelRatios = map[string]adaptor.ModelConfig{
	// -------------------------------------
	// Image Generation Models
	// -------------------------------------
	"imagen-3.0-generate-001":      {Ratio: 40.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.04 per image
	"imagen-3.0-generate-002":      {Ratio: 40.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.04 per image
	"imagen-3.0-fast-generate-001": {Ratio: 20.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.02 per image
	// -------------------------------------
	// Image Editing Models
	// -------------------------------------
	"imagen-3.0-capability-001": {Ratio: 50.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.05 per image
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *meta.Meta) {
	// No initialization needed
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	meta := meta.GetByContext(c)

	if request.ResponseFormat == nil || *request.ResponseFormat != "b64_json" {
		return nil, errors.New("only support b64_json response format")
	}
	if request.N <= 0 {
		request.N = 1 // Default to 1 if not specified
	}

	switch meta.Mode {
	case relaymode.ImagesGenerations:
		return convertImageCreateRequest(request)
	case relaymode.ImagesEdits:
		switch c.ContentType() {
		// case "application/json":
		// 	return ConvertJsonImageEditRequest(c)
		case "multipart/form-data":
			return ConvertMultipartImageEditRequest(c)
		default:
			return nil, errors.New("unsupported content type for image edit")
		}
	default:
		return nil, errors.New("not implemented")
	}
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, wrapErr *model.ErrorWithStatusCode) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	switch meta.Mode {
	case relaymode.ImagesEdits:
		return HandleImageEdit(c, resp)
	case relaymode.ImagesGenerations:
		return nil, handleImageGeneration(c, resp, respBody)
	default:
		return nil, openai.ErrorWrapper(errors.New("unsupported mode"), "unsupported_mode", http.StatusBadRequest)
	}
}

func handleImageGeneration(c *gin.Context, resp *http.Response, respBody []byte) *model.ErrorWithStatusCode {
	var imageResponse CreateImageResponse

	if resp.StatusCode != http.StatusOK {
		return openai.ErrorWrapper(errors.New(string(respBody)), "imagen_api_error", resp.StatusCode)
	}

	err := json.Unmarshal(respBody, &imageResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}

	// Convert to OpenAI format
	openaiResp := openai.ImageResponse{
		Created: time.Now().Unix(),
		Data:    make([]openai.ImageData, 0, len(imageResponse.Predictions)),
	}

	for _, prediction := range imageResponse.Predictions {
		openaiResp.Data = append(openaiResp.Data, openai.ImageData{
			B64Json: prediction.BytesBase64Encoded,
		})
	}

	respBytes, err := json.Marshal(openaiResp)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_failed", http.StatusInternalServerError)
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	_, err = c.Writer.Write(respBytes)
	if err != nil {
		return openai.ErrorWrapper(err, "write_response_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "vertex_ai_imagen"
}

func convertImageCreateRequest(request *model.ImageRequest) (any, error) {
	return CreateImageRequest{
		Instances: []createImageInstance{
			{
				Prompt: request.Prompt,
			},
		},
		Parameters: createImageParameters{
			SampleCount: request.N,
		},
	}, nil
}
