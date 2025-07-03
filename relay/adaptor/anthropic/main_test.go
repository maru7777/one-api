package anthropic

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/songquanpeng/one-api/relay/model"
)

func TestConvertRequest_EmptyAssistantMessageWithToolCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Test case: assistant message with empty content but with tool calls
	temp := 0.7
	openaiRequest := model.GeneralOpenAIRequest{
		Model:       "claude-3.5-sonnet",
		MaxTokens:   4096,
		Temperature: &temp,
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "What's the weather like?",
			},
			{
				Role:    "assistant",
				Content: "", // Empty content
				ToolCalls: []model.Tool{
					{
						Id:   "call_123",
						Type: "function",
						Function: model.Function{
							Name:      "get_weather",
							Arguments: `{"location": "San Francisco"}`,
						},
					},
				},
			},
			{
				Role:       "tool",
				Content:    "It's sunny and 72°F",
				ToolCallId: "call_123",
			},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: model.Function{
					Name:        "get_weather",
					Description: "Get weather information",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The location",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		},
	}

	claudeRequest, err := ConvertRequest(c, openaiRequest)
	require.NoError(t, err)

	// Find the assistant message (should be the one with tool_use)
	var assistantMessage *Message
	var toolMessage *Message

	for i := range claudeRequest.Messages {
		if claudeRequest.Messages[i].Role == "assistant" && len(claudeRequest.Messages[i].Content) > 0 {
			for _, content := range claudeRequest.Messages[i].Content {
				if content.Type == "tool_use" {
					assistantMessage = &claudeRequest.Messages[i]
					break
				}
			}
		}
		if claudeRequest.Messages[i].Role == "user" && len(claudeRequest.Messages[i].Content) > 0 {
			for _, content := range claudeRequest.Messages[i].Content {
				if content.Type == "tool_result" {
					toolMessage = &claudeRequest.Messages[i]
					break
				}
			}
		}
	}

	require.NotNil(t, assistantMessage, "Should have assistant message with tool_use")

	// Find the tool_use content in the assistant message
	var toolUseContent *Content
	for i := range assistantMessage.Content {
		if assistantMessage.Content[i].Type == "tool_use" {
			toolUseContent = &assistantMessage.Content[i]
			break
		}
	}

	require.NotNil(t, toolUseContent, "Should have tool_use content")

	// Verify it's a tool_use content
	assert.Equal(t, "tool_use", toolUseContent.Type)
	assert.Equal(t, "call_123", toolUseContent.Id)
	assert.Equal(t, "get_weather", toolUseContent.Name)

	// Verify there's no empty text content in the assistant message
	for _, content := range assistantMessage.Content {
		if content.Type == "text" {
			assert.NotEmpty(t, content.Text, "Text content should not be empty")
		}
	}

	// Verify tool result is properly formatted
	require.NotNil(t, toolMessage, "Should have tool result message")
	require.Len(t, toolMessage.Content, 1)
	assert.Equal(t, "tool_result", toolMessage.Content[0].Type)
	assert.Equal(t, "It's sunny and 72°F", toolMessage.Content[0].Content)
	assert.Equal(t, "call_123", toolMessage.Content[0].ToolUseId)
}

func TestConvertRequest_AssistantMessageWithTextAndToolCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Test case: assistant message with both text content and tool calls
	openaiRequest := model.GeneralOpenAIRequest{
		Model:     "claude-3.5-sonnet",
		MaxTokens: 4096,
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "What's the weather like?",
			},
			{
				Role:    "assistant",
				Content: "I'll check the weather for you.",
				ToolCalls: []model.Tool{
					{
						Id:   "call_123",
						Type: "function",
						Function: model.Function{
							Name:      "get_weather",
							Arguments: `{"location": "San Francisco"}`,
						},
					},
				},
			},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: model.Function{
					Name:        "get_weather",
					Description: "Get weather information",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The location",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		},
	}

	claudeRequest, err := ConvertRequest(c, openaiRequest)
	require.NoError(t, err)

	// Find the assistant message
	var assistantMessage *Message
	for i := range claudeRequest.Messages {
		if claudeRequest.Messages[i].Role == "assistant" {
			assistantMessage = &claudeRequest.Messages[i]
			break
		}
	}

	require.NotNil(t, assistantMessage, "Should have assistant message")
	require.Len(t, assistantMessage.Content, 2) // Should have both text and tool_use

	// Verify text content
	assert.Equal(t, "text", assistantMessage.Content[0].Type)
	assert.Equal(t, "I'll check the weather for you.", assistantMessage.Content[0].Text)

	// Verify tool_use content
	assert.Equal(t, "tool_use", assistantMessage.Content[1].Type)
	assert.Equal(t, "call_123", assistantMessage.Content[1].Id)
	assert.Equal(t, "get_weather", assistantMessage.Content[1].Name)
}

