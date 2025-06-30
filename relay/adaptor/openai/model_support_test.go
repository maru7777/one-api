package openai

import (
	"testing"
)

func TestIsModelsOnlySupportedByChatCompletionAPI(t *testing.T) {
	testCases := []struct {
		model    string
		expected bool
		name     string
	}{
		{"gpt-4", false, "GPT-4 should support Response API"},
		{"gpt-4o", false, "GPT-4o should support Response API"},
		{"gpt-3.5-turbo", false, "GPT-3.5-turbo should support Response API"},
		{"o1-preview", false, "o1-preview should support Response API"},
		{"o1-mini", false, "o1-mini should support Response API"},
		{"", false, "Empty model should support Response API"},
		{"unknown-model", false, "Unknown model should support Response API"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsModelsOnlySupportedByChatCompletionAPI(tc.model)
			if result != tc.expected {
				t.Errorf("IsModelsOnlySupportedByChatCompletionAPI(%q) = %v, want %v", tc.model, result, tc.expected)
			}
		})
	}
}

func TestIsModelsOnlySupportedByChatCompletionAPI_DefaultBehavior(t *testing.T) {
	// Test that the function returns false by default as specified
	// This ensures all models are allowed for Response API conversion initially
	models := []string{
		"gpt-4",
		"gpt-4o",
		"gpt-3.5-turbo",
		"claude-3",
		"random-model-name",
		"",
	}

	for _, model := range models {
		if IsModelsOnlySupportedByChatCompletionAPI(model) {
			t.Errorf("IsModelsOnlySupportedByChatCompletionAPI(%q) should return false by default, got true", model)
		}
	}
}
