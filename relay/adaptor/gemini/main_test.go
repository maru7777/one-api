package gemini

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

func TestCleanJsonSchemaForGemini(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name: "remove additionalProperties and convert types",
			input: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"steps": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"explanation": map[string]interface{}{
									"type": "string",
								},
								"output": map[string]interface{}{
									"type": "string",
								},
							},
							"required":             []interface{}{"explanation", "output"},
							"additionalProperties": false, // Should be removed
						},
					},
					"final_answer": map[string]interface{}{
						"type": "string",
					},
				},
				"required":             []interface{}{"steps", "final_answer"},
				"additionalProperties": false, // Should be removed
				"strict":               true,  // Should be removed
			},
			expected: map[string]interface{}{
				"type": "OBJECT",
				"properties": map[string]interface{}{
					"steps": map[string]interface{}{
						"type": "ARRAY",
						"items": map[string]interface{}{
							"type": "OBJECT",
							"properties": map[string]interface{}{
								"explanation": map[string]interface{}{
									"type": "STRING",
								},
								"output": map[string]interface{}{
									"type": "STRING",
								},
							},
							"required": []interface{}{"explanation", "output"},
						},
					},
					"final_answer": map[string]interface{}{
						"type": "STRING",
					},
				},
				"required": []interface{}{"steps", "final_answer"},
			},
		},
		{
			name: "preserve supported fields",
			input: map[string]interface{}{
				"type":    "string",
				"enum":    []interface{}{"option1", "option2"},
				"format":  "date-time",
				"minimum": 0,
				"maximum": 100,
			},
			expected: map[string]interface{}{
				"type":    "STRING",
				"enum":    []interface{}{"option1", "option2"},
				"format":  "date-time",
				"minimum": 0,
				"maximum": 100,
			},
		},
		{
			name: "handle array schema",
			input: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":                 "integer",
					"minimum":              1,
					"additionalProperties": false, // Should be removed
				},
				"minItems": 1,
				"maxItems": 10,
			},
			expected: map[string]interface{}{
				"type": "ARRAY",
				"items": map[string]interface{}{
					"type":    "INTEGER",
					"minimum": 1,
				},
				"minItems": 1,
				"maxItems": 10,
			},
		},
		{
			name:     "handle primitive values",
			input:    "test",
			expected: "test",
		},
		{
			name:     "handle nil",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJsonSchemaForGemini(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				resultJSON, _ := json.MarshalIndent(result, "", "  ")
				expectedJSON, _ := json.MarshalIndent(tt.expected, "", "  ")
				t.Errorf("cleanJsonSchemaForGemini() = %s, want %s", resultJSON, expectedJSON)
			}
		})
	}
}

func TestSchemaConversionExample(t *testing.T) {
	// Test the exact schema from the error trace
	problematicSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"steps": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"explanation": map[string]interface{}{
							"type": "string",
						},
						"output": map[string]interface{}{
							"type": "string",
						},
					},
					"required":             []interface{}{"explanation", "output"},
					"additionalProperties": false,
				},
			},
			"final_answer": map[string]interface{}{
				"type": "string",
			},
		},
		"required":             []interface{}{"steps", "final_answer"},
		"additionalProperties": false,
	}

	cleaned := cleanJsonSchemaForGemini(problematicSchema)

	// Verify that additionalProperties is removed
	cleanedMap, ok := cleaned.(map[string]interface{})
	if !ok {
		t.Fatal("Expected cleaned schema to be a map")
	}

	if _, exists := cleanedMap["additionalProperties"]; exists {
		t.Error("additionalProperties should be removed from root level")
	}

	// Check nested additionalProperties removal
	if properties, ok := cleanedMap["properties"].(map[string]interface{}); ok {
		if steps, ok := properties["steps"].(map[string]interface{}); ok {
			if items, ok := steps["items"].(map[string]interface{}); ok {
				if _, exists := items["additionalProperties"]; exists {
					t.Error("additionalProperties should be removed from nested items")
				}
			}
		}
	}

	// Verify type conversion
	if cleanedMap["type"] != "OBJECT" {
		t.Errorf("Expected type to be converted to OBJECT, got %v", cleanedMap["type"])
	}
}

