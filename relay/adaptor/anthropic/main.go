package anthropic

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/image"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

func stopReasonClaude2OpenAI(reason *string) string {
	if reason == nil {
		return ""
	}
	switch *reason {
	case "end_turn":
		return "stop"
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	default:
		return *reason
	}
}

// isModelSupportThinking is used to check if the model supports extended thinking
func isModelSupportThinking(model string) bool {
	// Claude 3.7 Sonnet
	if strings.Contains(model, "claude-3-7-sonnet") {
		return true
	}

	// Claude 4 models (Opus 4 and Sonnet 4)
	if strings.Contains(model, "claude-opus-4") || strings.Contains(model, "claude-sonnet-4") {
		return true
	}

	return false
}

func ConvertRequest(c *gin.Context, textRequest model.GeneralOpenAIRequest) (*Request, error) {
	claudeTools := make([]Tool, 0, len(textRequest.Tools))

	for _, tool := range textRequest.Tools {
		claudeTools = append(claudeTools, Tool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: InputSchema{
				Type:       tool.Function.Parameters["type"].(string),
				Properties: tool.Function.Parameters["properties"],
				Required:   tool.Function.Parameters["required"],
			},
		})
	}

	claudeRequest := Request{
		Model:       textRequest.Model,
		MaxTokens:   textRequest.MaxTokens,
		Temperature: textRequest.Temperature,
		TopP:        textRequest.TopP,
		TopK:        textRequest.TopK,
		Stream:      textRequest.Stream,
		Tools:       claudeTools,
		Thinking:    textRequest.Thinking,
	}

	if isModelSupportThinking(textRequest.Model) &&
		c.Request.URL.Query().Has("thinking") && claudeRequest.Thinking == nil {
		claudeRequest.Thinking = &model.Thinking{
			Type:         "enabled",
			BudgetTokens: int(math.Min(1024, float64(claudeRequest.MaxTokens/2))),
		}
	}

	if isModelSupportThinking(textRequest.Model) &&
		claudeRequest.Thinking != nil {
		if claudeRequest.MaxTokens <= 1024 {
			return nil, errors.New("max_tokens must be greater than 1024 when using extended thinking")
		}

		// top_p must be nil when using extended thinking
		claudeRequest.TopP = nil
	}

	if len(claudeTools) > 0 {
		claudeToolChoice := struct {
			Type string `json:"type"`
			Name string `json:"name,omitempty"`
		}{Type: "auto"} // default value https://docs.anthropic.com/en/docs/build-with-claude/tool-use#controlling-claudes-output
		if choice, ok := textRequest.ToolChoice.(map[string]any); ok {
			if function, ok := choice["function"].(map[string]any); ok {
				claudeToolChoice.Type = "tool"
				claudeToolChoice.Name = function["name"].(string)
			}
		} else if toolChoiceType, ok := textRequest.ToolChoice.(string); ok {
			if toolChoiceType == "any" {
				claudeToolChoice.Type = toolChoiceType
			}
		}
		claudeRequest.ToolChoice = claudeToolChoice
	}
	if claudeRequest.MaxTokens == 0 {
		claudeRequest.MaxTokens = config.DefaultMaxToken
	}
	// legacy model name mapping
	if claudeRequest.Model == "claude-instant-1" {
		claudeRequest.Model = "claude-instant-1.1"
	} else if claudeRequest.Model == "claude-2" {
		claudeRequest.Model = "claude-2.1"
	}
	for _, message := range textRequest.Messages {
		if message.Role == "system" && claudeRequest.System == "" {
			claudeRequest.System = message.StringContent()
			continue
		}
		claudeMessage := Message{
			Role: message.Role,
		}
		var content Content
		if message.IsStringContent() {
			stringContent := message.StringContent()

			if message.Role == "tool" {
				claudeMessage.Role = "user"
				content.Type = "tool_result"
				content.Content = stringContent
				content.ToolUseId = message.ToolCallId
				claudeMessage.Content = append(claudeMessage.Content, content)
			} else if stringContent != "" {
				// For assistant messages with thinking enabled, check if we need to add thinking block
				if message.Role == "assistant" && claudeRequest.Thinking != nil {
					// Check if this message has reasoning content that should be converted to thinking block
					var reasoningContent string
					if message.Reasoning != nil {
						reasoningContent = *message.Reasoning
					} else if message.ReasoningContent != nil {
						reasoningContent = *message.ReasoningContent
					} else if message.Thinking != nil {
						reasoningContent = *message.Thinking
					}

					// If we have reasoning content, add it as a thinking block first
					if reasoningContent != "" {
						thinkingContent := Content{
							Type:     "thinking",
							Thinking: &reasoningContent,
						}
						claudeMessage.Content = append(claudeMessage.Content, thinkingContent)
					}
				}

				// Only add text content if it's not empty
				content.Type = "text"
				content.Text = stringContent
				claudeMessage.Content = append(claudeMessage.Content, content)
			}

			// Add tool calls
			for i := range message.ToolCalls {
				inputParam := make(map[string]any)
				if err := json.Unmarshal([]byte(message.ToolCalls[i].Function.Arguments.(string)), &inputParam); err != nil {
					return nil, errors.Wrapf(err, "unmarshal tool call arguments for tool %s", message.ToolCalls[i].Function.Name)
				}
				claudeMessage.Content = append(claudeMessage.Content, Content{
					Type:  "tool_use",
					Id:    message.ToolCalls[i].Id,
					Name:  message.ToolCalls[i].Function.Name,
					Input: inputParam,
				})
			}

			// Claude requires at least one content block per message
			if len(claudeMessage.Content) == 0 {
				return nil, errors.Wrap(errors.New("message must have at least one content block"), "validate message content")
			}

			claudeRequest.Messages = append(claudeRequest.Messages, claudeMessage)
			continue
		}
		var contents []Content

		// For assistant messages with thinking enabled, check if we need to add thinking block first
		if message.Role == "assistant" && claudeRequest.Thinking != nil {
			// Check if this message has reasoning content that should be converted to thinking block
			var reasoningContent string
			if message.Reasoning != nil {
				reasoningContent = *message.Reasoning
			} else if message.ReasoningContent != nil {
				reasoningContent = *message.ReasoningContent
			} else if message.Thinking != nil {
				reasoningContent = *message.Thinking
			}

			// If we have reasoning content, add it as a thinking block first
			if reasoningContent != "" {
				thinkingContent := Content{
					Type:     "thinking",
					Thinking: &reasoningContent,
				}
				contents = append(contents, thinkingContent)
			}
		}

		openaiContent := message.ParseContent()
		for _, part := range openaiContent {
			var content Content
			if part.Type == model.ContentTypeText {
				content.Type = "text"
				if part.Text != nil && *part.Text != "" {
					// Only add text content if it's not empty
					content.Text = *part.Text
					contents = append(contents, content)
				}
			} else if part.Type == model.ContentTypeImageURL {
				content.Type = "image"
				content.Source = &ImageSource{
					Type: "base64",
				}
				mimeType, data, _ := image.GetImageFromUrl(part.ImageURL.Url)
				content.Source.MediaType = mimeType
				content.Source.Data = data
				contents = append(contents, content)
			}
		}

		// Add tool calls for non-string content messages
		for i := range message.ToolCalls {
			inputParam := make(map[string]any)
			if err := json.Unmarshal([]byte(message.ToolCalls[i].Function.Arguments.(string)), &inputParam); err != nil {
				return nil, errors.Wrapf(err, "unmarshal tool call arguments for tool %s", message.ToolCalls[i].Function.Name)
			}
			contents = append(contents, Content{
				Type:  "tool_use",
				Id:    message.ToolCalls[i].Id,
				Name:  message.ToolCalls[i].Function.Name,
				Input: inputParam,
			})
		}
		claudeMessage.Content = contents

		// Claude requires at least one content block per message
		if len(claudeMessage.Content) == 0 {
			return nil, errors.Wrap(errors.New("message must have at least one content block"), "validate message content")
		}

		claudeRequest.Messages = append(claudeRequest.Messages, claudeMessage)
	}
	return &claudeRequest, nil
}

