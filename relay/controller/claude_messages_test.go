package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestGetAndValidateClaudeMessagesRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		requestBody string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			requestBody: `{
				"model": "claude-3-sonnet-20240229",
				"max_tokens": 1024,
				"messages": [
					{
						"role": "user",
						"content": "Hello, how are you?"
					}
				]
			}`,
			expectError: false,
		},
		{
			name: "missing model",
			requestBody: `{
				"max_tokens": 1024,
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					}
				]
			}`,
			expectError: true,
			errorMsg:    "model is required",
		},
		{
			name: "missing max_tokens",
			requestBody: `{
				"model": "claude-3-sonnet-20240229",
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					}
				]
			}`,
			expectError: true,
			errorMsg:    "max_tokens must be greater than 0",
		},
		{
			name: "empty messages",
			requestBody: `{
				"model": "claude-3-sonnet-20240229",
				"max_tokens": 1024,
				"messages": []
			}`,
			expectError: true,
			errorMsg:    "messages array cannot be empty",
		},
		{
			name: "invalid message role",
			requestBody: `{
				"model": "claude-3-sonnet-20240229",
				"max_tokens": 1024,
				"messages": [
					{
						"role": "system",
						"content": "Hello"
					}
				]
			}`,
			expectError: true,
			errorMsg:    "message[0].role must be 'user' or 'assistant'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up request
			req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			// Call the function
			result, err := getAndValidateClaudeMessagesRequest(c)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "claude-3-sonnet-20240229", result.Model)
				assert.Equal(t, 1024, result.MaxTokens)
				assert.Len(t, result.Messages, 1)
			}
		})
	}
}

func TestGetClaudeMessagesPromptTokens(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		request  *ClaudeMessagesRequest
		expected int
	}{
		{
			name: "simple text message",
			request: &ClaudeMessagesRequest{
				Messages: []relaymodel.ClaudeMessage{
					{
						Role:    "user",
						Content: "Hello, how are you today?", // 24 chars -> 6 tokens
					},
				},
			},
			expected: 7, // "user" (4) + "Hello, how are you today?" (24) = 28 chars / 4 = 7 tokens
		},
		{
			name: "multiple messages",
			request: &ClaudeMessagesRequest{
				Messages: []relaymodel.ClaudeMessage{
					{
						Role:    "user",
						Content: "Hello", // 5 chars
					},
					{
						Role:    "assistant",
						Content: "Hi there!", // 9 chars
					},
				},
			},
			expected: 6, // "user" (4) + "Hello" (5) + "assistant" (9) + "Hi there!" (9) = 27 chars / 4 = 6.75 -> 6 tokens
		},
		{
			name: "with system prompt",
			request: &ClaudeMessagesRequest{
				System: "You are a helpful assistant.", // 28 chars
				Messages: []relaymodel.ClaudeMessage{
					{
						Role:    "user",
						Content: "Hello", // 5 chars
					},
				},
			},
			expected: 10, // System: "You are a helpful assistant." (28) + "user" (4) + "Hello" (5) = 37 chars / 4 = 9.25 -> 10 tokens
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getClaudeMessagesPromptTokens(ctx, tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClaudeMessagesRelayMode(t *testing.T) {
	// Test that the relay mode is correctly detected for /v1/messages
	mode := relaymode.GetByPath("/v1/messages")
	assert.Equal(t, relaymode.ClaudeMessages, mode)

	// Test other paths are not affected
	assert.Equal(t, relaymode.ChatCompletions, relaymode.GetByPath("/v1/chat/completions"))
	assert.Equal(t, relaymode.ResponseAPI, relaymode.GetByPath("/v1/responses"))
}

func TestClaudeMessagesRequestStructure(t *testing.T) {
	// Test that the Claude Messages request structure can be properly marshaled/unmarshaled
	originalRequest := &ClaudeMessagesRequest{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 1024,
		Messages: []relaymodel.ClaudeMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
			{
				Role:    "assistant",
				Content: "I'm doing well, thank you!",
			},
		},
		System:      "You are a helpful assistant.",
		Temperature: func() *float64 { f := 0.7; return &f }(),
		Stream:      func() *bool { b := true; return &b }(),
		Tools: []relaymodel.ClaudeTool{
			{
				Name:        "get_weather",
				Description: "Get the current weather",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and state",
						},
					},
				},
			},
		},
		ToolChoice: map[string]any{
			"type": "auto",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(originalRequest)
	require.NoError(t, err)

	// Unmarshal back
	var parsedRequest ClaudeMessagesRequest
	err = json.Unmarshal(jsonData, &parsedRequest)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, originalRequest.Model, parsedRequest.Model)
	assert.Equal(t, originalRequest.MaxTokens, parsedRequest.MaxTokens)
	assert.Len(t, parsedRequest.Messages, 2)
	assert.Equal(t, "user", parsedRequest.Messages[0].Role)
	assert.Equal(t, "assistant", parsedRequest.Messages[1].Role)
	assert.NotNil(t, parsedRequest.Temperature)
	assert.Equal(t, 0.7, *parsedRequest.Temperature)
	assert.NotNil(t, parsedRequest.Stream)
	assert.True(t, *parsedRequest.Stream)
	assert.Len(t, parsedRequest.Tools, 1)
	assert.Equal(t, "get_weather", parsedRequest.Tools[0].Name)
}