// Test case for the fix of: * GenerateContentRequest.contents[3].parts: contents.parts must not be empty.
func TestConvertRequestWithEmptyContentAndToolCalls(t *testing.T) {
	// Test case for OpenAI request with empty content but with tool calls
	// This reproduces the issue: * GenerateContentRequest.contents[3].parts: contents.parts must not be empty.
	openaiRequest := model.GeneralOpenAIRequest{
		Model: "gemini-2.5-flash",
		Messages: []model.Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "What time is it?",
			},
			{
				Role:    "assistant",
				Content: "", // Empty content with tool calls
				ToolCalls: []model.Tool{
					{
						Id:   "call_123",
						Type: "function",
						Function: model.Function{
							Name:      "get_current_time",
							Arguments: "{}",
						},
					},
				},
			},
			{
				Role:       "tool",
				ToolCallId: "call_123",
				Content:    `{"time": "2025-06-17T14:22:00Z"}`,
			},
		},
	}

	geminiRequest := ConvertRequest(openaiRequest)

	// Verify that no content has empty parts
	for i, content := range geminiRequest.Contents {
		if len(content.Parts) == 0 {
			t.Errorf("Content at index %d has empty parts, which violates Gemini API requirements", i)
		}
		for j, part := range content.Parts {
			// At least one field should be set in each part
			if part.Text == "" && part.InlineData == nil && part.FunctionCall == nil {
				t.Errorf("Content[%d].Parts[%d] is completely empty, which may cause API errors", i, j)
			}
		}
	}

	// Verify the structure is valid by attempting to marshal it
	_, err := json.Marshal(geminiRequest)
	if err != nil {
		t.Errorf("Failed to marshal converted request: %v", err)
	}

	// Verify that we have the expected number of contents
	// Should be: system (may be converted to SystemInstruction), user, assistant, tool
	expectedMinContents := 3 // user, assistant, tool (system might be moved to SystemInstruction)
	if len(geminiRequest.Contents) < expectedMinContents {
		t.Errorf("Expected at least %d contents, got %d", expectedMinContents, len(geminiRequest.Contents))
	}
}

func TestConvertRequestWithOnlyToolCalls(t *testing.T) {
	// Test case with a message that has only tool calls and no text content
	openaiRequest := model.GeneralOpenAIRequest{
		Model: "gemini-2.5-flash",
		Messages: []model.Message{
			{
				Role:    "assistant",
				Content: "", // Completely empty content
				ToolCalls: []model.Tool{
					{
						Id:   "call_456",
						Type: "function",
						Function: model.Function{
							Name:      "search_web",
							Arguments: `{"query": "weather today"}`,
						},
					},
				},
			},
		},
	}

	geminiRequest := ConvertRequest(openaiRequest)

	// Ensure no empty parts
	for i, content := range geminiRequest.Contents {
		if len(content.Parts) == 0 {
			t.Errorf("Content at index %d has empty parts", i)
		}
	}
}

func TestConvertRequestNormalMessage(t *testing.T) {
	// Test case for normal message to ensure we didn't break existing functionality
	openaiRequest := model.GeneralOpenAIRequest{
		Model: "gemini-2.5-flash",
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
			{
				Role:    "assistant",
				Content: "I'm doing well, thank you for asking!",
			},
		},
	}

	geminiRequest := ConvertRequest(openaiRequest)

	// Verify normal conversion still works
	if len(geminiRequest.Contents) != 2 {
		t.Errorf("Expected 2 contents, got %d", len(geminiRequest.Contents))
	}

	// Verify content is properly set
	for i, content := range geminiRequest.Contents {
		if len(content.Parts) == 0 {
			t.Errorf("Content at index %d has empty parts", i)
		}
		if len(content.Parts) > 0 && content.Parts[0].Text == "" {
			t.Errorf("Content at index %d has empty text", i)
		}
	}
}

func TestConvertRequestWithToolCallsConversion(t *testing.T) {
	// Test that OpenAI tool calls are properly converted to Gemini function calls
	openaiRequest := model.GeneralOpenAIRequest{
		Model: "gemini-2.5-flash",
		Messages: []model.Message{
			{
				Role:    "assistant",
				Content: "", // Empty content but with tool calls
				ToolCalls: []model.Tool{
					{
						Id:   "call_123",
						Type: "function",
						Function: model.Function{
							Name:      "get_current_time",
							Arguments: `{"timezone": "UTC"}`,
						},
					},
				},
			},
		},
	}

	geminiRequest := ConvertRequest(openaiRequest)

	// Verify that tool calls are converted to function calls
	if len(geminiRequest.Contents) != 1 {
		t.Errorf("Expected 1 content, got %d", len(geminiRequest.Contents))
	}

	content := geminiRequest.Contents[0]
	if len(content.Parts) != 1 {
		t.Errorf("Expected 1 part, got %d", len(content.Parts))
	}

	part := content.Parts[0]
	if part.FunctionCall == nil {
		t.Error("Expected function call to be set")
	} else {
		if part.FunctionCall.FunctionName != "get_current_time" {
			t.Errorf("Expected function name 'get_current_time', got '%s'", part.FunctionCall.FunctionName)
		}

		// Verify arguments are parsed correctly
		args, ok := part.FunctionCall.Arguments.(map[string]interface{})
		if !ok {
			t.Errorf("Expected arguments to be parsed as map, got %T", part.FunctionCall.Arguments)
		} else {
			if timezone, exists := args["timezone"]; !exists || timezone != "UTC" {
				t.Errorf("Expected timezone argument to be 'UTC', got %v", timezone)
			}
		}
	}
}