// https://docs.anthropic.com/claude/reference/messages-streaming
func StreamResponseClaude2OpenAI(c *gin.Context, claudeResponse *StreamResponse) (*openai.ChatCompletionsStreamResponse, *Response) {
	var response *Response
	var responseText string
	var reasoningText string
	var stopReason string
	tools := make([]model.Tool, 0)

	switch claudeResponse.Type {
	case "message_start":
		return nil, claudeResponse.Message
	case "content_block_start":
		if claudeResponse.ContentBlock != nil {
			responseText = claudeResponse.ContentBlock.Text
			if claudeResponse.ContentBlock.Thinking != nil {
				reasoningText = *claudeResponse.ContentBlock.Thinking
			}

			if claudeResponse.ContentBlock.Type == "tool_use" {
				// Set index for streaming tool calls - use the current index in the tools slice
				index := len(tools)
				tools = append(tools, model.Tool{
					Id:   claudeResponse.ContentBlock.Id,
					Type: "function",
					Function: model.Function{
						Name:      claudeResponse.ContentBlock.Name,
						Arguments: "",
					},
					Index: &index, // Set index for streaming delta accumulation
				})
			}
		}
	case "content_block_delta":
		if claudeResponse.Delta != nil {
			responseText = claudeResponse.Delta.Text
			if claudeResponse.Delta.Thinking != nil {
				reasoningText = *claudeResponse.Delta.Thinking
			}

			if claudeResponse.Delta.Type == "input_json_delta" {
				// For input_json_delta, we should update the last tool call's arguments, not create a new one
				// The index should match the last tool call that was started in content_block_start
				if len(tools) > 0 {
					// Update the last tool call's arguments (this is a delta for the existing tool call)
					lastIndex := len(tools) - 1
					lastTool := tools[lastIndex]
					if existingArgs, ok := lastTool.Function.Arguments.(string); ok {
						lastTool.Function.Arguments = existingArgs + claudeResponse.Delta.PartialJson
					} else {
						lastTool.Function.Arguments = claudeResponse.Delta.PartialJson
					}
					// Keep the same index as the original tool call
					tools[lastIndex] = lastTool
				} else {
					// Fallback: create new tool call if no existing tool call found
					index := 0
					tools = append(tools, model.Tool{
						Function: model.Function{
							Arguments: claudeResponse.Delta.PartialJson,
						},
						Index: &index, // Set index for streaming delta accumulation
					})
				}
			}
		}
	case "message_delta":
		if claudeResponse.Usage != nil {
			response = &Response{
				Usage: *claudeResponse.Usage,
			}
		}
		if claudeResponse.Delta != nil && claudeResponse.Delta.StopReason != nil {
			stopReason = *claudeResponse.Delta.StopReason
		}
	case "thinking_delta":
		if claudeResponse.Delta != nil && claudeResponse.Delta.Thinking != nil {
			reasoningText = *claudeResponse.Delta.Thinking
		}
	case "ping",
		"message_stop",
		"content_block_stop":
	default:
		logger.SysErrorf("unknown stream response type %q", claudeResponse.Type)
	}

	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = responseText
	choice.Delta.SetReasoningContent(c.Query("reasoning_format"), reasoningText)
	if len(tools) > 0 {
		choice.Delta.Content = nil // compatible with other OpenAI derivative applications, like LobeOpenAICompatibleFactory ...
		choice.Delta.ToolCalls = tools
	}
	choice.Delta.Role = "assistant"
	finishReason := stopReasonClaude2OpenAI(&stopReason)
	if finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	var openaiResponse openai.ChatCompletionsStreamResponse
	openaiResponse.Object = "chat.completion.chunk"
	openaiResponse.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
	return &openaiResponse, response
}

