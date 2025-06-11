package openai

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

func TestConvertChatCompletionToResponseAPI(t *testing.T) {
	// Test basic conversion
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "Hello, world!"},
		},
		MaxTokens:   100,
		Temperature: floatPtr(0.7),
		TopP:        floatPtr(0.9),
		Stream:      true,
		User:        "test-user",
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify basic fields
	if responseAPI.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", responseAPI.Model)
	}

	if *responseAPI.MaxOutputTokens != 100 {
		t.Errorf("Expected max_output_tokens 100, got %d", *responseAPI.MaxOutputTokens)
	}

	if *responseAPI.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", *responseAPI.Temperature)
	}

	if *responseAPI.TopP != 0.9 {
		t.Errorf("Expected top_p 0.9, got %f", *responseAPI.TopP)
	}

	if !*responseAPI.Stream {
		t.Error("Expected stream to be true")
	}

	if *responseAPI.User != "test-user" {
		t.Errorf("Expected user 'test-user', got '%s'", *responseAPI.User)
	}

	// Verify input conversion
	if len(responseAPI.Input) != 1 {
		t.Errorf("Expected 1 input item, got %d", len(responseAPI.Input))
	}

	inputMessage, ok := responseAPI.Input[0].(model.Message)
	if !ok {
		t.Error("Expected input item to be model.Message type")
	}

	if inputMessage.Role != "user" {
		t.Errorf("Expected message role 'user', got '%s'", inputMessage.Role)
	}

	if inputMessage.StringContent() != "Hello, world!" {
		t.Errorf("Expected message content 'Hello, world!', got '%s'", inputMessage.StringContent())
	}
}

func TestConvertWithSystemMessage(t *testing.T) {
	// Test system message conversion to instructions
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello!"},
		},
		MaxTokens: 50,
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify system message is converted to instructions
	if responseAPI.Instructions == nil {
		t.Error("Expected instructions to be set")
	} else if *responseAPI.Instructions != "You are a helpful assistant." {
		t.Errorf("Expected instructions 'You are a helpful assistant.', got '%s'", *responseAPI.Instructions)
	}

	// Verify system message is removed from input
	if len(responseAPI.Input) != 1 {
		t.Errorf("Expected 1 input item after system message removal, got %d", len(responseAPI.Input))
	}

	inputMessage, ok := responseAPI.Input[0].(model.Message)
	if !ok {
		t.Error("Expected input item to be model.Message type")
	}

	if inputMessage.Role != "user" {
		t.Errorf("Expected remaining message to be user role, got '%s'", inputMessage.Role)
	}
}

func TestConvertWithTools(t *testing.T) {
	// Test tools conversion
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "What's the weather?"},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: model.Function{
					Name:        "get_weather",
					Description: "Get current weather",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify tools are preserved
	if len(responseAPI.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(responseAPI.Tools))
	}

	if responseAPI.Tools[0].Name != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got '%s'", responseAPI.Tools[0].Name)
	}

	if responseAPI.ToolChoice != "auto" {
		t.Errorf("Expected tool_choice 'auto', got '%v'", responseAPI.ToolChoice)
	}
}

func TestConvertWithResponseFormat(t *testing.T) {
	// Test response format conversion
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "Generate JSON"},
		},
		ResponseFormat: &model.ResponseFormat{
			Type: "json_object",
			JsonSchema: &model.JSONSchema{
				Name:        "response_schema",
				Description: "Test schema",
				Schema: map[string]interface{}{
					"type": "object",
				},
			},
		},
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify response format conversion
	if responseAPI.Text == nil {
		t.Error("Expected text config to be set")
	}

	if responseAPI.Text.Type == nil || *responseAPI.Text.Type != "json_object" {
		t.Error("Expected text type to be 'json_object'")
	}

	if responseAPI.Text.JSONSchema == nil {
		t.Error("Expected JSON schema to be set")
	}

	if responseAPI.Text.JSONSchema.Name != "response_schema" {
		t.Errorf("Expected schema name 'response_schema', got '%s'", responseAPI.Text.JSONSchema.Name)
	}
}

