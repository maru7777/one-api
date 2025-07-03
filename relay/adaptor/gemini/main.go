package gemini

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/image"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/adaptor/geminiOpenaiCompatible"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model"
)

// https://ai.google.dev/docs/gemini_api_overview?hl=zh-cn

const (
	VisionMaxImageNum = 16
)

var mimeTypeMap = map[string]string{
	"json_object": "application/json",
	"text":        "text/plain",
}

// cleanJsonSchemaForGemini removes unsupported fields and converts types for Gemini API compatibility
func cleanJsonSchemaForGemini(schema interface{}) interface{} {
	switch v := schema.(type) {
	case map[string]interface{}:
		cleaned := make(map[string]interface{})

		// List of supported fields in Gemini (from official documentation)
		supportedFields := map[string]bool{
			"anyOf": true, "enum": true, "format": true, "items": true,
			"maximum": true, "maxItems": true, "minimum": true, "minItems": true,
			"nullable": true, "properties": true, "propertyOrdering": true,
			"required": true, "type": true,
		}

		// Type mapping from lowercase to uppercase (Gemini requirement)
		typeMapping := map[string]string{
			"object":  "OBJECT",
			"array":   "ARRAY",
			"string":  "STRING",
			"number":  "NUMBER",
			"integer": "INTEGER",
			"boolean": "BOOLEAN",
			"null":    "NULL",
		}

		for key, value := range v {
			// Skip unsupported fields like additionalProperties, description, strict
			if !supportedFields[key] {
				continue
			}

			switch key {
			case "type":
				// Convert type to uppercase if it's a string
				if typeStr, ok := value.(string); ok {
					if mappedType, exists := typeMapping[strings.ToLower(typeStr)]; exists {
						cleaned[key] = mappedType
					} else {
						cleaned[key] = strings.ToUpper(typeStr)
					}
				} else {
					cleaned[key] = value
				}
			case "properties":
				// Handle properties object - recursively clean each property
				if props, ok := value.(map[string]interface{}); ok {
					cleanedProps := make(map[string]interface{})
					for propKey, propValue := range props {
						cleanedProps[propKey] = cleanJsonSchemaForGemini(propValue)
					}
					cleaned[key] = cleanedProps
				} else {
					cleaned[key] = value
				}
			case "items":
				// Handle array items schema - recursively clean
				cleaned[key] = cleanJsonSchemaForGemini(value)
			default:
				// For other supported fields, recursively clean if they're objects/arrays
				cleaned[key] = cleanJsonSchemaForGemini(value)
			}
		}
		return cleaned
	case []interface{}:
		// Clean arrays recursively
		cleaned := make([]interface{}, len(v))
		for i, item := range v {
			cleaned[i] = cleanJsonSchemaForGemini(item)
		}
		return cleaned
	default:
		// Return primitive values as-is
		return v
	}
}

