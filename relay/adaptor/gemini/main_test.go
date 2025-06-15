package gemini

import (
	"encoding/json"
	"reflect"
	"testing"
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