func TestConvertRequestWithToolRole(t *testing.T) {
	// Create a request with a tool role message
	req := model.GeneralOpenAIRequest{
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
			{
				Role:    "assistant",
				Content: "",
				ToolCalls: []model.Tool{
					{
						Id:   "call_123",
						Type: "function",
						Function: model.Function{
							Name:      "get_weather",
							Arguments: `{"location": "New York"}`,
						},
					},
				},
			},
			{
				Role:       "tool",
				Content:    `{"temperature": 22, "condition": "sunny"}`,
				ToolCallId: "call_123",
			},
		},
	}

	// Convert the request
	result := ConvertRequest(req)

	// Verify that the tool role was converted to user role
	if len(result.Contents) < 3 {
		t.Fatalf("Expected at least 3 contents, got %d", len(result.Contents))
	}

	// Check the tool response (should be converted to user)
	toolResponse := result.Contents[2]
	if toolResponse.Role != "user" {
		t.Errorf("Expected tool role to be converted to 'user', got '%s'", toolResponse.Role)
	}

	// Verify the content is preserved
	if len(toolResponse.Parts) == 0 {
		t.Error("Expected tool response to have parts")
	} else if toolResponse.Parts[0].Text != `{"temperature": 22, "condition": "sunny"}` {
		t.Errorf("Expected tool response content to be preserved, got '%s'", toolResponse.Parts[0].Text)
	}

	// Verify no invalid roles remain
	for i, content := range result.Contents {
		if content.Role != "user" && content.Role != "model" {
			t.Errorf("Invalid role at index %d: '%s'. Only 'user' and 'model' are allowed", i, content.Role)
		}
	}
}

func TestCleanFunctionParameters(t *testing.T) {
	// Test data that mirrors the error in the log
	input := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"dateRange": map[string]interface{}{
				"type":                 "object",
				"description":          "Date range for events",
				"additionalProperties": false, // This should be removed
				"properties": map[string]interface{}{
					"start": map[string]interface{}{
						"type":        "string",
						"format":      "date",
						"description": "Start date",
					},
					"end": map[string]interface{}{
						"type":        "string",
						"format":      "date",
						"description": "End date",
					},
				},
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
				"minLength":   1,
				"maxLength":   200,
			},
		},
		"required":             []interface{}{"query"},
		"additionalProperties": false,                            // This should be removed
		"description":          "Function to search crypto news", // This should be removed
		"strict":               true,                             // This should be removed
	}

	result := cleanFunctionParameters(input)

	// Convert to JSON and back to check structure
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check that top-level additionalProperties is removed
	if _, exists := parsed["additionalProperties"]; exists {
		t.Error("additionalProperties should be removed at top level")
	}

	// Check that description is removed
	if _, exists := parsed["description"]; exists {
		t.Error("description should be removed at top level")
	}

	// Check that strict is removed
	if _, exists := parsed["strict"]; exists {
		t.Error("strict should be removed at top level")
	}

	// Check that properties are preserved
	if _, exists := parsed["properties"]; !exists {
		t.Error("properties should be preserved")
	}

	// Check nested additionalProperties is removed
	properties, ok := parsed["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be a map")
	}

	dateRange, ok := properties["dateRange"].(map[string]interface{})
	if !ok {
		t.Fatal("dateRange should be a map")
	}

	if _, exists := dateRange["additionalProperties"]; exists {
		t.Error("additionalProperties should be removed from nested dateRange object")
	}

	// Check that nested description is preserved (only top-level description is removed)
	if _, exists := dateRange["description"]; !exists {
		t.Error("description should be preserved in nested objects")
	}

	// Check that the structure is valid
	if dateRange["type"] != "object" {
		t.Error("type should be preserved")
	}

	// Print the cleaned result for manual inspection
	t.Logf("Cleaned function parameters: %s", string(jsonBytes))
}