// Helper function to create float64 pointers
func floatPtr(f float64) *float64 {
	return &f
}

// TestConvertResponseAPIToChatCompletion tests the conversion from Response API format back to ChatCompletion format
func TestConvertResponseAPIToChatCompletion(t *testing.T) {
	// Create a Response API response
	responseAPI := &ResponseAPIResponse{
		Id:        "resp_123",
		Object:    "response",
		CreatedAt: 1234567890,
		Status:    "completed",
		Model:     "gpt-4",
		Output: []OutputItem{
			{
				Type:   "message",
				Id:     "msg_123",
				Status: "completed",
				Role:   "assistant",
				Content: []OutputContent{
					{
						Type: "output_text",
						Text: "Hello! How can I help you today?",
					},
				},
			},
		},
		Usage: &model.Usage{
			PromptTokens:     10,
			CompletionTokens: 8,
			TotalTokens:      18,
		},
	}

	// Convert to ChatCompletion format
	chatCompletion := ConvertResponseAPIToChatCompletion(responseAPI)

	// Verify basic fields
	if chatCompletion.Id != "resp_123" {
		t.Errorf("Expected id 'resp_123', got '%s'", chatCompletion.Id)
	}

	if chatCompletion.Object != "chat.completion" {
		t.Errorf("Expected object 'chat.completion', got '%s'", chatCompletion.Object)
	}

	if chatCompletion.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", chatCompletion.Model)
	}

	if chatCompletion.Created != 1234567890 {
		t.Errorf("Expected created 1234567890, got %d", chatCompletion.Created)
	}

	// Verify choices
	if len(chatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(chatCompletion.Choices))
	}

	choice := chatCompletion.Choices[0]
	if choice.Index != 0 {
		t.Errorf("Expected choice index 0, got %d", choice.Index)
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	if choice.Message.Reasoning != nil {
		t.Errorf("Expected reasoning to be nil, got '%s'", *choice.Message.Reasoning)
	}

	if choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", choice.FinishReason)
	}

	// Verify usage
	if chatCompletion.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt_tokens 10, got %d", chatCompletion.Usage.PromptTokens)
	}

	if chatCompletion.Usage.CompletionTokens != 8 {
		t.Errorf("Expected completion_tokens 8, got %d", chatCompletion.Usage.CompletionTokens)
	}

	if chatCompletion.Usage.TotalTokens != 18 {
		t.Errorf("Expected total_tokens 18, got %d", chatCompletion.Usage.TotalTokens)
	}
}