func ResponseClaude2OpenAI(c *gin.Context, claudeResponse *Response) *openai.TextResponse {
	var responseText string
	var reasoningText string

	tools := make([]model.Tool, 0)
	for _, v := range claudeResponse.Content {
		switch v.Type {
		case "thinking":
			if v.Thinking != nil {
				reasoningText += *v.Thinking
			} else {
				logger.Errorf(context.Background(), "thinking is nil in response")
			}
		case "text":
			responseText += v.Text
		default:
			logger.Warnf(context.Background(), "unknown response type %q", v.Type)
		}

		if v.Type == "tool_use" {
			args, _ := json.Marshal(v.Input)
			tools = append(tools, model.Tool{
				Id:   v.Id,
				Type: "function", // compatible with other OpenAI derivative applications
				Function: model.Function{
					Name:      v.Name,
					Arguments: string(args),
				},
			})
		}
	}

	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:      "assistant",
			Content:   responseText,
			Reasoning: &reasoningText,
			Name:      nil,
			ToolCalls: tools,
		},
		FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
	}
	choice.Message.SetReasoningContent(c.Query("reasoning_format"), reasoningText)
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", claudeResponse.Id),
		Model:   claudeResponse.Model,
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	createdTime := helper.GetTimestamp()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	common.SetEventStreamHeaders(c)

	var usage model.Usage
	var modelName string
	var id string
	var lastToolCallChoice openai.ChatCompletionsStreamResponseChoice

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 6 || !strings.HasPrefix(data, "data:") {
			continue
		}
		data = strings.TrimPrefix(data, "data:")
		data = strings.TrimSpace(data)

		logger.Debugf(c.Request.Context(), "stream <- %q\n", data)

		var claudeResponse StreamResponse
		err := json.Unmarshal([]byte(data), &claudeResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}

		response, meta := StreamResponseClaude2OpenAI(c, &claudeResponse)
		if meta != nil {
			usage.PromptTokens += meta.Usage.InputTokens
			usage.CompletionTokens += meta.Usage.OutputTokens
			if len(meta.Id) > 0 { // only message_start has an id, otherwise it's a finish_reason event.
				modelName = meta.Model
				id = fmt.Sprintf("chatcmpl-%s", meta.Id)
				continue
			} else { // finish_reason case
				if len(lastToolCallChoice.Delta.ToolCalls) > 0 {
					lastArgs := &lastToolCallChoice.Delta.ToolCalls[len(lastToolCallChoice.Delta.ToolCalls)-1].Function
					if len(lastArgs.Arguments.(string)) == 0 { // compatible with OpenAI sending an empty object `{}` when no arguments.
						lastArgs.Arguments = "{}"
						response.Choices[len(response.Choices)-1].Delta.Content = nil
						response.Choices[len(response.Choices)-1].Delta.ToolCalls = lastToolCallChoice.Delta.ToolCalls
					}
				}
				// Add usage information to the final chunk
				if usage.PromptTokens > 0 || usage.CompletionTokens > 0 {
					usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
					if response != nil {
						response.Usage = &usage
					}
				}
			}
		}
		if response == nil {
			continue
		}

		response.Id = id
		response.Model = modelName
		response.Created = createdTime

		for _, choice := range response.Choices {
			if len(choice.Delta.ToolCalls) > 0 {
				lastToolCallChoice = choice
			}
		}
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
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &usage
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

	logger.Debugf(c.Request.Context(), "response <- %s\n", string(responseBody))

	var claudeResponse Response
	err = json.Unmarshal(responseBody, &claudeResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if claudeResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: claudeResponse.Error.Message,
				Type:    claudeResponse.Error.Type,
				Param:   "",
				Code:    claudeResponse.Error.Type,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := ResponseClaude2OpenAI(c, &claudeResponse)
	fullTextResponse.Model = modelName
	usage := model.Usage{
		PromptTokens:     claudeResponse.Usage.InputTokens,
		CompletionTokens: claudeResponse.Usage.OutputTokens,
		TotalTokens:      claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
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
