package ali

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/relay/adaptor"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

// https://help.aliyun.com/zh/dashscope/developer-reference/api-details

type Adaptor struct {
	meta *meta.Meta
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	fullRequestURL := ""
	switch meta.Mode {
	case relaymode.Embeddings:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/embeddings/text-embedding/text-embedding", meta.BaseURL)
	case relaymode.ImagesGenerations:
		switch meta.ActualModelName {
		case "wanx2.1-t2i-turbo", "wanx2.1-t2i-plus", "wanx2.0-t2i-turbo", "wanx-v1", "wanx-poster-generation-v1", "stable-diffusion-xl", "stable-diffusion-v1.5", "stable-diffusion-3.5-large", "stable-diffusion-3.5-large-turbo", "flux-schnell", "flux-dev", "flux-merged", "wanx-ast":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text2image/image-synthesis", meta.BaseURL)
		case "wanx2.1-imageedit", "wanx-sketch-to-image-lite", "wanx-x-painting", "image-instance-segmentation", "image-erase-completion", "aitryon", "aitryon-refiner", "aitryon-parsing-v1":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/image2image/image-synthesis", meta.BaseURL)
		case "wanx-style-repaint-v1", "wanx-style-cosplay-v1":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/image-generation/generation", meta.BaseURL)
		case "image-out-painting":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/image2image/out-painting", meta.BaseURL)
		case "wanx-virtualmodel", "virtualmodel-v2", "shoemodel-v1":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/virtualmodel/generation", meta.BaseURL)
		case "wanx-background-generation-v2":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/background-generation/generation", meta.BaseURL)
		case "facechain-generation":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/album/gen_potrait", meta.BaseURL)
		case "wordart-semantic":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/wordart/semantic", meta.BaseURL)
		case "wordart-texture":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/wordart/texture", meta.BaseURL)
		case "wordart-surnames":
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/wordart/surnames", meta.BaseURL)
		default:
			fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text-generation/generation", meta.BaseURL)
		}
	default:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text-generation/generation", meta.BaseURL)
	}

	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	if meta.IsStream {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-DashScope-SSE", "enable")
	}
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)

	if meta.Mode == relaymode.ImagesGenerations {
		req.Header.Set("X-DashScope-Async", "enable")
	}
	if a.meta.Config.Plugin != "" {
		req.Header.Set("X-DashScope-Plugin", a.meta.Config.Plugin)
	}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case relaymode.Embeddings:
		aliEmbeddingRequest := ConvertEmbeddingRequest(*request)
		return aliEmbeddingRequest, nil
	default:
		aliRequest := ConvertRequest(*request)
		return aliRequest, nil
	}
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	aliimageRequest := &ImageRequest{}
	err := common.UnmarshalBodyReusable(c, aliimageRequest)
	if err != nil {
		return nil, err
	}
	c.Set("temp_n", aliimageRequest.Parameters.N)
	c.Set("temp_model", aliimageRequest.Model)
	c.Set("temp_size", aliimageRequest.Parameters.Size)
	c.Set("temp_quality", "")
	// var isModelMapped bool
	a.meta.OriginModelName = aliimageRequest.Model
	aliimageRequest.Model = a.meta.ActualModelName
	// isModelMapped = a.meta.OriginModelName != a.meta.ActualModelName
	metalib.Set2Context(c, a.meta)
	// Convert the original image model
	aliimageRequest.Model = metalib.GetMappedModelName(aliimageRequest.Model, billingratio.ImageOriginModelName)
	return aliimageRequest, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, resp)
	} else {
		switch meta.Mode {
		case relaymode.Embeddings:
			err, usage = EmbeddingHandler(c, resp)
		case relaymode.ImagesGenerations:
			err, usage = ImageHandler(c, resp)
		default:
			err, usage = Handler(c, resp)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "ali"
}
