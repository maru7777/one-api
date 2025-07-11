package relay

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/apitype"
)

// TestAdapterPricingImplementations tests that all major adapters have proper pricing implementations
func TestAdapterPricingImplementations(t *testing.T) {
	testCases := []struct {
		name        string
		apiType     int
		sampleModel string
		expectEmpty bool // true if we expect empty pricing (uses DefaultPricingMethods)
	}{
		{"OpenAI", apitype.OpenAI, "gpt-4", false},
		{"Anthropic", apitype.Anthropic, "claude-3-sonnet-20240229", false},
		{"Zhipu", apitype.Zhipu, "glm-4", false},
		{"Ali", apitype.Ali, "qwen-turbo", false},
		{"Baidu", apitype.Baidu, "ERNIE-4.0-8K", false},
		{"Tencent", apitype.Tencent, "hunyuan-lite", false},
		{"Gemini", apitype.Gemini, "gemini-pro", false},
		{"Xunfei", apitype.Xunfei, "Spark-Lite", false},
		{"VertexAI", apitype.VertexAI, "gemini-pro", false},
		// Adapters that still use DefaultPricingMethods (expected to have empty pricing)
		{"Ollama", apitype.Ollama, "llama2", true},
		{"Cohere", apitype.Cohere, "command", false},
		{"Coze", apitype.Coze, "coze-chat", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adaptor := GetAdaptor(tc.apiType)
			if adaptor == nil {
				t.Fatalf("No adaptor found for %s (type: %d)", tc.name, tc.apiType)
			}

			// Test GetDefaultModelPricing
			defaultPricing := adaptor.GetDefaultModelPricing()

			if tc.expectEmpty {
				if len(defaultPricing) != 0 {
					t.Errorf("%s: Expected empty pricing map, got %d models", tc.name, len(defaultPricing))
				}
				return // Skip further tests for adapters expected to have empty pricing
			}

			if len(defaultPricing) == 0 {
				t.Errorf("%s: GetDefaultModelPricing returned empty map", tc.name)
				return
			}

			t.Logf("%s: Found %d models with pricing", tc.name, len(defaultPricing))

			// Test GetModelRatio
			ratio := adaptor.GetModelRatio(tc.sampleModel)
			if ratio <= 0 {
				t.Errorf("%s: GetModelRatio for %s returned invalid ratio: %f", tc.name, tc.sampleModel, ratio)
			}

			// Test GetCompletionRatio
			completionRatio := adaptor.GetCompletionRatio(tc.sampleModel)
			if completionRatio <= 0 {
				t.Errorf("%s: GetCompletionRatio for %s returned invalid ratio: %f", tc.name, tc.sampleModel, completionRatio)
			}

			t.Logf("%s: Model %s - ratio=%.6f, completion_ratio=%.2f", tc.name, tc.sampleModel, ratio, completionRatio)
		})
	}
}

// TestSpecificAdapterPricing tests specific pricing for individual adapters
func TestSpecificAdapterPricing(t *testing.T) {
	t.Run("Ali_Pricing", func(t *testing.T) {
		adaptor := GetAdaptor(apitype.Ali)
		if adaptor == nil {
			t.Fatal("Ali adaptor not found")
		}

		// Ali uses RMB pricing with ratio.MilliTokensRmb = 3.5
		testModels := map[string]struct {
			expectedRatio           float64
			expectedCompletionRatio float64
		}{
			"qwen-turbo": {0.3 * 3.5, 1.0}, // 0.3 * ratio.MilliTokensRmb
			"qwen-plus":  {0.8 * 3.5, 1.0}, // 0.8 * ratio.MilliTokensRmb
			"qwen-max":   {2.4 * 3.5, 1.0}, // 2.4 * ratio.MilliTokensRmb
		}

		for model, expected := range testModels {
			ratio := adaptor.GetModelRatio(model)
			completionRatio := adaptor.GetCompletionRatio(model)

			if ratio != expected.expectedRatio {
				t.Errorf("Ali %s: expected ratio %.6f, got %.6f", model, expected.expectedRatio, ratio)
			}
			if completionRatio != expected.expectedCompletionRatio {
				t.Errorf("Ali %s: expected completion ratio %.2f, got %.2f", model, expected.expectedCompletionRatio, completionRatio)
			}
		}
	})

	t.Run("Gemini_Pricing", func(t *testing.T) {
		adaptor := GetAdaptor(apitype.Gemini)
		if adaptor == nil {
			t.Fatal("Gemini adaptor not found")
		}

		// Gemini uses USD pricing with ratio.MilliTokensUsd = 0.5
		testModels := map[string]struct {
			expectedRatio           float64
			expectedCompletionRatio float64
		}{
			"gemini-pro":       {0.5 * 0.5, 3.0},   // 0.5 * ratio.MilliTokensUsd
			"gemini-1.5-flash": {0.075 * 0.5, 4.0}, // 0.075 * ratio.MilliTokensUsd
			"gemini-1.5-pro":   {1.25 * 0.5, 4.0},  // 1.25 * ratio.MilliTokensUsd
		}

		for model, expected := range testModels {
			ratio := adaptor.GetModelRatio(model)
			completionRatio := adaptor.GetCompletionRatio(model)

			if ratio != expected.expectedRatio {
				t.Errorf("Gemini %s: expected ratio %.9f, got %.9f", model, expected.expectedRatio, ratio)
			}
			if completionRatio != expected.expectedCompletionRatio {
				t.Errorf("Gemini %s: expected completion ratio %.2f, got %.2f", model, expected.expectedCompletionRatio, completionRatio)
			}
		}
	})

	t.Run("VertexAI_Pricing", func(t *testing.T) {
		adaptor := GetAdaptor(apitype.VertexAI)
		if adaptor == nil {
			t.Fatal("VertexAI adaptor not found")
		}

		// VertexAI should have the same pricing as Gemini for shared models
		testModels := []string{"gemini-pro", "gemini-1.5-flash", "gemini-1.5-pro"}

		for _, model := range testModels {
			ratio := adaptor.GetModelRatio(model)
			completionRatio := adaptor.GetCompletionRatio(model)

			if ratio <= 0 {
				t.Errorf("VertexAI %s: invalid ratio %.9f", model, ratio)
			}
			if completionRatio <= 0 {
				t.Errorf("VertexAI %s: invalid completion ratio %.2f", model, completionRatio)
			}

			t.Logf("VertexAI %s: ratio=%.9f, completion_ratio=%.2f", model, ratio, completionRatio)
		}
	})
}

