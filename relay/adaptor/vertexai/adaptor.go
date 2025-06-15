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

type Adaptor struct{}

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
