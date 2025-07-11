package openai

import (
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestGetRequestURLWithModelSupport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adaptor := &Adaptor{}

	testCases := []struct {
		name          string
		modelName     string
		expectedPath  string
		expectedURL   string
		shouldConvert bool
	}{
		{
			name:          "GPT-4 should use Response API",
			modelName:     "gpt-4",
			expectedPath:  "/v1/responses",
			expectedURL:   "https://api.openai.com/v1/responses",
			shouldConvert: true,
		},
		{
			name:          "GPT-4 Search should use ChatCompletion API",
			modelName:     "gpt-4-search-2024-12-20",
			expectedPath:  "/v1/chat/completions",
			expectedURL:   "https://api.openai.com/v1/chat/completions",
			shouldConvert: false,
		},
		{
			name:          "GPT-4o Search should use ChatCompletion API",
			modelName:     "gpt-4o-search-2024-12-20",
			expectedPath:  "/v1/chat/completions",
			expectedURL:   "https://api.openai.com/v1/chat/completions",
			shouldConvert: false,
		},
		{
			name:          "Regular GPT-4o should use Response API",
			modelName:     "gpt-4o",
			expectedPath:  "/v1/responses",
			expectedURL:   "https://api.openai.com/v1/responses",
			shouldConvert: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gin context
			c, _ := gin.CreateTestContext(nil)

			// Setup meta with ChatCompletion mode and OpenAI channel
			relayMeta := &meta.Meta{
				Mode:            relaymode.ChatCompletions,
				ChannelType:     channeltype.OpenAI,
				BaseURL:         "https://api.openai.com",
				RequestURLPath:  "/v1/chat/completions", // Original ChatCompletion path
				ActualModelName: tc.modelName,
			}

			// Store meta in context
			meta.Set2Context(c, relayMeta)

			// Get the request URL
			requestURL, err := adaptor.GetRequestURL(relayMeta)
			if err != nil {
				t.Fatalf("GetRequestURL failed: %v", err)
			}

			// Verify the URL matches expectation
			if requestURL != tc.expectedURL {
				t.Errorf("Expected URL %s, got %s", tc.expectedURL, requestURL)
			}

			// Verify model support detection
			onlyChatCompletion := IsModelsOnlySupportedByChatCompletionAPI(tc.modelName)
			if tc.shouldConvert && onlyChatCompletion {
				t.Errorf("Model %s should support Response API but IsModelsOnlySupportedByChatCompletionAPI returned true", tc.modelName)
			}
			if !tc.shouldConvert && !onlyChatCompletion {
				t.Errorf("Model %s should only support ChatCompletion API but IsModelsOnlySupportedByChatCompletionAPI returned false", tc.modelName)
			}

			t.Logf("✓ %s: URL=%s, OnlyChatCompletion=%v", tc.name, requestURL, onlyChatCompletion)
		})
	}
}

func TestModelSupportConsistencyBetweenURLAndConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adaptor := &Adaptor{}

	testModels := []string{
		"gpt-4",
		"gpt-4o",
		"gpt-4-search-2024-12-20",
		"gpt-4o-search-2024-12-20",
		"gpt-3.5-turbo-search-enhanced",
		"gpt-3.5-turbo",
		"o1-preview",
	}

	for _, modelName := range testModels {
		t.Run(modelName, func(t *testing.T) {
			// Setup gin context
			c, _ := gin.CreateTestContext(nil)

			// Setup meta
			relayMeta := &meta.Meta{
				Mode:            relaymode.ChatCompletions,
				ChannelType:     channeltype.OpenAI,
				BaseURL:         "https://api.openai.com",
				RequestURLPath:  "/v1/chat/completions",
				ActualModelName: modelName,
			}
			meta.Set2Context(c, relayMeta)

			// Get URL decision
			requestURL, err := adaptor.GetRequestURL(relayMeta)
			if err != nil {
				t.Fatalf("GetRequestURL failed: %v", err)
			}

			// Get conversion decision
			onlyChatCompletion := IsModelsOnlySupportedByChatCompletionAPI(modelName)

			// Check consistency
			usesResponseAPI := requestURL == "https://api.openai.com/v1/responses"
			usesChatCompletionAPI := requestURL == "https://api.openai.com/v1/chat/completions"

			if onlyChatCompletion && !usesChatCompletionAPI {
				t.Errorf("Model %s is marked as ChatCompletion-only but URL uses Response API: %s", modelName, requestURL)
			}

			if !onlyChatCompletion && !usesResponseAPI {
				t.Errorf("Model %s should use Response API but URL uses ChatCompletion API: %s", modelName, requestURL)
			}

			t.Logf("✓ Model %s: OnlyChat=%v, URL=%s", modelName, onlyChatCompletion, requestURL)
		})
	}
}