// TestConvertResponseAPIToChatCompletionWithFunctionCall tests the conversion with function calls
func TestConvertResponseAPIToChatCompletionWithFunctionCall(t *testing.T) {
	// Create a Response API response with function call (based on the real example)
	responseAPI := &ResponseAPIResponse{
		Id:        "resp_67ca09c5efe0819096d0511c92b8c890096610f474011cc0",
		Object:    "response",
		CreatedAt: 1741294021,
		Status:    "completed",
		Model:     "gpt-4.1-2025-04-14",
		Output: []OutputItem{
			{
				Type:      "function_call",
				Id:        "fc_67ca09c6bedc8190a7abfec07b1a1332096610f474011cc0",
				CallId:    "call_unLAR8MvFNptuiZK6K6HCy5k",
				Name:      "get_current_weather",
				Arguments: "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}",
				Status:    "completed",
			},
		},
		Usage: &model.Usage{
			PromptTokens:     291,
			CompletionTokens: 23,
			TotalTokens:      314,
		},
	}

	// Convert to ChatCompletion format
	chatCompletion := ConvertResponseAPIToChatCompletion(responseAPI)

	// Verify basic fields
	if chatCompletion.Id != "resp_67ca09c5efe0819096d0511c92b8c890096610f474011cc0" {
		t.Errorf("Expected id 'resp_67ca09c5efe0819096d0511c92b8c890096610f474011cc0', got '%s'", chatCompletion.Id)
	}

	if chatCompletion.Model != "gpt-4.1-2025-04-14" {
		t.Errorf("Expected model 'gpt-4.1-2025-04-14', got '%s'", chatCompletion.Model)
	}

	// Verify choices
	if len(chatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(chatCompletion.Choices))
	}

	choice := chatCompletion.Choices[0]
	if choice.Index != 0 {
		t.Errorf("Expected choice index 0, got %d", choice.Index)
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	// Verify tool calls
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}

	toolCall := choice.Message.ToolCalls[0]
	if toolCall.Id != "call_unLAR8MvFNptuiZK6K6HCy5k" {
		t.Errorf("Expected tool call id 'call_unLAR8MvFNptuiZK6K6HCy5k', got '%s'", toolCall.Id)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", toolCall.Type)
	}

	if toolCall.Function.Name != "get_current_weather" {
		t.Errorf("Expected function name 'get_current_weather', got '%s'", toolCall.Function.Name)
	}

	expectedArgs := "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}"
	if toolCall.Function.Arguments != expectedArgs {
		t.Errorf("Expected arguments '%s', got '%s'", expectedArgs, toolCall.Function.Arguments)
	}

	if choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", choice.FinishReason)
	}

	// Verify usage
	if chatCompletion.Usage.PromptTokens != 291 {
		t.Errorf("Expected prompt_tokens 291, got %d", chatCompletion.Usage.PromptTokens)
	}

	if chatCompletion.Usage.CompletionTokens != 23 {
		t.Errorf("Expected completion_tokens 23, got %d", chatCompletion.Usage.CompletionTokens)
	}

	if chatCompletion.Usage.TotalTokens != 314 {
		t.Errorf("Expected total_tokens 314, got %d", chatCompletion.Usage.TotalTokens)
	}
}

// TestConvertResponseAPIStreamToChatCompletion tests the conversion from Response API streaming format to ChatCompletion streaming format
func TestConvertResponseAPIStreamToChatCompletion(t *testing.T) {
	// Create a Response API streaming chunk
	responseAPIChunk := &ResponseAPIResponse{
		Id:        "resp_123",
		Object:    "response",
		CreatedAt: 1234567890,
		Status:    "in_progress",
		Model:     "gpt-4",
		Output: []OutputItem{
			{
				Type:   "message",
				Id:     "msg_123",
				Status: "in_progress",
				Role:   "assistant",
				Content: []OutputContent{
					{
						Type: "output_text",
						Text: "Hello",
					},
				},
			},
		},
	}

	// Convert to ChatCompletion streaming format
	streamChunk := ConvertResponseAPIStreamToChatCompletion(responseAPIChunk)

	// Verify basic fields
	if streamChunk.Id != "resp_123" {
		t.Errorf("Expected id 'resp_123', got '%s'", streamChunk.Id)
	}

	if streamChunk.Object != "chat.completion.chunk" {
		t.Errorf("Expected object 'chat.completion.chunk', got '%s'", streamChunk.Object)
	}

	if streamChunk.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", streamChunk.Model)
	}

	if streamChunk.Created != 1234567890 {
		t.Errorf("Expected created 1234567890, got %d", streamChunk.Created)
	}

	// Verify choices
	if len(streamChunk.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(streamChunk.Choices))
	}

	choice := streamChunk.Choices[0]
	if choice.Index != 0 {
		t.Errorf("Expected choice index 0, got %d", choice.Index)
	}

	if choice.Delta.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Delta.Role)
	}

	if choice.Delta.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", choice.Delta.Content)
	}

	// For in_progress status, finish_reason should be nil
	if choice.FinishReason != nil {
		t.Errorf("Expected finish_reason to be nil for in_progress status, got '%s'", *choice.FinishReason)
	}

	// Test completed status
	responseAPIChunk.Status = "completed"
	streamChunk = ConvertResponseAPIStreamToChatCompletion(responseAPIChunk)
	choice = streamChunk.Choices[0]

	if choice.FinishReason == nil || *choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop' for completed status, got %v", choice.FinishReason)
	}
}

