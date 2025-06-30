package replicate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

type Adaptor struct {
	meta *meta.Meta
}

// ConvertImageRequest implements adaptor.Adaptor.
// func (a *Adaptor) ConvertImageRequest(_ *gin.Context, request *model.ImageRequest) (any, error) {
// 	return nil, errors.New("should call replicate.ConvertImageRequest instead")
// }

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	meta := meta.GetByContext(c)
	// if request.ResponseFormat != "b64_json" {
	//  return nil, errors.New("only support b64_json response format")
	// }
	// if request.N != 1 && request.N != 0 {
	//  return nil, errors.New("only support N=1")
	// }
	// 首先将原始请求序列化为JSON
	inputJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal input: %w, please check the request", err)
	}
	// switch meta.Mode {
	// case relaymode.ImagesGenerations:
	//  return convertImageCreateRequest(request)
	// case relaymode.ImagesEdits:
	//  return convertImageRemixRequest(c)
	// default:
	//  return nil, errors.New("not implemented")
	// }
	// 创建新的Replicate图像请求对象
	replicateImageRequest := &replicateImageRequest{}

	// 创建嵌套的JSON结构
	nestedJSON := map[string]any{
		// "model": request.Model,
		"input": json.RawMessage(inputJSON),
	}

	// 将嵌套JSON序列化
	nestedJSONBytes, err := json.Marshal(nestedJSON)
	if err != nil {
		return nil, fmt.Errorf("marshal nested: %w, please check the request", err)
	}
	// if replicateImageRequest.Input.NumberOfImages != nil {
	//     *replicateImageRequest.Input.NumberOfImages = 1
	// } else if replicateImageRequest.Input.NumOutputs != nil {
	//     *replicateImageRequest.Input.NumOutputs = 1

	// 反序列化到Replicate请求结构
	if err := json.Unmarshal(nestedJSONBytes, &replicateImageRequest); err != nil {
		return nil, fmt.Errorf("unmarshal to ReplicateImageRequest: %w, please check the request", err)
	}

	// 也可以从原始请求体直接解析
	// rawRequest := &ImageRequest{}
	// err = common.UnmarshalBodyReusable(c, rawRequest)
	// if err == nil {
	// 	// 如果成功解析，合并字段
	// 	replicateImageRequest = rawRequest
	// }

	// 记录原始模型名称 其实中间件里已经做过了 但是为了以往万一 再覆盖一次
	meta.OriginModelName = request.Model
	// 使用映射表获取实际模型名称 比如输入模型名称是 flux-dev 实际模型名称是 black-forest-labs/flux-dev
	meta.ActualModelName = metalib.GetMappedModelName(request.Model, ReplicateModelMapping)
	// 设置图像生成数量(计费)
	if replicateImageRequest.Input.NumOutputs != 0 {
		meta.VendorContext["PicNumber"] = replicateImageRequest.Input.NumOutputs
	} else if replicateImageRequest.Input.NumberOfImages != 0 {
		meta.VendorContext["PicNumber"] = replicateImageRequest.Input.NumberOfImages
	} else {
		meta.VendorContext["PicNumber"] = 1
	}
	// 设置图像尺寸，如果有的话(计费)
	if replicateImageRequest.Input.Size != "" {
		meta.VendorContext["PicSize"] = replicateImageRequest.Input.Size
	} else if replicateImageRequest.Input.Resolution != "" {
		meta.VendorContext["PicSize"] = replicateImageRequest.Input.Resolution
	} else {
		meta.VendorContext["PicSize"] = ""
	}
	meta.VendorContext["Model"] = meta.ActualModelName
	meta.VendorContext["Quality"] = ""
	metalib.Set2Context(c, meta)
	return replicateImageRequest, nil
}

// func convertImageCreateRequest(request *model.ImageRequest) (any, error) {
// 	return DrawImageRequest{
// 		Input: ImageInput{
// 			Steps:           25,
// 			Prompt:          request.Prompt,
// 			Guidance:        3,
// 			Seed:            int(time.Now().UnixNano()),
// 			SafetyTolerance: 5,
// 			NImages:         1, // replicate will always return 1 image
// 			Width:           1440,
// 			Height:          1440,
// 			AspectRatio:     "1:1",
// 		},
// 	}, nil
// }

// func convertImageRemixRequest(c *gin.Context) (any, error) {
// 	// recover request body
// 	requestBody, err := common.GetRequestBody(c)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "get request body")
// 	}
// 	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

// 	rawReq := new(model.OpenaiImageEditRequest)
// 	if err := c.ShouldBind(rawReq); err != nil {
// 		return nil, errors.Wrap(err, "parse image edit form")
// 	}

// 	return Convert2FluxRemixRequest(rawReq)
// }

// ConvertRequest converts the request to the format that the target API expects.
func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if !request.Stream {
		// TODO: support non-stream mode
		return nil, errors.Errorf("replicate models only support stream mode now, please set stream=true")
	}

	// Build the prompt from OpenAI messages
	var promptBuilder strings.Builder
	for _, message := range request.Messages {
		switch msgCnt := message.Content.(type) {
		case string:
			promptBuilder.WriteString(message.Role)
			promptBuilder.WriteString(": ")
			promptBuilder.WriteString(msgCnt)
			promptBuilder.WriteString("\n")
		default:
		}
	}

	replicateRequest := ReplicateChatRequest{
		Input: ChatInput{
			Prompt:           promptBuilder.String(),
			MaxTokens:        request.MaxTokens,
			Temperature:      1.0,
			TopP:             1.0,
			PresencePenalty:  0.0,
			FrequencyPenalty: 0.0,
		},
	}

	// Map optional fields
	if request.Temperature != nil {
		replicateRequest.Input.Temperature = *request.Temperature
	}
	if request.TopP != nil {
		replicateRequest.Input.TopP = *request.TopP
	}
	if request.PresencePenalty != nil {
		replicateRequest.Input.PresencePenalty = *request.PresencePenalty
	}
	if request.FrequencyPenalty != nil {
		replicateRequest.Input.FrequencyPenalty = *request.FrequencyPenalty
	}
	if request.MaxTokens > 0 {
		replicateRequest.Input.MaxTokens = request.MaxTokens
	} else if request.MaxTokens == 0 {
		replicateRequest.Input.MaxTokens = 500
	}

	return replicateRequest, nil
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	if !slices.Contains(ModelList, meta.ActualModelName) {
		return "", errors.Errorf("model %s not supported", meta.ActualModelName)
	}

	return fmt.Sprintf("https://api.replicate.com/v1/models/%s/predictions", meta.ActualModelName), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	logger.Info(c, "send request to replicate")
	resp, err := adaptor.DoRequestHelper(a, c, meta, requestBody)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return resp, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	switch meta.Mode {
	case relaymode.ImagesGenerations,
		relaymode.ImagesEdits:
		err, usage = ImageHandler(c, resp)
	case relaymode.ChatCompletions:
		err, usage = ChatHandler(c, resp)
	default:
		err = openai.ErrorWrapper(errors.New("not implemented"), "not_implemented", http.StatusInternalServerError)
	}

	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "replicate"
}
