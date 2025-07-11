package openai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestChannelSpecificConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a sample ChatCompletion request
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "Hello, world!"},
		},
		Stream: false,
	}

	// Test cases
	testCases := []struct {
		channelType      int
		expectConversion bool
		name             string
	}{
		{channeltype.OpenAI, true, "OpenAI"},
		{channeltype.Azure, false, "Azure"},
		{channeltype.AI360, false, "AI360"},
		{channeltype.Moonshot, false, "Moonshot"},
		{channeltype.Groq, false, "Groq"},
		{channeltype.DeepSeek, false, "DeepSeek"},
		{channeltype.OpenRouter, false, "OpenRouter"},
		{channeltype.OpenAICompatible, false, "OpenAI Compatible"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{}

			// Create meta context with the channel type
			testMeta := &meta.Meta{
				Mode:           relaymode.ChatCompletions,
				ChannelType:    tc.channelType,
				RequestURLPath: "/v1/chat/completions",
				BaseURL:        "https://api.openai.com",
			}
			c.Set("meta", testMeta)

			// Create adaptor
			adaptor := &Adaptor{}
			adaptor.Init(testMeta)

			// Test URL generation
			url, err := adaptor.GetRequestURL(testMeta)
			if err != nil {
				t.Fatalf("GetRequestURL failed: %v", err)
			}

			// Check if URL was converted to /responses
			urlConverted := (url == "https://api.openai.com/v1/responses")

			// Test request conversion
			convertedReq, err := adaptor.ConvertRequest(c, relaymode.ChatCompletions, chatRequest)
			if err != nil {
				t.Fatalf("ConvertRequest failed: %v", err)
			}

			// Check if request was converted to ResponseAPIRequest
			_, isResponseAPI := convertedReq.(*ResponseAPIRequest)

			// Verify expectations
			if tc.expectConversion {
				if !urlConverted {
					t.Errorf("Expected URL conversion for %s but got: %s", tc.name, url)
				}
				if !isResponseAPI {
					t.Errorf("Expected request conversion for %s but request was not converted", tc.name)
				}
				t.Logf("✓ %s: Correctly converted to Response API", tc.name)
			} else {
				if urlConverted {
					t.Errorf("Did not expect URL conversion for %s but got: %s", tc.name, url)
				}
				if isResponseAPI {
					t.Errorf("Did not expect request conversion for %s but request was converted", tc.name)
				}
				t.Logf("✓ %s: Correctly kept as ChatCompletion API", tc.name)
			}
		})
	}
}

func TestModelSpecificConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test different models with OpenAI channel type
	testCases := []struct {
		model            string
		expectConversion bool
		name             string
	}{
		{"gpt-4", true, "GPT-4 should be converted"},
		{"gpt-4o", true, "GPT-4o should be converted"},
		{"gpt-3.5-turbo", true, "GPT-3.5-turbo should be converted"},
		{"o1-preview", true, "o1-preview should be converted"},
		// Add future models that only support ChatCompletion here
		// {"legacy-model", false, "Legacy model should not be converted"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a sample ChatCompletion request with specific model
			chatRequest := &model.GeneralOpenAIRequest{
				Model: tc.model,
				Messages: []model.Message{
					{Role: "user", Content: "Hello, world!"},
				},
				Stream: false,
			}

			// Create Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{}

			// Create meta context with OpenAI channel type
			testMeta := &meta.Meta{
				Mode:           relaymode.ChatCompletions,
				ChannelType:    channeltype.OpenAI, // Always OpenAI for this test
				RequestURLPath: "/v1/chat/completions",
				BaseURL:        "https://api.openai.com",
			}
			c.Set("meta", testMeta)

			// Create adaptor
			adaptor := &Adaptor{}
			adaptor.Init(testMeta)

			// Test request conversion
			convertedReq, err := adaptor.ConvertRequest(c, relaymode.ChatCompletions, chatRequest)
			if err != nil {
				t.Fatalf("ConvertRequest failed: %v", err)
			}

			// Check if request was converted to ResponseAPIRequest
			_, isResponseAPI := convertedReq.(*ResponseAPIRequest)

			// Verify expectations
			if tc.expectConversion {
				if !isResponseAPI {
					t.Errorf("Expected request conversion for model %s but request was not converted", tc.model)
				}
				t.Logf("✓ Model %s: Correctly converted to Response API", tc.model)
			} else {
				if isResponseAPI {
					t.Errorf("Did not expect request conversion for model %s but request was converted", tc.model)
				}
				t.Logf("✓ Model %s: Correctly kept as ChatCompletion API", tc.model)
			}
		})
	}
}