func TestConvertRequest_EmptyTextContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Test case: message with empty text content (using ParseContent path)
	// This should result in an error since the message has no valid content
	openaiRequest := model.GeneralOpenAIRequest{
		Model:     "claude-3.5-sonnet",
		MaxTokens: 4096,
		Messages: []model.Message{
			{
				Role: "user",
				Content: []map[string]interface{}{
					{
						"type": "text",
						"text": "", // Empty text
					},
				},
			},
		},
	}

	_, err := ConvertRequest(c, openaiRequest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message must have at least one content block")
}

func TestConvertRequest_ValidJSONGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Test the original failing case from the bug report
	openaiRequest := model.GeneralOpenAIRequest{
		Model:     "claude-3.5-sonnet",
		MaxTokens: 4096,
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "[User Context: Current time is 2025-06-17 10:22]\n\nUser Request: 今天早上要刷牙",
			},
			{
				Role:    "assistant",
				Content: "", // Empty content
				ToolCalls: []model.Tool{
					{
						Id:   "initial_datetime_call",
						Type: "function",
						Function: model.Function{
							Name:      "get_current_datetime",
							Arguments: "{}",
						},
					},
				},
			},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: model.Function{
					Name:        "get_current_datetime",
					Description: "Get current date and time",
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		},
	}

	claudeRequest, err := ConvertRequest(c, openaiRequest)
	require.NoError(t, err)

	// Verify the JSON can be marshaled without issues
	jsonData, err := json.Marshal(claudeRequest)
	require.NoError(t, err)

	// Find the assistant message
	var assistantMessage *Message
	for i := range claudeRequest.Messages {
		if claudeRequest.Messages[i].Role == "assistant" {
			assistantMessage = &claudeRequest.Messages[i]
			break
		}
	}

	require.NotNil(t, assistantMessage, "Should have assistant message")
	require.Len(t, assistantMessage.Content, 1) // Should only have tool_use

	// Verify it's a tool_use content
	assert.Equal(t, "tool_use", assistantMessage.Content[0].Type)
	assert.Equal(t, "initial_datetime_call", assistantMessage.Content[0].Id)
	assert.Equal(t, "get_current_datetime", assistantMessage.Content[0].Name)

	// Verify the JSON doesn't contain empty text fields
	var jsonObj map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonObj)
	require.NoError(t, err)

	// Check that no message content has empty text type
	messages := jsonObj["messages"].([]interface{})
	for _, msg := range messages {
		msgMap := msg.(map[string]interface{})
		if content, ok := msgMap["content"].([]interface{}); ok {
			for _, c := range content {
				contentMap := c.(map[string]interface{})
				if contentMap["type"] == "text" {
					// If it's a text type, it must have non-empty text
					text, exists := contentMap["text"]
					assert.True(t, exists, "text field must exist for text type")
					assert.NotEmpty(t, text, "text field must not be empty for text type")
				}
			}
		}
	}
}

func TestConvertRequest_MessageWithNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// Test case: message with no content at all
	openaiRequest := model.GeneralOpenAIRequest{
		Model:     "claude-3.5-sonnet",
		MaxTokens: 4096,
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
			{
				Role:    "assistant",
				Content: "", // Empty content and no tool calls
			},
		},
	}

	_, err := ConvertRequest(c, openaiRequest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message must have at least one content block")
}