// Setting safety to the lowest possible values since Gemini is already powerless enough
func ConvertRequest(textRequest model.GeneralOpenAIRequest) *ChatRequest {
	geminiRequest := ChatRequest{
		Contents: make([]ChatContent, 0, len(textRequest.Messages)),
		SafetySettings: []ChatSafetySettings{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_CIVIC_INTEGRITY",
				Threshold: config.GeminiSafetySetting,
			},
		},
		GenerationConfig: ChatGenerationConfig{
			Temperature:        textRequest.Temperature,
			TopP:               textRequest.TopP,
			MaxOutputTokens:    textRequest.MaxTokens,
			ResponseModalities: geminiOpenaiCompatible.GetModelModalities(textRequest.Model),
		},
	}
	if textRequest.ResponseFormat != nil {
		if mimeType, ok := mimeTypeMap[textRequest.ResponseFormat.Type]; ok {
			geminiRequest.GenerationConfig.ResponseMimeType = mimeType
		}
		if textRequest.ResponseFormat.JsonSchema != nil {
			// Clean the schema to remove unsupported properties for Gemini
			cleanedSchema := cleanJsonSchemaForGemini(textRequest.ResponseFormat.JsonSchema.Schema)
			geminiRequest.GenerationConfig.ResponseSchema = cleanedSchema
			geminiRequest.GenerationConfig.ResponseMimeType = mimeTypeMap["json_object"]
		}
	}

	// FIX(https://github.com/Laisky/one-api/issues/60):
	// Gemini's function call supports fewer parameters than OpenAI's,
	// so a conversion is needed here to keep only the parameters supported by Gemini.
	if textRequest.Tools != nil {
		convertedGeminiFunctions := make([]model.Function, 0, len(textRequest.Tools))
		for _, tool := range textRequest.Tools {
			delete(tool.Function.Parameters, "additionalProperties")
			convertedGeminiFunctions = append(convertedGeminiFunctions, model.Function{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
				Required:    tool.Function.Required,
			})
		}
		geminiRequest.Tools = []ChatTools{
			{
				FunctionDeclarations: convertedGeminiFunctions,
			},
		}
	} else if textRequest.Functions != nil {
		for _, function := range textRequest.Functions {
			delete(function.Parameters, "additionalProperties")
			geminiRequest.Tools = append(geminiRequest.Tools, ChatTools{
				FunctionDeclarations: []model.Function{
					{
						Name:        function.Name,
						Description: function.Description,
						Parameters:  function.Parameters,
						Required:    function.Required,
					},
				},
			})
		}
	}

	if geminiRequest.GenerationConfig.TopP != nil &&
		(*geminiRequest.GenerationConfig.TopP < 0 || *geminiRequest.GenerationConfig.TopP > 1) {
		geminiRequest.GenerationConfig.TopP = nil
	}

	shouldAddDummyModelMessage := false
	for _, message := range textRequest.Messages {
		// Start with initial content based on message string content
		initialText := message.StringContent()
		var parts []Part

		// Add text content if it's not empty
		if initialText != "" {
			parts = append(parts, Part{
				Text: initialText,
			})
		}

		// Handle OpenAI tool calls - convert them to Gemini function calls
		if len(message.ToolCalls) > 0 {
			for _, toolCall := range message.ToolCalls {
				// Parse the arguments from JSON string to interface{}
				var args interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments.(string)), &args); err != nil {
					// If parsing fails, use the raw string
					args = toolCall.Function.Arguments
				}

				parts = append(parts, Part{
					FunctionCall: &FunctionCall{
						FunctionName: toolCall.Function.Name,
						Arguments:    args,
					},
				})
			}
		}

		// Parse structured content and add additional parts
		openaiContent := message.ParseContent()
		imageNum := 0
		for _, part := range openaiContent {
			if part.Type == model.ContentTypeText && part.Text != nil && *part.Text != "" {
				// Only add if we haven't already added this text from StringContent()
				if *part.Text != initialText {
					parts = append(parts, Part{
						Text: *part.Text,
					})
				}
			} else if part.Type == model.ContentTypeImageURL {
				imageNum += 1
				if imageNum > VisionMaxImageNum {
					continue
				}
				mimeType, data, _ := image.GetImageFromUrl(part.ImageURL.Url)
				parts = append(parts, Part{
					InlineData: &InlineData{
						MimeType: mimeType,
						Data:     data,
					},
				})
			}
		}

		// If we have no parts at all (empty content with tool calls), add a minimal text part
		// to satisfy Gemini's requirement that parts cannot be empty
		if len(parts) == 0 {
			parts = append(parts, Part{
				Text: " ", // Minimal non-empty text to satisfy Gemini's requirements
			})
		}

		content := ChatContent{
			Role:  message.Role,
			Parts: parts,
		}

		// there's no assistant role in gemini and API shall vomit if Role is not user or model
		if content.Role == "assistant" {
			content.Role = "model"
		}
		// Converting system prompt to prompt from user for the same reason
		if content.Role == "system" {
			shouldAddDummyModelMessage = true
			if IsModelSupportSystemInstruction(textRequest.Model) {
				geminiRequest.SystemInstruction = &content
				geminiRequest.SystemInstruction.Role = ""
				continue
			} else {
				content.Role = "user"
			}
		}
		// Handle tool responses - convert to user role with function response format
		if content.Role == "tool" {
			// Tool responses in OpenAI are converted to user messages in Gemini
			// with the function response content
			content.Role = "user"
			// Keep the original content as text, as Gemini expects function responses
			// to be handled in a different format than OpenAI
		}

		geminiRequest.Contents = append(geminiRequest.Contents, content)

		// If a system message is the last message, we need to add a dummy model message to make gemini happy
		if shouldAddDummyModelMessage {
			geminiRequest.Contents = append(geminiRequest.Contents, ChatContent{
				Role: "model",
				Parts: []Part{
					{
						Text: "Okay",
					},
				},
			})
			shouldAddDummyModelMessage = false
		}
	}

	return &geminiRequest
}

