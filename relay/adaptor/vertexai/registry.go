package vertexai

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor/geminiOpenaiCompatible"
	claude "github.com/songquanpeng/one-api/relay/adaptor/vertexai/claude"
	gemini "github.com/songquanpeng/one-api/relay/adaptor/vertexai/gemini"
	"github.com/songquanpeng/one-api/relay/adaptor/vertexai/imagen"
	"github.com/songquanpeng/one-api/relay/adaptor/vertexai/veo"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type VertexAIModelType int

const (
	VertexAIClaude VertexAIModelType = iota + 1
	VertexAIGemini
	VertexAIImagen
	VertexAIVeo
)

var modelMapping = map[string]VertexAIModelType{}
var modelList = []string{}

func init() {
	// register vertex claude models
	modelList = append(modelList, claude.ModelList...)
	for _, model := range claude.ModelList {
		modelMapping[model] = VertexAIClaude
	}

	// register vertex gemini models
	modelList = append(modelList, geminiOpenaiCompatible.ModelList...)
	for _, model := range geminiOpenaiCompatible.ModelList {
		modelMapping[model] = VertexAIGemini
	}

	// register vertex imagen models
	modelList = append(modelList, imagen.ModelList...)
	for _, model := range imagen.ModelList {
		modelMapping[model] = VertexAIImagen
	}

	// register vertex veo models
	modelList = append(modelList, veo.ModelList...)
	for _, model := range veo.ModelList {
		modelMapping[model] = VertexAIVeo
	}
}

type innerAIAdapter interface {
	ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error)
	ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error)
	DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode)
}

func GetAdaptor(model string) innerAIAdapter {
	adaptorType := modelMapping[model]
	switch adaptorType {
	case VertexAIClaude:
		return &claude.Adaptor{}
	case VertexAIGemini:
		return &gemini.Adaptor{}
	case VertexAIImagen:
		return &imagen.Adaptor{}
	case VertexAIVeo:
		return &veo.Adaptor{}
	default:
		return nil
	}
}
