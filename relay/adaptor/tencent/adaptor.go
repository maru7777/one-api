package tencent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

// https://cloud.tencent.com/document/api/1729/101837

type Adaptor struct {
	// Sign      string
	// Action    string
	// Version   string
	// Timestamp int64
}

func (a *Adaptor) Init(meta *meta.Meta) {
	// meta.VendorContext["Action"] = "ChatCompletions"
	// meta.VendorContext["Version"] = "2023-09-01"
	// meta.VendorContext["Timestamp"] = helper.GetTimestamp()
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ImagesGenerations:
		switch meta.VendorContext["Action"] {
		case "SubmitHunyuanImageJob", "QueryHunyuanImageJob", "SubmitHunyuanImageChatJob", "QueryHunyuanImageChatJob", "TextToImageLite", "SubmitHunyuanTo3DJob", "QueryHunyuanTo3DJob":
			meta.VendorContext["Host"] = "hunyuan.tencentcloudapi.com"
			return "https://hunyuan.tencentcloudapi.com/", nil
		case "SubmitTrainPortraitModelJob", "QueryTrainPortraitModelJob", "SubmitMemeJob", "QueryMemeJob", "SubmitGlamPicJob", "QueryGlamPicJob", "ChangeClothes", "ReplaceBackground", "SketchToImage", "RefineImage", "ImageInpaintingRemoval", "ImageOutpainting":
			meta.VendorContext["Host"] = "aiart.tencentcloudapi.com"
			return "https://aiart.tencentcloudapi.com/", nil
		default:
			return "", fmt.Errorf("unsupported Action: %s", meta.VendorContext["Action"])
		}
	default:
		return meta.BaseURL + "/", nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	// adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", meta.VendorContext["Sign"].(string))
	req.Header.Set("X-TC-Action", meta.VendorContext["Action"].(string))
	req.Header.Set("X-TC-Version", meta.VendorContext["Version"].(string))
	req.Header.Set("X-TC-Timestamp", meta.VendorContext["Timestamp"].(string))
	req.Header.Set("X-TC-Region", meta.VendorContext["Region"].(string))
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	meta := metalib.GetByContext(c)
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	_, secretId, secretKey, err := ParseConfig(apiKey)
	meta.VendorContext["SecretId"] = secretId
	meta.VendorContext["SecretKey"] = secretKey
	if err != nil {
		return nil, err
	}
	var convertedRequest any
	switch relayMode {
	case relaymode.Embeddings:
		meta.VendorContext["Action"] = "GetEmbedding"
		convertedRequest = ConvertEmbeddingRequest(*request)
	default:
		meta.VendorContext["Action"] = "ChatCompletions"
		convertedRequest = ConvertRequest(*request)
	}
	meta.VendorContext["Version"] = "2023-09-01"
	meta.VendorContext["Host"] = "hunyuan.tencentcloudapi.com"
	// tencentImageRequest.Model,tencentImageRequest.Action, tencentImageRequest.Version, tencentImageRequest.Region = "", "", ""
	meta.VendorContext["Sign"] = GetSign(convertedRequest, meta) //为了保持时间戳一致 先获取sign 再从meta.VendorContext["Timestamp"]获取相同的时间
	meta.VendorContext["Timestamp"] = meta.VendorContext["Timestamp"].(string)
	meta.VendorContext["Region"] = "ap-guangzhou"
	return convertedRequest, nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	inputJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal input: %w, please check the request", err)
	}
	tencentImageRequest := &TenentImageRequest{}
	if err := json.Unmarshal(inputJSON, &tencentImageRequest); err != nil {
		return nil, fmt.Errorf("unmarshal to TencentImageRequest: %w, please check the request", err)
	}
	//根据模型获取action
	action := modelToActionMap[tencentImageRequest.Model]
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "Bearer ") //获取真实apikey十分奇怪 没有写入meta 而是写到了请求头
	_, secretId, secretKey, err := ParseConfig(apiKey)
	if err != nil {
		return nil, err
	}
	meta := metalib.GetByContext(c)
	meta.VendorContext["SecretId"] = secretId
	meta.VendorContext["SecretKey"] = secretKey
	meta.VendorContext["Action"] = action
	switch action {
	case "SubmitTrainPortraitModelJob", "QueryTrainPortraitModelJob", "SubmitMemeJob", "QueryMemeJob", "SubmitGlamPicJob", "QueryGlamPicJob", "ChangeClothes", "ReplaceBackground", "SketchToImage", "RefineImage", "ImageInpaintingRemoval", "ImageOutpainting", "ImageToImage", "GenerateAvatar", "UploadTrainPortraitImages", "SubmitDrawPortraitJob", "QueryDrawPortraitJob":
		meta.VendorContext["Version"] = "2022-12-29"
	case "SubmitHunyuanImageJob", "QueryHunyuanImageJob", "SubmitHunyuanImageChatJob", "QueryHunyuanImageChatJob", "TextToImageLite", "SubmitHunyuanTo3DJob", "QueryHunyuanTo3DJob":
		meta.VendorContext["Version"] = "2023-09-01"
	default:
		return nil, fmt.Errorf("unsupported action: %s", tencentImageRequest.Action)
	}
	switch action {
	case "SubmitHunyuanImageJob", "QueryHunyuanImageJob", "SubmitHunyuanImageChatJob", "QueryHunyuanImageChatJob", "TextToImageLite", "SubmitHunyuanTo3DJob", "QueryHunyuanTo3DJob":
		meta.VendorContext["Host"] = "hunyuan.tencentcloudapi.com"
	case "SubmitTrainPortraitModelJob", "QueryTrainPortraitModelJob", "SubmitMemeJob", "QueryMemeJob", "SubmitGlamPicJob", "QueryGlamPicJob", "ChangeClothes", "ReplaceBackground", "SketchToImage", "RefineImage", "ImageInpaintingRemoval", "ImageOutpainting":
		meta.VendorContext["Host"] = "aiart.tencentcloudapi.com"
	default:
		return nil, fmt.Errorf("unsupported action: %s", tencentImageRequest.Action)
	}
	tencentImageRequest.Model, tencentImageRequest.Action, tencentImageRequest.Version, tencentImageRequest.Region = "", "", "", "" //让他们序列化后消失
	meta.VendorContext["Sign"] = GetSign(tencentImageRequest, meta)
	meta.VendorContext["Timestamp"] = meta.VendorContext["Timestamp"].(string)
	meta.VendorContext["Region"] = "ap-guangzhou"
	// 设置计费相关
	if tencentImageRequest.Num != 0 {
		meta.VendorContext["PicNumber"] = tencentImageRequest.Num
	} else if tencentImageRequest.ImageNum != 0 {
		meta.VendorContext["PicNumber"] = tencentImageRequest.ImageNum
	} else {
		meta.VendorContext["PicNumber"] = 1
	}
	if tencentImageRequest.Resolution != "" {
		meta.VendorContext["PicSize"] = tencentImageRequest.Resolution
	} else {
		meta.VendorContext["PicSize"] = ""
	}
	meta.VendorContext["Model"] = tencentImageRequest.Model //如有另外除了model-action之外还有别的模型映射时 规则是这里得是真实的模型名称
	meta.VendorContext["Quality"] = tencentImageRequest.Clarity
	return tencentImageRequest, nil
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// Convert Claude Messages API request to OpenAI format first
	openaiRequest := &model.GeneralOpenAIRequest{
		Model:       request.Model,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Stream:      request.Stream != nil && *request.Stream,
		Stop:        request.StopSequences,
	}

	// Convert system prompt
	if request.System != nil {
		switch system := request.System.(type) {
		case string:
			if system != "" {
				openaiRequest.Messages = append(openaiRequest.Messages, model.Message{
					Role:    "system",
					Content: system,
				})
			}
		case []any:
			// For structured system content, extract text parts
			var systemParts []string
			for _, block := range system {
				if blockMap, ok := block.(map[string]any); ok {
					if text, exists := blockMap["text"]; exists {
						if textStr, ok := text.(string); ok {
							systemParts = append(systemParts, textStr)
						}
					}
				}
			}
			if len(systemParts) > 0 {
				systemText := strings.Join(systemParts, "\n")
				openaiRequest.Messages = append(openaiRequest.Messages, model.Message{
					Role:    "system",
					Content: systemText,
				})
			}
		}
	}

	// Convert messages
	for _, msg := range request.Messages {
		openaiMessage := model.Message{
			Role: msg.Role,
		}

		// Convert content based on type
		switch content := msg.Content.(type) {
		case string:
			// Simple string content
			openaiMessage.Content = content
		case []any:
			// Structured content blocks - convert to OpenAI format
			var contentParts []model.MessageContent
			for _, block := range content {
				if blockMap, ok := block.(map[string]any); ok {
					if blockType, exists := blockMap["type"]; exists {
						switch blockType {
						case "text":
							if text, exists := blockMap["text"]; exists {
								if textStr, ok := text.(string); ok {
									contentParts = append(contentParts, model.MessageContent{
										Type: "text",
										Text: &textStr,
									})
								}
							}
						case "image":
							if source, exists := blockMap["source"]; exists {
								if sourceMap, ok := source.(map[string]any); ok {
									imageURL := model.ImageURL{}
									if mediaType, exists := sourceMap["media_type"]; exists {
										if data, exists := sourceMap["data"]; exists {
											if dataStr, ok := data.(string); ok {
												// Convert to data URL format
												imageURL.Url = fmt.Sprintf("data:%s;base64,%s", mediaType, dataStr)
											}
										}
									}
									contentParts = append(contentParts, model.MessageContent{
										Type:     "image_url",
										ImageURL: &imageURL,
									})
								}
							}
						}
					}
				}
			}
			if len(contentParts) > 0 {
				openaiMessage.Content = contentParts
			}
		default:
			// Fallback: convert to string
			if contentBytes, err := json.Marshal(content); err == nil {
				openaiMessage.Content = string(contentBytes)
			}
		}

		openaiRequest.Messages = append(openaiRequest.Messages, openaiMessage)
	}

	// Convert tools
	for _, tool := range request.Tools {
		openaiTool := model.Tool{
			Type: "function",
			Function: model.Function{
				Name:        tool.Name,
				Description: tool.Description,
			},
		}

		// Convert input schema
		if tool.InputSchema != nil {
			if schemaMap, ok := tool.InputSchema.(map[string]any); ok {
				openaiTool.Function.Parameters = schemaMap
			}
		}

		openaiRequest.Tools = append(openaiRequest.Tools, openaiTool)
	}

	// Convert tool choice
	if request.ToolChoice != nil {
		openaiRequest.ToolChoice = request.ToolChoice
	}

	// Mark this as a Claude Messages conversion for response handling
	c.Set(ctxkey.ClaudeMessagesConversion, true)
	c.Set(ctxkey.OriginalClaudeRequest, request)

	// Now convert using Tencent's existing logic
	return a.ConvertRequest(c, relaymode.ChatCompletions, openaiRequest)
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
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
	if meta.IsStream {
		var responseText string
		err, responseText = StreamHandler(c, resp)
		usage = openai.ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
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
	return adaptor.GetModelListFromPricing(ModelRatios)
}

func (a *Adaptor) GetChannelName() string {
	return "tencent"
}

// Pricing methods - Tencent adapter manages its own model pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelConfig {
	const MilliRmb = 0.0001

	// Direct map definition - much easier to maintain and edit
	// Pricing from https://cloud.tencent.com/document/product/1729/97731
	return ModelRatios
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default Tencent pricing
	return 4.5 * 0.0001 // Default RMB pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for Tencent
	return 1.0
}