// TestPricingConsistency tests that pricing methods are consistent
func TestPricingConsistency(t *testing.T) {
	adapters := []struct {
		name    string
		apiType int
	}{
		{"Ali", apitype.Ali},
		{"Baidu", apitype.Baidu},
		{"Tencent", apitype.Tencent},
		{"Gemini", apitype.Gemini},
		{"Xunfei", apitype.Xunfei},
		{"VertexAI", apitype.VertexAI},
	}

	for _, adapter := range adapters {
		t.Run(adapter.name+"_Consistency", func(t *testing.T) {
			adaptor := GetAdaptor(adapter.apiType)
			if adaptor == nil {
				t.Fatalf("%s adaptor not found", adapter.name)
			}

			defaultPricing := adaptor.GetDefaultModelPricing()
			if len(defaultPricing) == 0 {
				t.Fatalf("%s: No default pricing found", adapter.name)
			}

			// Test that GetModelRatio and GetCompletionRatio return consistent values
			// with what's in GetDefaultModelPricing
			for model, expectedPrice := range defaultPricing {
				actualRatio := adaptor.GetModelRatio(model)
				actualCompletionRatio := adaptor.GetCompletionRatio(model)

				if actualRatio != expectedPrice.Ratio {
					t.Errorf("%s %s: GetModelRatio (%.9f) != DefaultModelPricing.Ratio (%.9f)",
						adapter.name, model, actualRatio, expectedPrice.Ratio)
				}

				if actualCompletionRatio != expectedPrice.CompletionRatio {
					t.Errorf("%s %s: GetCompletionRatio (%.2f) != DefaultModelPricing.CompletionRatio (%.2f)",
						adapter.name, model, actualCompletionRatio, expectedPrice.CompletionRatio)
				}
			}
		})
	}
}

// TestFallbackPricing tests that adapters return reasonable fallback pricing for unknown models
func TestFallbackPricing(t *testing.T) {
	adapters := []struct {
		name    string
		apiType int
	}{
		{"Ali", apitype.Ali},
		{"Baidu", apitype.Baidu},
		{"Tencent", apitype.Tencent},
		{"Gemini", apitype.Gemini},
		{"Xunfei", apitype.Xunfei},
		{"VertexAI", apitype.VertexAI},
	}

	unknownModel := "unknown-test-model-12345"

	for _, adapter := range adapters {
		t.Run(adapter.name+"_Fallback", func(t *testing.T) {
			adaptor := GetAdaptor(adapter.apiType)
			if adaptor == nil {
				t.Fatalf("%s adaptor not found", adapter.name)
			}

			// Test fallback pricing for unknown model
			ratio := adaptor.GetModelRatio(unknownModel)
			completionRatio := adaptor.GetCompletionRatio(unknownModel)

			if ratio <= 0 {
				t.Errorf("%s: Fallback ratio for unknown model should be > 0, got %.9f", adapter.name, ratio)
			}

			if completionRatio <= 0 {
				t.Errorf("%s: Fallback completion ratio for unknown model should be > 0, got %.2f", adapter.name, completionRatio)
			}

			t.Logf("%s fallback pricing: ratio=%.9f, completion_ratio=%.2f", adapter.name, ratio, completionRatio)
		})
	}
}
