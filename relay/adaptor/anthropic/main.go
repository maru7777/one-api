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
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
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
	if strings.Contains(model, "claude-3-5") ||
		strings.Contains(model, "claude-2") ||
		strings.Contains(model, "claude-instant-1") {
		return false
	}

	return true
}

// ConvertClaudeRequest converts a Claude Messages API request to anthropic.Request format
func ConvertClaudeRequest(c *gin.Context, claudeRequest model.ClaudeRequest) (*Request, error) {
	// Convert tools
	claudeTools := make([]Tool, 0, len(claudeRequest.Tools))
	for _, tool := range claudeRequest.Tools {
		// Convert InputSchema from any to InputSchema struct
		var inputSchema InputSchema
		if tool.InputSchema != nil {
			if schemaMap, ok := tool.InputSchema.(map[string]any); ok {
				if schemaType, exists := schemaMap["type"]; exists {
					if typeStr, ok := schemaType.(string); ok {
						inputSchema.Type = typeStr
					}
				}
				if properties, exists := schemaMap["properties"]; exists {
					inputSchema.Properties = properties
				}
				if required, exists := schemaMap["required"]; exists {
					inputSchema.Required = required
				}
			}
		}

		claudeTools = append(claudeTools, Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: inputSchema,
		})
	}

	// Convert messages
	claudeMessages := make([]Message, 0, len(claudeRequest.Messages))
	for _, msg := range claudeRequest.Messages {
		claudeMessage := Message{
			Role: msg.Role,
		}

		// Convert content based on type
		switch content := msg.Content.(type) {
		case string:
			// Simple string content
			claudeMessage.Content = []Content{
				{
					Type: "text",
					Text: content,
				},
			}
		case []any:
			// Structured content blocks
			for _, block := range content {
				if blockMap, ok := block.(map[string]any); ok {
					contentBlock := Content{}
					if blockType, exists := blockMap["type"]; exists {
						if typeStr, ok := blockType.(string); ok {
							contentBlock.Type = typeStr
						}
					}
					if text, exists := blockMap["text"]; exists {
						if textStr, ok := text.(string); ok {
							contentBlock.Text = textStr
						}
					}
					// Handle image content
					if source, exists := blockMap["source"]; exists {
						if sourceMap, ok := source.(map[string]any); ok {
							contentBlock.Source = &ImageSource{}
							if sourceType, exists := sourceMap["type"]; exists {
								if typeStr, ok := sourceType.(string); ok {
									contentBlock.Source.Type = typeStr
								}
							}
							if mediaType, exists := sourceMap["media_type"]; exists {
								if mediaTypeStr, ok := mediaType.(string); ok {
									contentBlock.Source.MediaType = mediaTypeStr
								}
							}
							if data, exists := sourceMap["data"]; exists {
								if dataStr, ok := data.(string); ok {
									contentBlock.Source.Data = dataStr
								}
							}
						}
					}
					claudeMessage.Content = append(claudeMessage.Content, contentBlock)
				}
			}
		default:
			// Try to marshal and unmarshal as Content slice
			contentBytes, err := json.Marshal(content)
			if err != nil {
				return nil, errors.Wrap(err, "failed to marshal message content")
			}
			var contentBlocks []Content
			if err := json.Unmarshal(contentBytes, &contentBlocks); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal message content")
			}
			claudeMessage.Content = contentBlocks
		}

		claudeMessages = append(claudeMessages, claudeMessage)
	}

	// Convert system prompt
	var systemPrompt string
	if claudeRequest.System != nil {
		switch system := claudeRequest.System.(type) {
		case string:
			systemPrompt = system
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
			systemPrompt = strings.Join(systemParts, "\n")
		}
	}

	// Build the request
	request := &Request{
		Model:         claudeRequest.Model,
		MaxTokens:     claudeRequest.MaxTokens,
		Messages:      claudeMessages,
		System:        systemPrompt,
		Temperature:   claudeRequest.Temperature,
		TopP:          claudeRequest.TopP,
		Stream:        claudeRequest.Stream != nil && *claudeRequest.Stream,
		StopSequences: claudeRequest.StopSequences,
		Tools:         claudeTools,
		ToolChoice:    claudeRequest.ToolChoice,
		Thinking:      claudeRequest.Thinking,
	}

	// Handle TopK (convert from *int to int)
	if claudeRequest.TopK != nil {
		request.TopK = *claudeRequest.TopK
	}

	return request, nil
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

	// Track if we need to use fallback mode (will be set if any signature restoration fails)
	var useFallbackMode bool

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

					// If we have reasoning content, handle it appropriately
					if reasoningContent != "" {
						var signatureRestored bool
						thinkingContent := Content{
							Type:     "thinking",
							Thinking: &reasoningContent,
						}

						// Try to restore signature from cache if available
						if tokenID, exists := c.Get(ctxkey.TokenId); exists {
							if tokenIDInt, ok := tokenID.(int); ok {
								tokenIDStr := getTokenIDFromRequest(tokenIDInt)
								conversationID := generateConversationID(textRequest.Messages)

								// Store conversation ID in context for later use
								c.Set(ctxkey.ConversationId, conversationID)

								// Try to restore signature for this thinking block
								messageIndex := len(claudeRequest.Messages) // Current message index
								cacheKey := generateSignatureKey(tokenIDStr, conversationID, messageIndex, 0)

								if signature := GetSignatureCache().Get(cacheKey); signature != nil {
									thinkingContent.Signature = signature
									signatureRestored = true
								}
							}
						}

						// If signature was not restored, use fallback approach
						if !signatureRestored {
							// Set fallback mode flag
							useFallbackMode = true

							// Convert thinking content to <think> format and prepend to text content
							thinkingPrefix := fmt.Sprintf("<think>%s</think>\n\n", reasoningContent)

							// Find the first text content and prepend thinking
							for i := range claudeMessage.Content {
								if claudeMessage.Content[i].Type == "text" {
									claudeMessage.Content[i].Text = thinkingPrefix + claudeMessage.Content[i].Text
									break
								}
							}

							// If no text content found, create one with thinking prefix and original text
							if len(claudeMessage.Content) == 0 {
								claudeMessage.Content = append(claudeMessage.Content, Content{
									Type: "text",
									Text: thinkingPrefix + stringContent,
								})
							}
						} else {
							// Signature was restored, use proper thinking block
							claudeMessage.Content = append([]Content{thinkingContent}, claudeMessage.Content...)
						}
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

		// Store reasoning content for later processing
		var reasoningContent string
		var needsThinkingProcessing bool
		if message.Role == "assistant" && claudeRequest.Thinking != nil {
			// Check if this message has reasoning content that should be converted to thinking block
			if message.Reasoning != nil {
				reasoningContent = *message.Reasoning
				needsThinkingProcessing = true
			} else if message.ReasoningContent != nil {
				reasoningContent = *message.ReasoningContent
				needsThinkingProcessing = true
			} else if message.Thinking != nil {
				reasoningContent = *message.Thinking
				needsThinkingProcessing = true
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

		// Process thinking content after content parsing
		if needsThinkingProcessing && reasoningContent != "" {
			var signatureRestored bool
			thinkingContent := Content{
				Type:     "thinking",
				Thinking: &reasoningContent,
			}

			// Try to restore signature from cache if available
			if tokenID, exists := c.Get(ctxkey.TokenId); exists {
				if tokenIDInt, ok := tokenID.(int); ok {
					tokenIDStr := getTokenIDFromRequest(tokenIDInt)
					conversationID := generateConversationID(textRequest.Messages)

					// Store conversation ID in context for later use
					c.Set(ctxkey.ConversationId, conversationID)

					// Try to restore signature for this thinking block
					messageIndex := len(claudeRequest.Messages) // Current message index
					cacheKey := generateSignatureKey(tokenIDStr, conversationID, messageIndex, 0)

					if signature := GetSignatureCache().Get(cacheKey); signature != nil {
						thinkingContent.Signature = signature
						signatureRestored = true
					}
				}
			}

			// If signature was not restored, use fallback approach
			if !signatureRestored {
				// Set fallback mode flag
				useFallbackMode = true

				// Convert thinking content to <think> format and prepend to text content
				thinkingPrefix := fmt.Sprintf("<think>%s</think>\n\n", reasoningContent)

				// Get the original response text from the message
				originalText := ""
				if message.IsStringContent() {
					originalText = message.StringContent()
				} else {
					// Try to get text from Content directly
					if str, ok := message.Content.(string); ok {
						originalText = str
					}
				}

				// Find the first text content and prepend thinking
				foundTextContent := false
				for i := range contents {
					if contents[i].Type == "text" {
						contents[i].Text = thinkingPrefix + contents[i].Text
						foundTextContent = true
						break
					}
				}

				// If no text content found, create one with thinking prefix and original text
				if !foundTextContent {
					contents = append(contents, Content{
						Type: "text",
						Text: thinkingPrefix + originalText,
					})
				}
			} else {
				// Signature was restored, use proper thinking block
				contents = append([]Content{thinkingContent}, contents...)
			}
		}

		claudeMessage.Content = contents

		// Claude requires at least one content block per message
		if len(claudeMessage.Content) == 0 {
			return nil, errors.Wrap(errors.New("message must have at least one content block"), "validate message content")
		}

		claudeRequest.Messages = append(claudeRequest.Messages, claudeMessage)
	}

	// If fallback mode was used, disable thinking to avoid Claude validation errors
	if useFallbackMode && claudeRequest.Thinking != nil {
		claudeRequest.Thinking = nil
	}

	return &claudeRequest, nil
}

// https://docs.anthropic.com/claude/reference/messages-streaming
func StreamResponseClaude2OpenAI(c *gin.Context, claudeResponse *StreamResponse) (*openai.ChatCompletionsStreamResponse, *Response) {
	var response *Response
	var responseText string
	var reasoningText string
	var signatureText string
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
			if claudeResponse.ContentBlock.Signature != nil {
				signatureText = *claudeResponse.ContentBlock.Signature
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
			if claudeResponse.Delta.Type == "signature_delta" && claudeResponse.Delta.Signature != nil {
				signatureText = *claudeResponse.Delta.Signature
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
	case "signature_delta":
		if claudeResponse.Delta != nil && claudeResponse.Delta.Signature != nil {
			signatureText = *claudeResponse.Delta.Signature
		}
	case "ping",
		"message_stop",
		"content_block_stop":
	default:
		logger.SysErrorf("unknown stream response type %q", claudeResponse.Type)
	}

	// Cache signature if present (for thinking blocks)
	if signatureText != "" && (reasoningText != "" || claudeResponse.Type == "signature_delta") {
		// Get token ID from context
		if tokenID, exists := c.Get(ctxkey.TokenId); exists {
			if tokenIDInt, ok := tokenID.(int); ok {
				// We need the original request to generate conversation ID
				// For now, we'll cache with a temporary key and update it later
				// This will be properly handled in the request conversion phase
				tokenIDStr := getTokenIDFromRequest(tokenIDInt)
				tempKey := fmt.Sprintf("temp_sig:%s:%d", tokenIDStr, time.Now().UnixNano())
				GetSignatureCache().Store(tempKey, signatureText)

				// Store the temp key in context for later use
				c.Set(ctxkey.TempSignatureKey, tempKey)
			}
		}
	}

	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = responseText
	if reasoningText != "" {
		choice.Delta.SetReasoningContent(c.Query("reasoning_format"), reasoningText)
	}
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
	for i, v := range claudeResponse.Content {
		switch v.Type {
		case "thinking", "redacted_thinking":
			if v.Thinking != nil {
				reasoningText += *v.Thinking
			} else {
				logger.Errorf(context.Background(), "thinking is nil in response")
			}
			// Cache signature if present
			if v.Signature != nil {
				// Cache the signature for future use
				if tokenID, exists := c.Get(ctxkey.TokenId); exists {
					if tokenIDInt, ok := tokenID.(int); ok {
						// Get conversation ID from request context or generate it
						var conversationID string
						if convID, exists := c.Get(ctxkey.ConversationId); exists {
							conversationID = convID.(string)
						} else {
							// We'll need the original request messages to generate conversation ID
							// For now, use a temporary approach
							conversationID = fmt.Sprintf("temp_conv_%d", time.Now().UnixNano())
						}

						tokenIDStr := getTokenIDFromRequest(tokenIDInt)
						cacheKey := generateSignatureKey(tokenIDStr, conversationID, 0, i) // messageIndex=0 for response
						GetSignatureCache().Store(cacheKey, *v.Signature)
					}
				}
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
	if reasoningText != "" {
		choice.Message.SetReasoningContent(c.Query("reasoning_format"), reasoningText)
	}
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", claudeResponse.Id),
		Model:   claudeResponse.Model,
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

// ClaudeNativeStreamHandler handles streaming responses in Claude native format without conversion to OpenAI format
func ClaudeNativeStreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
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

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 6 || !strings.HasPrefix(data, "data:") {
			continue
		}
		data = strings.TrimPrefix(data, "data:")
		data = strings.TrimSpace(data)

		logger.Debugf(c.Request.Context(), "stream <- %q\n", data)

		// For Claude native streaming, we pass through the events directly
		c.Writer.Write([]byte("data: " + data + "\n\n"))
		c.Writer.(http.Flusher).Flush()

		// Parse the response to extract usage and model info
		var claudeResponse StreamResponse
		err := json.Unmarshal([]byte(data), &claudeResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}

		// Extract usage info from message_delta
		if claudeResponse.Type == "message_delta" && claudeResponse.Usage != nil {
			usage.PromptTokens += claudeResponse.Usage.InputTokens
			usage.CompletionTokens += claudeResponse.Usage.OutputTokens
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	// Send final data: [DONE] to close the stream
	c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.(http.Flusher).Flush()

	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &usage
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

// ClaudeNativeHandler handles responses in Claude native format without conversion to OpenAI format
func ClaudeNativeHandler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
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

	// Return response in Claude's native format
	claudeResponse.Model = modelName
	usage := model.Usage{
		PromptTokens:     claudeResponse.Usage.InputTokens,
		CompletionTokens: claudeResponse.Usage.OutputTokens,
		TotalTokens:      claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
	}

	jsonResponse, err := json.Marshal(claudeResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
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