// TestConvertResponseAPIStreamToChatCompletionWithFunctionCall tests streaming conversion with function calls
func TestConvertResponseAPIStreamToChatCompletionWithFunctionCall(t *testing.T) {
	// Create a Response API streaming chunk with function call
	responseAPIChunk := &ResponseAPIResponse{
		Id:        "resp_123",
		Object:    "response",
		CreatedAt: 1234567890,
		Status:    "completed",
		Model:     "gpt-4",
		Output: []OutputItem{
			{
				Type:      "function_call",
				Id:        "fc_123",
				CallId:    "call_456",
				Name:      "get_weather",
				Arguments: "{\"location\":\"Boston\"}",
				Status:    "completed",
			},
		},
	}

	// Convert to ChatCompletion streaming format
	streamChunk := ConvertResponseAPIStreamToChatCompletion(responseAPIChunk)

	// Verify basic fields
	if streamChunk.Id != "resp_123" {
		t.Errorf("Expected id 'resp_123', got '%s'", streamChunk.Id)
	}

	if streamChunk.Object != "chat.completion.chunk" {
		t.Errorf("Expected object 'chat.completion.chunk', got '%s'", streamChunk.Object)
	}

	// Verify choices
	if len(streamChunk.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(streamChunk.Choices))
	}

	choice := streamChunk.Choices[0]
	if choice.Index != 0 {
		t.Errorf("Expected choice index 0, got %d", choice.Index)
	}

	if choice.Delta.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Delta.Role)
	}

	// Verify tool calls
	if len(choice.Delta.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Delta.ToolCalls))
	}

	toolCall := choice.Delta.ToolCalls[0]
	if toolCall.Id != "call_456" {
		t.Errorf("Expected tool call id 'call_456', got '%s'", toolCall.Id)
	}

	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got '%s'", toolCall.Function.Name)
	}

	if toolCall.Function.Arguments != "{\"location\":\"Boston\"}" {
		t.Errorf("Expected arguments '{\"location\":\"Boston\"}', got '%s'", toolCall.Function.Arguments)
	}

	// For completed status, finish_reason should be "stop"
	if choice.FinishReason == nil || *choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop' for completed status, got %v", choice.FinishReason)
	}
}

