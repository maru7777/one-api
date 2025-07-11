package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestStructuredOutputCostCalculation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                    string
		request                 *model.GeneralOpenAIRequest
		expectedToolsCost       int64
		completionTokens        int
		shouldHaveStructuredFee bool
	}{
		{
			name: "Request with JSON schema should have additional cost",
			request: &model.GeneralOpenAIRequest{
				Model: "gpt-4o",
				ResponseFormat: &model.ResponseFormat{
					Type: "json_schema",
					JsonSchema: &model.JSONSchema{
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
			completionTokens:        1000,
			shouldHaveStructuredFee: true,
		},
		{
			name: "Request without response format should have no additional cost",
			request: &model.GeneralOpenAIRequest{
				Model: "gpt-4o",
			},
			completionTokens:        1000,
			shouldHaveStructuredFee: false,
		},
		{
			name: "Request with text response format should have no additional cost",
			request: &model.GeneralOpenAIRequest{
				Model: "gpt-4o",
				ResponseFormat: &model.ResponseFormat{
					Type: "text",
				},
			},
			completionTokens:        1000,
			shouldHaveStructuredFee: false,
		},
		{
			name: "Request with json_schema but no schema should have no additional cost",
			request: &model.GeneralOpenAIRequest{
				Model: "gpt-4o",
				ResponseFormat: &model.ResponseFormat{
					Type: "json_schema",
				},
			},
			completionTokens:        1000,
			shouldHaveStructuredFee: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin context with proper HTTP request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Create mock response
			mockResponse := &SlimTextResponse{
				Usage: model.Usage{
					PromptTokens:     100,
					CompletionTokens: tt.completionTokens,
					TotalTokens:      100 + tt.completionTokens,
				},
			}

			responseBody, _ := json.Marshal(mockResponse)
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")

			// Set up context with request
			c.Set(ctxkey.ConvertedRequest, tt.request)

			// Create meta
			meta := &meta.Meta{
				ActualModelName: tt.request.Model,
				ChannelType:     channeltype.OpenAI,
				PromptTokens:    100,
				Mode:            relaymode.ChatCompletions,
			}

			// Create adaptor and call DoResponse
			adaptor := &Adaptor{
				ChannelType: channeltype.OpenAI,
			}

			usage, err := adaptor.DoResponse(c, resp, meta)

			if err != nil {
				t.Fatalf("DoResponse failed: %v", err)
			}

			if usage == nil {
				t.Fatal("Usage should not be nil")
			}

			// Check completion tokens
			if usage.CompletionTokens != tt.completionTokens {
				t.Errorf("Expected completion tokens %d, got %d", tt.completionTokens, usage.CompletionTokens)
			}

			// Check structured output cost
			if tt.shouldHaveStructuredFee {
				if usage.ToolsCost == 0 {
					t.Error("Expected structured output cost to be applied, but ToolsCost is 0")
				}

				// Calculate expected cost (25% of completion tokens * model ratio)
				modelRatio := ratio.GetModelRatioWithChannel(tt.request.Model, channeltype.OpenAI, nil)
				expectedStructuredCost := int64(float64(tt.completionTokens) * 0.25 * modelRatio)

				// Allow for rounding differences due to math.Ceil
				if usage.ToolsCost < expectedStructuredCost || usage.ToolsCost > expectedStructuredCost+1 {
					t.Errorf("Expected structured output cost around %d, got %d", expectedStructuredCost, usage.ToolsCost)
				}

				t.Logf("Structured output cost correctly applied: %d for model %s", usage.ToolsCost, tt.request.Model)
			} else {
				if usage.ToolsCost != 0 {
					t.Errorf("Expected no structured output cost, but got %d", usage.ToolsCost)
				}
			}
		})
	}
}

func TestStructuredOutputCostWithOriginalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test case where the request is stored in RequestModel context key
	request := &model.GeneralOpenAIRequest{
		Model: "gpt-4o",
		ResponseFormat: &model.ResponseFormat{
			Type: "json_schema",
			JsonSchema: &model.JSONSchema{
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

	// Setup gin context with proper HTTP request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Create mock response
	completionTokens := 500
	mockResponse := &SlimTextResponse{
		Usage: model.Usage{
			PromptTokens:     100,
			CompletionTokens: completionTokens,
			TotalTokens:      600,
		},
	}

	responseBody, _ := json.Marshal(mockResponse)
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
		Header:     make(http.Header),
	}
	resp.Header.Set("Content-Type", "application/json")

	// Set up context with request in RequestModel key (not ConvertedRequest)
	c.Set(ctxkey.RequestModel, request)

	// Create meta
	meta := &meta.Meta{
		ActualModelName: request.Model,
		ChannelType:     channeltype.OpenAI,
		PromptTokens:    100,
		Mode:            relaymode.ChatCompletions,
	}

	// Create adaptor and call DoResponse
	adaptor := &Adaptor{
		ChannelType: channeltype.OpenAI,
	}

	usage, err := adaptor.DoResponse(c, resp, meta)

	if err != nil {
		t.Fatalf("DoResponse failed: %v", err)
	}

	if usage == nil {
		t.Fatal("Usage should not be nil")
	}

	// Check that structured output cost is applied
	if usage.ToolsCost == 0 {
		t.Error("Expected structured output cost to be applied from RequestModel context, but ToolsCost is 0")
	}

	// Calculate expected cost
	modelRatio := ratio.GetModelRatioWithChannel(request.Model, channeltype.OpenAI, nil)
	expectedStructuredCost := int64(float64(completionTokens) * 0.25 * modelRatio)

	// Allow for rounding differences due to math.Ceil
	if usage.ToolsCost < expectedStructuredCost || usage.ToolsCost > expectedStructuredCost+1 {
		t.Errorf("Expected structured output cost around %d, got %d", expectedStructuredCost, usage.ToolsCost)
	}

	t.Logf("Structured output cost correctly applied from RequestModel context: %d", usage.ToolsCost)
}
