package ali

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
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

// ConvertImageRequest 转换 ImageRequest
func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	meta := meta.GetByContext(c)
	inputJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal input: %w, please check the request", err)
	}
	aliImageRequest := &AliImageRequest{}
	nestedJSON := map[string]any{
		"model":             request.Model,
		"input":             json.RawMessage(inputJSON),
		"parameters":        json.RawMessage(inputJSON),
		"resources":         request.Resources,
		"training_file_ids": request.TrainingFileIDs,
	}
	nestedJSONBytes, err := json.Marshal(nestedJSON)
	if err != nil {
		return nil, fmt.Errorf("marshal nested: %w, please check the request", err)
	}
	if err := json.Unmarshal(nestedJSONBytes, &aliImageRequest); err != nil {
		return nil, fmt.Errorf("unmarshal to AliImageRequest: %w, please check the request", err)
	}
	meta.OriginModelName = aliImageRequest.Model
	meta.ActualModelName = metalib.GetMappedModelName(aliImageRequest.Model, AliModelMapping)
	aliImageRequest.Model = meta.ActualModelName
	metalib.Set2Context(c, meta)
	if aliImageRequest.Parameters != nil && isZero(reflect.ValueOf(*aliImageRequest.Parameters)) {
		aliImageRequest.Parameters = nil //置为nil后,该字段可以在序列化时被自动删除.不确定是否必须
	}
	if aliImageRequest.Input != nil && isZero(reflect.ValueOf(*aliImageRequest.Input)) {
		aliImageRequest.Input = nil //置为nil后,该字段可以在序列化时被自动删除.不确定是否必须
	}
	// 设置图片数量(计费)
	if aliImageRequest.Input.GenerateNum != 0 {
		c.Set("temp_n", aliImageRequest.Input.GenerateNum)
	} else if aliImageRequest.Parameters != nil && aliImageRequest.Parameters.N != 0 {
		c.Set("temp_n", aliImageRequest.Parameters.N)
	} else {
		c.Set("temp_n", 1)
	}
	// 设置图片尺寸(计费)
	if aliImageRequest.Parameters != nil && aliImageRequest.Parameters.Size != "" {
		c.Set("temp_size", aliImageRequest.Parameters.Size)
	} else {
		c.Set("temp_size", "")
	}
	c.Set("temp_model", aliImageRequest.Model)
	c.Set("temp_quality", "")

	// if aliimageRequest.Resources != nil && len(*aliimageRequest.Resources) == 0 {
	// 	aliimageRequest.Resources = nil //因为Resources底层是切片 切片默认值已经是nil 就不需要特殊处理了
	// }
	// if aliimageRequest.TrainingFileIds != nil && len(*aliimageRequest.TrainingFileIds) == 0 {
	// 	aliimageRequest.TrainingFileIds = nil //置为nil后,该字段可以在序列化时被自动删除
	// }
	return aliImageRequest, nil
}

// isZero 检查值是否为零值（包括结构体是否所有字段为零值）
func isZero(v reflect.Value) bool {
	// 检查值是否无效
	if !v.IsValid() {
		return true
	}
	// 根据类型处理
	switch v.Kind() {
	case reflect.Ptr:
		// 指针类型：检查是否为 nil，若非 nil 则解引用
		if v.IsNil() {
			return true
		}
		return isZero(v.Elem())
	case reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan:
		// 这些类型支持 IsNil
		return v.IsNil()
	case reflect.Struct:
		// 检查结构体所有字段是否为零值
		for i := 0; i < v.NumField(); i++ {
			if !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		// 其他类型（int, string 等）使用 IsZero
		return v.IsZero()
	}
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