// TestConvertResponseAPIToChatCompletionWithReasoning tests the conversion with reasoning content
func TestConvertResponseAPIToChatCompletionWithReasoning(t *testing.T) {
	// Create a Response API response with reasoning content (based on the real example)
	responseAPI := &ResponseAPIResponse{
		Id:        "resp_6848f7a7ac94819cba6af50194a156e7050d57f0136932b5",
		Object:    "response",
		CreatedAt: 1749612455,
		Status:    "completed",
		Model:     "o3-2025-04-16",
		Output: []OutputItem{
			{
				Id:   "rs_6848f7a7f800819ca52a87ae9a6a59ef050d57f0136932b5",
				Type: "reasoning",
				Summary: []OutputContent{
					{
						Type: "summary_text",
						Text: "**Telling a joke**\n\nThe user asked for a joke, which is a straightforward request. There's no conflict with the guidelines, so I can definitely comply.",
					},
				},
			},
			{
				Id:     "msg_6848f7abc86c819c877542f4a72a3f1d050d57f0136932b5",
				Type:   "message",
				Status: "completed",
				Role:   "assistant",
				Content: []OutputContent{
					{
						Type: "output_text",
						Text: "Why don't scientists trust atoms?\n\nBecause they make up everything!",
					},
				},
			},
		},
		Usage: &model.Usage{
			PromptTokens:     9,
			CompletionTokens: 83,
			TotalTokens:      92,
		},
	}

	// Convert to ChatCompletion format
	chatCompletion := ConvertResponseAPIToChatCompletion(responseAPI)

	// Verify basic fields
	if chatCompletion.Id != "resp_6848f7a7ac94819cba6af50194a156e7050d57f0136932b5" {
		t.Errorf("Expected id 'resp_6848f7a7ac94819cba6af50194a156e7050d57f0136932b5', got '%s'", chatCompletion.Id)
	}

	if chatCompletion.Model != "o3-2025-04-16" {
		t.Errorf("Expected model 'o3-2025-04-16', got '%s'", chatCompletion.Model)
	}

	// Verify choices
	if len(chatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(chatCompletion.Choices))
	}

	choice := chatCompletion.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	expectedContent := "Why don't scientists trust atoms?\n\nBecause they make up everything!"
	if choice.Message.Content != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, choice.Message.Content)
	}

	// Verify reasoning content is properly extracted
	if choice.Message.Reasoning == nil {
		t.Fatal("Expected reasoning content to be present, got nil")
	}

	expectedReasoning := "**Telling a joke**\n\nThe user asked for a joke, which is a straightforward request. There's no conflict with the guidelines, so I can definitely comply."
	if *choice.Message.Reasoning != expectedReasoning {
		t.Errorf("Expected reasoning '%s', got '%s'", expectedReasoning, *choice.Message.Reasoning)
	}

	if choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", choice.FinishReason)
	}

	// Verify usage
	if chatCompletion.Usage.PromptTokens != 9 {
		t.Errorf("Expected prompt_tokens 9, got %d", chatCompletion.Usage.PromptTokens)
	}

	if chatCompletion.Usage.CompletionTokens != 83 {
		t.Errorf("Expected completion_tokens 83, got %d", chatCompletion.Usage.CompletionTokens)
	}

	if chatCompletion.Usage.TotalTokens != 92 {
		t.Errorf("Expected total_tokens 92, got %d", chatCompletion.Usage.TotalTokens)
	}
}

// TestFunctionCallWorkflow tests the complete function calling workflow:
// ChatCompletion -> ResponseAPI -> ResponseAPI Response -> ChatCompletion
func TestFunctionCallWorkflow(t *testing.T) {
	// Step 1: Create original ChatCompletion request with tools
	originalRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "What's the weather like in Boston today?"},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: model.Function{
					Name:        "get_current_weather",
					Description: "Get the current weather in a given location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. San Francisco, CA",
							},
							"unit": map[string]interface{}{
								"type": "string",
								"enum": []string{"celsius", "fahrenheit"},
							},
						},
						"required": []string{"location", "unit"},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	// Step 2: Convert ChatCompletion to Response API format
	responseAPIRequest := ConvertChatCompletionToResponseAPI(originalRequest)

	// Verify tools are preserved in request
	if len(responseAPIRequest.Tools) != 1 {
		t.Fatalf("Expected 1 tool in request, got %d", len(responseAPIRequest.Tools))
	}

	if responseAPIRequest.Tools[0].Name != "get_current_weather" {
		t.Errorf("Expected tool name 'get_current_weather', got '%s'", responseAPIRequest.Tools[0].Name)
	}

	if responseAPIRequest.ToolChoice != "auto" {
		t.Errorf("Expected tool_choice 'auto', got '%v'", responseAPIRequest.ToolChoice)
	}

	// Step 3: Create a Response API response with function call (simulates upstream response)
	responseAPIResponse := &ResponseAPIResponse{
		Id:        "resp_67ca09c5efe0819096d0511c92b8c890096610f474011cc0",
		Object:    "response",
		CreatedAt: 1741294021,
		Status:    "completed",
		Model:     "gpt-4.1-2025-04-14",
		Output: []OutputItem{
			{
				Type:      "function_call",
				Id:        "fc_67ca09c6bedc8190a7abfec07b1a1332096610f474011cc0",
				CallId:    "call_unLAR8MvFNptuiZK6K6HCy5k",
				Name:      "get_current_weather",
				Arguments: "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}",
				Status:    "completed",
			},
		},
		Usage: &model.Usage{
			PromptTokens:     291,
			CompletionTokens: 23,
			TotalTokens:      314,
		},
	}

	// Step 4: Convert Response API response back to ChatCompletion format
	finalChatCompletion := ConvertResponseAPIToChatCompletion(responseAPIResponse)

	// Step 5: Verify the final ChatCompletion response preserves all function call information
	if len(finalChatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(finalChatCompletion.Choices))
	}

	choice := finalChatCompletion.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	// Verify tool calls are preserved
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}

	toolCall := choice.Message.ToolCalls[0]
	if toolCall.Id != "call_unLAR8MvFNptuiZK6K6HCy5k" {
		t.Errorf("Expected tool call id 'call_unLAR8MvFNptuiZK6K6HCy5k', got '%s'", toolCall.Id)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", toolCall.Type)
	}

	if toolCall.Function.Name != "get_current_weather" {
		t.Errorf("Expected function name 'get_current_weather', got '%s'", toolCall.Function.Name)
	}

	expectedArgs := "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}"
	if toolCall.Function.Arguments != expectedArgs {
		t.Errorf("Expected arguments '%s', got '%s'", expectedArgs, toolCall.Function.Arguments)
	}

	// Verify usage is preserved
	if finalChatCompletion.Usage.PromptTokens != 291 {
		t.Errorf("Expected prompt_tokens 291, got %d", finalChatCompletion.Usage.PromptTokens)
	}

	if finalChatCompletion.Usage.CompletionTokens != 23 {
		t.Errorf("Expected completion_tokens 23, got %d", finalChatCompletion.Usage.CompletionTokens)
	}

	if finalChatCompletion.Usage.TotalTokens != 314 {
		t.Errorf("Expected total_tokens 314, got %d", finalChatCompletion.Usage.TotalTokens)
	}

	t.Log("Function call workflow test completed successfully!")
	t.Logf("Original request tools: %d", len(originalRequest.Tools))
	t.Logf("Response API request tools: %d", len(responseAPIRequest.Tools))
	t.Logf("Final response tool calls: %d", len(choice.Message.ToolCalls))
}

