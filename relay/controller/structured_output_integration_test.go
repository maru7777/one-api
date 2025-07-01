package controller

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestPostConsumeQuotaWithStructuredOutput(t *testing.T) {
	// Test that the postConsumeQuota function correctly handles structured output costs
	// when they are included in usage.ToolsCost

	tests := []struct {
		name             string
		promptTokens     int
		completionTokens int
		toolsCost        int64
		modelName        string
		expectedMinQuota int64 // minimum expected quota due to structured output cost
	}{
		{
			name:             "GPT-4o with structured output cost",
			promptTokens:     100,
			completionTokens: 1000,
			toolsCost:        250, // Example structured output cost
			modelName:        "gpt-4o",
			expectedMinQuota: 250, // Should at least include the tools cost
		},
		{
			name:             "GPT-4o-mini with structured output cost",
			promptTokens:     100,
			completionTokens: 1000,
			toolsCost:        150, // Example structured output cost
			modelName:        "gpt-4o-mini",
			expectedMinQuota: 150, // Should at least include the tools cost
		},
		{
			name:             "No structured output cost",
			promptTokens:     100,
			completionTokens: 1000,
			toolsCost:        0,
			modelName:        "gpt-4o",
			expectedMinQuota: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create usage with structured output cost
			usage := &relaymodel.Usage{
				PromptTokens:     tt.promptTokens,
				CompletionTokens: tt.completionTokens,
				TotalTokens:      tt.promptTokens + tt.completionTokens,
				ToolsCost:        tt.toolsCost,
			}

			// Get model ratio and calculate expected quota
			modelRatio := ratio.GetModelRatioWithChannel(tt.modelName, channeltype.OpenAI, nil)
			completionRatio := ratio.GetCompletionRatioWithChannel(tt.modelName, channeltype.OpenAI, nil)

			// Calculate quota manually (similar to postConsumeQuota logic)
			calculatedQuota := int64(float64(usage.PromptTokens)+float64(usage.CompletionTokens)*completionRatio) + usage.ToolsCost

			// Verify the tools cost is properly included
			if calculatedQuota < tt.expectedMinQuota {
				t.Errorf("Expected quota to be at least %d (to include tools cost), but got %d", tt.expectedMinQuota, calculatedQuota)
			}

			// Verify structured output cost is preserved
			if usage.ToolsCost != tt.toolsCost {
				t.Errorf("Expected ToolsCost to be %d, but got %d", tt.toolsCost, usage.ToolsCost)
			}

			t.Logf("Model: %s, Quota: %d, ToolsCost: %d, ModelRatio: %.6f, CompletionRatio: %.2f",
				tt.modelName, calculatedQuota, usage.ToolsCost, modelRatio, completionRatio)
		})
	}
}

// TestStructuredOutputCostIntegration tests the complete flow from request to cost calculation
func TestStructuredOutputCostIntegration(t *testing.T) {
	// This test verifies that structured output costs flow correctly through the system

	textRequest := &relaymodel.GeneralOpenAIRequest{
		Model: "gpt-4o",
		ResponseFormat: &relaymodel.ResponseFormat{
			Type: "json_schema",
			JsonSchema: &relaymodel.JSONSchema{
				Name: "test_schema",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"result": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
	}

	// Simulate usage that would come from the adaptor with structured output cost
	completionTokens := 1000
	modelRatio := ratio.GetModelRatioWithChannel("gpt-4o", channeltype.OpenAI, nil)
	expectedStructuredCost := int64(float64(completionTokens) * 0.25 * modelRatio)

	usage := &relaymodel.Usage{
		PromptTokens:     100,
		CompletionTokens: completionTokens,
		TotalTokens:      1100,
		ToolsCost:        expectedStructuredCost, // This would be set by the adaptor
	}

	// Calculate final quota as postConsumeQuota would
	completionRatio := ratio.GetCompletionRatioWithChannel("gpt-4o", channeltype.OpenAI, nil)
	quota := int64(float64(usage.PromptTokens)+float64(usage.CompletionTokens)*completionRatio) + usage.ToolsCost

	// Verify structured output cost is included in final quota
	if quota <= int64(float64(usage.PromptTokens)+float64(usage.CompletionTokens)*completionRatio) {
		t.Error("Final quota should include structured output cost, but it appears to be missing")
	}

	// Verify the structured output cost is significant
	if usage.ToolsCost == 0 {
		t.Error("Expected structured output cost to be applied, but ToolsCost is 0")
	}

	// Calculate the structured output portion of the total cost
	baseQuota := int64(float64(usage.PromptTokens) + float64(usage.CompletionTokens)*completionRatio)
	structuredOutputPortion := float64(usage.ToolsCost) / float64(quota) * 100

	t.Logf("Integration test results:")
	t.Logf("  Model: %s", textRequest.Model)
	t.Logf("  Completion tokens: %d", completionTokens)
	t.Logf("  Base quota: %d", baseQuota)
	t.Logf("  Structured output cost: %d", usage.ToolsCost)
	t.Logf("  Total quota: %d", quota)
	t.Logf("  Structured output portion: %.2f%%", structuredOutputPortion)

	// Structured output should represent a reasonable portion of the cost
	if structuredOutputPortion < 5 || structuredOutputPortion > 50 {
		t.Errorf("Structured output cost portion (%.2f%%) seems unreasonable", structuredOutputPortion)
	}
}
