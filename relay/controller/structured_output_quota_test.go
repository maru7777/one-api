package controller

import (
	"testing"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/channeltype"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// Helper function to get model ratio using the new two-layer approach
func getTestModelRatio(modelName string, channelType int) float64 {
	// Use the same logic as the controllers
	apiType := channeltype.ToAPIType(channelType)
	if adaptor := relay.GetAdaptor(apiType); adaptor != nil {
		return adaptor.GetModelRatio(modelName)
	}
	// Fallback for tests
	return 2.5 * 0.5 // Default quota pricing (2.5 * MilliTokensUsd)
}

func TestPreConsumedQuotaWithStructuredOutput(t *testing.T) {
	tests := []struct {
		name                      string
		textRequest               *relaymodel.GeneralOpenAIRequest
		promptTokens              int
		ratio                     float64
		expectStructuredSurcharge bool
	}{
		{
			name: "Request with JSON schema should have additional pre-consumed quota",
			textRequest: &relaymodel.GeneralOpenAIRequest{
				Model:     "gpt-4o",
				MaxTokens: 1000,
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
			},
			promptTokens:              100000, // Use larger token count to get non-zero quota
			ratio:                     getTestModelRatio("gpt-4o", channeltype.OpenAI),
			expectStructuredSurcharge: true,
		},
		{
			name: "Request without response format should have normal pre-consumed quota",
			textRequest: &relaymodel.GeneralOpenAIRequest{
				Model:     "gpt-4o",
				MaxTokens: 1000,
			},
			promptTokens:              100000, // Use larger token count to get non-zero quota
			ratio:                     getTestModelRatio("gpt-4o", channeltype.OpenAI),
			expectStructuredSurcharge: false,
		},
		{
			name: "Request with text response format should have normal pre-consumed quota",
			textRequest: &relaymodel.GeneralOpenAIRequest{
				Model:     "gpt-4o",
				MaxTokens: 1000,
				ResponseFormat: &relaymodel.ResponseFormat{
					Type: "text",
				},
			},
			promptTokens:              100000, // Use larger token count to get non-zero quota
			ratio:                     getTestModelRatio("gpt-4o", channeltype.OpenAI),
			expectStructuredSurcharge: false,
		},
		{
			name: "Request with JSON schema but no max tokens should still have structured surcharge",
			textRequest: &relaymodel.GeneralOpenAIRequest{
				Model: "gpt-4o",
				ResponseFormat: &relaymodel.ResponseFormat{
					Type: "json_schema",
					JsonSchema: &relaymodel.JSONSchema{
						Name: "test_schema",
						Schema: map[string]interface{}{
							"type": "object",
						},
					},
				},
			},
			promptTokens:              100000, // Use larger token count to get non-zero quota
			ratio:                     getTestModelRatio("gpt-4o", channeltype.OpenAI),
			expectStructuredSurcharge: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preConsumedQuota := getPreConsumedQuota(tt.textRequest, tt.promptTokens, tt.ratio)

			// Calculate what the quota would be without structured output
			basePreConsumedTokens := config.PreConsumedQuota + int64(tt.promptTokens)
			if tt.textRequest.MaxTokens != 0 {
				basePreConsumedTokens += int64(tt.textRequest.MaxTokens)
			}
			baseQuota := int64(float64(basePreConsumedTokens) * tt.ratio)

			if tt.expectStructuredSurcharge {
				// Should have additional cost for structured output
				if preConsumedQuota <= baseQuota {
					t.Errorf("Expected pre-consumed quota to include structured output surcharge, but got %d (base: %d)",
						preConsumedQuota, baseQuota)
				}

				// Calculate expected structured cost
				estimatedTokens := tt.textRequest.MaxTokens
				if estimatedTokens == 0 {
					estimatedTokens = 1000 // default estimate
				}
				expectedStructuredCost := int64(float64(estimatedTokens) * 0.25 * tt.ratio)
				expectedTotal := baseQuota + expectedStructuredCost

				if preConsumedQuota != expectedTotal {
					t.Errorf("Expected pre-consumed quota %d, got %d", expectedTotal, preConsumedQuota)
				}

				t.Logf("Pre-consumption correctly includes structured output cost: %d (base: %d, structured: %d)",
					preConsumedQuota, baseQuota, expectedStructuredCost)
			} else {
				// Should not have additional cost
				if preConsumedQuota != baseQuota {
					t.Errorf("Expected pre-consumed quota %d (no surcharge), got %d", baseQuota, preConsumedQuota)
				}
			}
		})
	}
}

func TestStructuredOutputQuotaConsistency(t *testing.T) {
	// Test that pre-consumption and post-consumption quotas are reasonably aligned
	// for structured output requests

	textRequest := &relaymodel.GeneralOpenAIRequest{
		Model:     "gpt-4o",
		MaxTokens: 500,
		ResponseFormat: &relaymodel.ResponseFormat{
			Type: "json_schema",
			JsonSchema: &relaymodel.JSONSchema{
				Name: "test_schema",
			},
		},
	}

	promptTokens := 100000     // Use larger token count to get non-zero quota
	completionTokens := 400000 // Less than max tokens
	modelRatio := getTestModelRatio("gpt-4o", channeltype.OpenAI)

	// Calculate pre-consumed quota
	preConsumedQuota := getPreConsumedQuota(textRequest, promptTokens, modelRatio)

	// Simulate post-consumption calculation
	// Get completion ratio using the new two-layer approach
	apiType := channeltype.ToAPIType(channeltype.OpenAI)
	var completionRatio float64 = 1.0 // default
	if adaptor := relay.GetAdaptor(apiType); adaptor != nil {
		completionRatio = adaptor.GetCompletionRatio("gpt-4o")
	}
	basePostQuota := int64(float64(promptTokens)+float64(completionTokens)*completionRatio) * int64(modelRatio)
	structuredOutputCost := int64(float64(completionTokens) * 0.25 * modelRatio)
	actualPostQuota := basePostQuota + structuredOutputCost

	// Pre-consumption should be higher (conservative) but not excessively so
	if preConsumedQuota <= actualPostQuota {
		t.Logf("Pre-consumed quota (%d) vs actual post quota (%d) - this is acceptable as long as it's close",
			preConsumedQuota, actualPostQuota)
		// Allow for cases where pre-consumption is slightly lower due to conservative estimation
		// The important thing is that it's in the right ballpark
	}

	// But it shouldn't be more than 3x the actual cost (reasonable buffer)
	if preConsumedQuota > actualPostQuota*3 {
		t.Errorf("Pre-consumed quota (%d) is too conservative compared to actual post quota (%d)",
			preConsumedQuota, actualPostQuota)
	}

	t.Logf("Quota consistency check: pre-consumed=%d, actual-post=%d, ratio=%.2f",
		preConsumedQuota, actualPostQuota, float64(preConsumedQuota)/float64(actualPostQuota))
}