func TestConvertWithLegacyFunctions(t *testing.T) {
	// Test legacy functions conversion
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "What's the weather?"},
		},
		Functions: []model.Function{
			{
				Name:        "get_current_weather",
				Description: "Get current weather",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]interface{}{
							"type": "string",
							"enum": []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location"},
				},
			},
		},
		FunctionCall: "auto",
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify functions are converted to tools
	if len(responseAPI.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(responseAPI.Tools))
	}

	if responseAPI.Tools[0].Type != "function" {
		t.Errorf("Expected tool type 'function', got '%s'", responseAPI.Tools[0].Type)
	}

	if responseAPI.Tools[0].Name != "get_current_weather" {
		t.Errorf("Expected function name 'get_current_weather', got '%s'", responseAPI.Tools[0].Name)
	}

	if responseAPI.ToolChoice != "auto" {
		t.Errorf("Expected tool_choice 'auto', got '%v'", responseAPI.ToolChoice)
	}

	// Verify the function parameters are preserved
	if responseAPI.Tools[0].Parameters == nil {
		t.Error("Expected function parameters to be preserved")
	}

	// Verify properties are preserved
	if props, ok := responseAPI.Tools[0].Parameters["properties"].(map[string]interface{}); ok {
		if location, ok := props["location"].(map[string]interface{}); ok {
			if location["type"] != "string" {
				t.Errorf("Expected location type 'string', got '%v'", location["type"])
			}
		} else {
			t.Error("Expected location property to be preserved")
		}
	} else {
		t.Error("Expected properties to be preserved")
	}
}