func TestCleanFunctionParametersWithArray(t *testing.T) {
	// Test with array that contains objects with additionalProperties
	input := []interface{}{
		map[string]interface{}{
			"additionalProperties": false,
			"type":                 "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	result := cleanFunctionParameters(input)

	resultArray, ok := result.([]interface{})
	if !ok {
		t.Fatal("result should be an array")
	}

	if len(resultArray) != 1 {
		t.Fatal("result array should have one element")
	}

	resultObj, ok := resultArray[0].(map[string]interface{})
	if !ok {
		t.Fatal("result array element should be a map")
	}

	if _, exists := resultObj["additionalProperties"]; exists {
		t.Error("additionalProperties should be removed from array element")
	}

	if resultObj["type"] != "object" {
		t.Error("type should be preserved in array element")
	}
}

func TestCleanFunctionParametersGeminiCompatibility(t *testing.T) {
	// Test case that reproduces the exact error from the log
	// This is based on the actual function schema that was causing the error
	input := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"dateRange": map[string]interface{}{
				"additionalProperties": false, // This causes the Gemini error
				"description":          "Date range for events (default: today to 7 days later)",
				"properties": map[string]interface{}{
					"end": map[string]interface{}{
						"description": "End date (default 7 days later, not after 90 days)",
						"format":      "date",
						"type":        "string",
					},
					"start": map[string]interface{}{
						"description": "Start date (default today, not before today)",
						"format":      "date",
						"type":        "string",
					},
				},
				"type": "object",
			},
			"location": map[string]interface{}{
				"description": "Location description (only when user explicitly requests address) - e.g., 'New York', 'London', 'San Francisco Bay Area'",
				"maxLength":   100,
				"type":        "string",
			},
			"query": map[string]interface{}{
				"description": "Semantic search query (natural language) for crypto news (e.g., 'Bitcoin DeFi events', 'Ethereum conferences', 'token unlocks')",
				"maxLength":   200,
				"minLength":   1,
				"type":        "string",
			},
		},
		"required": []interface{}{"query"},
	}

	result := cleanFunctionParameters(input)

	// Verify the result doesn't contain any additionalProperties
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("result should be a map")
	}

	if err := verifyNoAdditionalProperties(resultMap); err != nil {
		t.Errorf("Found additionalProperties in cleaned result: %v", err)
	}

	// Verify that important fields are preserved
	if resultMap["type"] != "object" {
		t.Error("type should be preserved")
	}

	if _, exists := resultMap["properties"]; !exists {
		t.Error("properties should be preserved")
	}

	if _, exists := resultMap["required"]; !exists {
		t.Error("required should be preserved")
	}
}

// verifyNoAdditionalProperties recursively checks that no additionalProperties fields exist
func verifyNoAdditionalProperties(obj interface{}) error {
	switch v := obj.(type) {
	case map[string]interface{}:
		if _, exists := v["additionalProperties"]; exists {
			return errors.New("found additionalProperties in object")
		}
		for key, value := range v {
			if err := verifyNoAdditionalProperties(value); err != nil {
				return fmt.Errorf("in field %s: %w", key, err)
			}
		}
	case []interface{}:
		for i, item := range v {
			if err := verifyNoAdditionalProperties(item); err != nil {
				return fmt.Errorf("in array index %d: %w", i, err)
			}
		}
	}
	return nil
}

func TestOriginalLogErrorFixed(t *testing.T) {
	// This reproduces the exact case from the log that was failing:
	// "Invalid JSON payload received. Unknown name \"additionalProperties\"
	// at 'tools[0].function_declarations[0].parameters.properties[0].value': Cannot find field."

	// Simulating the original OpenAI request with function that has nested additionalProperties
	openAIRequest := model.GeneralOpenAIRequest{
		Model: "gemini-2.5-flash",
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "find some news events",
			},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: model.Function{
					Name:        "search_crypto_news",
					Description: "Search for cryptocurrency news and events",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"dateRange": map[string]any{
								"additionalProperties": false, // This was causing the error
								"description":          "Date range for events",
								"properties": map[string]any{
									"end": map[string]any{
										"description": "End date",
										"format":      "date",
										"type":        "string",
									},
									"start": map[string]any{
										"description": "Start date",
										"format":      "date",
										"type":        "string",
									},
								},
								"type": "object",
							},
							"query": map[string]any{
								"description": "Search query",
								"maxLength":   200,
								"minLength":   1,
								"type":        "string",
							},
						},
						"required": []any{"query"},
					},
				},
			},
		},
	}

	// Convert the request - this should not panic or fail
	geminiRequest := ConvertRequest(openAIRequest)

	// Verify the conversion worked
	if geminiRequest == nil {
		t.Fatal("ConvertRequest returned nil")
	}

	if len(geminiRequest.Tools) == 0 {
		t.Fatal("Tools should not be empty")
	}

	// Extract the function declarations to verify they're clean
	tool := geminiRequest.Tools[0]
	functions, ok := tool.FunctionDeclarations.([]model.Function)
	if !ok {
		t.Fatal("FunctionDeclarations should be []model.Function")
	}

	if len(functions) == 0 {
		t.Fatal("FunctionDeclarations should not be empty")
	}

	function := functions[0]

	// Verify the function parameters no longer contain additionalProperties
	if err := verifyNoAdditionalProperties(function.Parameters); err != nil {
		t.Errorf("Function parameters still contain additionalProperties: %v", err)
	}

	t.Logf("Successfully converted request without additionalProperties errors")
}
