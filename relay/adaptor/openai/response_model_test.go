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

	if responseAPI.Tools[0].Function.Name != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got '%s'", responseAPI.Tools[0].Function.Name)
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