// TestLegacyFunctionCallWorkflow tests the complete legacy function calling workflow:
// ChatCompletion with Functions -> ResponseAPI -> ResponseAPI Response -> ChatCompletion
func TestLegacyFunctionCallWorkflow(t *testing.T) {
	// Step 1: Create original ChatCompletion request with legacy functions
	originalRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "What's the weather like in Boston today?"},
		},
		Functions: []model.Function{
			{
				Name:        "get_current_weather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]interface{}{
							"type": "string",
							"enum": []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location", "unit"},
				},
			},
		},
		FunctionCall: "auto",
	}

	// Step 2: Convert ChatCompletion to Response API format
	responseAPIRequest := ConvertChatCompletionToResponseAPI(originalRequest)

	// Verify functions are converted to tools in request
	if len(responseAPIRequest.Tools) != 1 {
		t.Fatalf("Expected 1 tool in request, got %d", len(responseAPIRequest.Tools))
	}

	if responseAPIRequest.Tools[0].Name != "get_current_weather" {
		t.Errorf("Expected tool name 'get_current_weather', got '%s'", responseAPIRequest.Tools[0].Name)
	}

	if responseAPIRequest.ToolChoice != "auto" {
		t.Errorf("Expected tool_choice 'auto', got '%v'", responseAPIRequest.ToolChoice)
	}

	// Step 3: Create mock Response API response (simulating what the API would return)
	responseAPIResponse := &ResponseAPIResponse{
		Id:        "resp_legacy_test",
		Object:    "response",
		CreatedAt: 1741294021,
		Status:    "completed",
		Model:     "gpt-4.1-2025-04-14",
		Output: []OutputItem{
			{
				Type:      "function_call",
				Id:        "fc_legacy_test",
				CallId:    "call_legacy_test_123",
				Name:      "get_current_weather",
				Arguments: "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}",
				Status:    "completed",
			},
		},
		ParallelToolCalls: true,
		ToolChoice:        "auto",
		Tools: []model.Tool{
			{
				Type: "function",
				Function: model.Function{
					Name:        "get_current_weather",
					Description: "Get the current weather in a given location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. San Francisco, CA",
							},
							"unit": map[string]interface{}{
								"type": "string",
								"enum": []string{"celsius", "fahrenheit"},
							},
						},
						"required": []string{"location", "unit"},
					},
				},
			},
		},
		Usage: &model.Usage{
			PromptTokens:     291,
			CompletionTokens: 23,
			TotalTokens:      314,
		},
	}

	// Step 4: Convert Response API response back to ChatCompletion format
	finalChatCompletion := ConvertResponseAPIToChatCompletion(responseAPIResponse)

	// Step 5: Verify the final ChatCompletion response preserves all function call information
	if len(finalChatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(finalChatCompletion.Choices))
	}

	choice := finalChatCompletion.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	// Verify tool calls are preserved
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}

	toolCall := choice.Message.ToolCalls[0]
	if toolCall.Id != "call_legacy_test_123" {
		t.Errorf("Expected tool call id 'call_legacy_test_123', got '%s'", toolCall.Id)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", toolCall.Type)
	}

	if toolCall.Function.Name != "get_current_weather" {
		t.Errorf("Expected function name 'get_current_weather', got '%s'", toolCall.Function.Name)
	}

	expectedArgs := "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}"
	if toolCall.Function.Arguments != expectedArgs {
		t.Errorf("Expected arguments '%s', got '%s'", expectedArgs, toolCall.Function.Arguments)
	}

	// Verify usage is preserved
	if finalChatCompletion.Usage.PromptTokens != 291 {
		t.Errorf("Expected prompt_tokens 291, got %d", finalChatCompletion.Usage.PromptTokens)
	}

	if finalChatCompletion.Usage.CompletionTokens != 23 {
		t.Errorf("Expected completion_tokens 23, got %d", finalChatCompletion.Usage.CompletionTokens)
	}

	if finalChatCompletion.Usage.TotalTokens != 314 {
		t.Errorf("Expected total_tokens 314, got %d", finalChatCompletion.Usage.TotalTokens)
	}

	t.Log("Legacy function call workflow test completed successfully!")
	t.Logf("Original request functions: %d", len(originalRequest.Functions))
	t.Logf("Response API request tools: %d", len(responseAPIRequest.Tools))
	t.Logf("Final response tool calls: %d", len(choice.Message.ToolCalls))
}