func ConvertEmbeddingRequest(request model.GeneralOpenAIRequest) *BatchEmbeddingRequest {
	inputs := request.ParseInput()
	requests := make([]EmbeddingRequest, len(inputs))
	model := fmt.Sprintf("models/%s", request.Model)

	for i, input := range inputs {
		requests[i] = EmbeddingRequest{
			Model: model,
			Content: ChatContent{
				Parts: []Part{
					{
						Text: input,
					},
				},
			},
		}
	}

	return &BatchEmbeddingRequest{
		Requests: requests,
	}
}

type ChatResponse struct {
	Candidates     []ChatCandidate    `json:"candidates"`
	PromptFeedback ChatPromptFeedback `json:"promptFeedback"`
}

func (g *ChatResponse) GetResponseText() string {
	if g == nil {
		return ""
	}
	if len(g.Candidates) > 0 && len(g.Candidates[0].Content.Parts) > 0 {
		return g.Candidates[0].Content.Parts[0].Text
	}
	return ""
}

type ChatCandidate struct {
	Content       ChatContent        `json:"content"`
	FinishReason  string             `json:"finishReason"`
	Index         int64              `json:"index"`
	SafetyRatings []ChatSafetyRating `json:"safetyRatings"`
}

type ChatSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type ChatPromptFeedback struct {
	SafetyRatings []ChatSafetyRating `json:"safetyRatings"`
}

func getToolCalls(candidate *ChatCandidate) []model.Tool {
	var toolCalls []model.Tool

	item := candidate.Content.Parts[0]
	if item.FunctionCall == nil {
		return toolCalls
	}
	argsBytes, err := json.Marshal(item.FunctionCall.Arguments)
	if err != nil {
		logger.FatalLog("getToolCalls failed: " + err.Error())
		return toolCalls
	}
	toolCall := model.Tool{
		Id:   fmt.Sprintf("call_%s", random.GetUUID()),
		Type: "function",
		Function: model.Function{
			Arguments: string(argsBytes),
			Name:      item.FunctionCall.FunctionName,
		},
	}
	toolCalls = append(toolCalls, toolCall)
	return toolCalls
}

func responseGeminiChat2OpenAI(response *ChatResponse) *openai.TextResponse {
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", random.GetUUID()),
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: make([]openai.TextResponseChoice, 0, len(response.Candidates)),
	}
	for i, candidate := range response.Candidates {
		choice := openai.TextResponseChoice{
			Index: i,
			Message: model.Message{
				Role: "assistant",
			},
			FinishReason: constant.StopFinishReason,
		}
		if len(candidate.Content.Parts) > 0 {
			if candidate.Content.Parts[0].FunctionCall != nil {
				choice.Message.ToolCalls = getToolCalls(&candidate)
			} else {
				// Handle text and image content
				var builder strings.Builder
				var contentItems []model.MessageContent

				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						// For text parts
						if i > 0 {
							builder.WriteString("\n")
						}
						builder.WriteString(part.Text)

						// Add to content items
						contentItems = append(contentItems, model.MessageContent{
							Type: model.ContentTypeText,
							Text: &part.Text,
						})
					}

					if part.InlineData != nil && part.InlineData.MimeType != "" && part.InlineData.Data != "" {
						// For inline image data
						imageURL := &model.ImageURL{
							// The data is already base64 encoded
							Url: fmt.Sprintf("data:%s;base64,%s", part.InlineData.MimeType, part.InlineData.Data),
						}

						contentItems = append(contentItems, model.MessageContent{
							Type:     model.ContentTypeImageURL,
							ImageURL: imageURL,
						})
					}
				}

				// If we have multiple content types, use structured content format
				if len(contentItems) > 1 || (len(contentItems) == 1 && contentItems[0].Type != model.ContentTypeText) {
					choice.Message.Content = contentItems
				} else {
					// Otherwise use the simple string content format
					choice.Message.Content = builder.String()
				}
			}
		} else {
			choice.Message.Content = ""
			choice.FinishReason = candidate.FinishReason
		}

		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponseGeminiChat2OpenAI(geminiResponse *ChatResponse) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Role = "assistant"

	// Check if we have any candidates
	if len(geminiResponse.Candidates) == 0 {
		return nil
	}

	// Get the first candidate
	candidate := geminiResponse.Candidates[0]

	// Check if there are parts in the content
	if len(candidate.Content.Parts) == 0 {
		return nil
	}

	// Handle different content types in the parts
	for _, part := range candidate.Content.Parts {
		// Handle text content
		if part.Text != "" {
			// Store as string for simple text responses
			textContent := part.Text
			choice.Delta.Content = textContent
		}

		// Handle image content
		if part.InlineData != nil && part.InlineData.MimeType != "" && part.InlineData.Data != "" {
			// Create a structured response for image content
			imageUrl := fmt.Sprintf("data:%s;base64,%s", part.InlineData.MimeType, part.InlineData.Data)

			// If we already have text content, create a mixed content response
			if strContent, ok := choice.Delta.Content.(string); ok && strContent != "" {
				// Convert the existing text content and add the image
				messageContents := []model.MessageContent{
					{
						Type: model.ContentTypeText,
						Text: &strContent,
					},
					{
						Type: model.ContentTypeImageURL,
						ImageURL: &model.ImageURL{
							Url: imageUrl,
						},
					},
				}
				choice.Delta.Content = messageContents
			} else {
				// Only have image content
				choice.Delta.Content = []model.MessageContent{
					{
						Type: model.ContentTypeImageURL,
						ImageURL: &model.ImageURL{
							Url: imageUrl,
						},
					},
				}
			}
		}

		// Handle function calls (if present)
		if part.FunctionCall != nil {
			choice.Delta.ToolCalls = getToolCalls(&candidate)
		}
	}

	// Create response
	var response openai.ChatCompletionsStreamResponse
	response.Id = fmt.Sprintf("chatcmpl-%s", random.GetUUID())
	response.Created = helper.GetTimestamp()
	response.Object = "chat.completion.chunk"
	response.Model = "gemini"
	response.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}

	return &response
}

func embeddingResponseGemini2OpenAI(response *EmbeddingResponse) *openai.EmbeddingResponse {
	openAIEmbeddingResponse := openai.EmbeddingResponse{
		Object: "list",
		Data:   make([]openai.EmbeddingResponseItem, 0, len(response.Embeddings)),
		Model:  "gemini-embedding",
		Usage:  model.Usage{TotalTokens: 0},
	}
	for _, item := range response.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, openai.EmbeddingResponseItem{
			Object:    `embedding`,
			Index:     0,
			Embedding: item.Values,
		})
	}
	return &openAIEmbeddingResponse
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, string) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	buffer := make([]byte, 10*1024*1024) // 10MB buffer
	scanner.Buffer(buffer, len(buffer))

	common.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		data = strings.TrimSpace(data)

		if !strings.HasPrefix(data, "data: ") {
			continue
		}
		data = strings.TrimPrefix(data, "data: ")
		data = strings.TrimSuffix(data, "\"")

		var geminiResponse ChatResponse
		err := json.Unmarshal([]byte(data), &geminiResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}

		response := streamResponseGeminiChat2OpenAI(&geminiResponse)
		if response == nil {
			continue
		}

		responseText += response.Choices[0].Delta.StringContent()

		err = render.ObjectData(c, response)
		if err != nil {
			logger.SysError(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	render.Done(c)

	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}

	return nil, responseText
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var geminiResponse ChatResponse
	err = json.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if len(geminiResponse.Candidates) == 0 {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: "No candidates returned",
				Type:    "server_error",
				Param:   "",
				Code:    500,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseGeminiChat2OpenAI(&geminiResponse)
	fullTextResponse.Model = modelName
	completionTokens := openai.CountTokenText(geminiResponse.GetResponseText(), modelName)
	usage := model.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}

func EmbeddingHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var geminiEmbeddingResponse EmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &geminiEmbeddingResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if geminiEmbeddingResponse.Error != nil {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: geminiEmbeddingResponse.Error.Message,
				Type:    "gemini_error",
				Param:   "",
				Code:    geminiEmbeddingResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := embeddingResponseGemini2OpenAI(&geminiEmbeddingResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
