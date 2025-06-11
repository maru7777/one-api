package openai

import (
	"github.com/songquanpeng/one-api/relay/model"
)

// ResponseAPIRequest represents the OpenAI Response API request structure
// https://platform.openai.com/docs/api-reference/responses
type ResponseAPIRequest struct {
	Input              []any                          `json:"input"`                          // Required: Text, image, or file inputs to the model
	Model              string                         `json:"model"`                          // Required: Model ID used to generate the response
	Background         *bool                          `json:"background,omitempty"`           // Optional: Whether to run the model response in the background
	Include            []string                       `json:"include,omitempty"`              // Optional: Additional output data to include
	Instructions       *string                        `json:"instructions,omitempty"`         // Optional: System message as the first item in the model's context
	MaxOutputTokens    *int                           `json:"max_output_tokens,omitempty"`    // Optional: Upper bound for the number of tokens
	Metadata           any                            `json:"metadata,omitempty"`             // Optional: Set of 16 key-value pairs
	ParallelToolCalls  *bool                          `json:"parallel_tool_calls,omitempty"`  // Optional: Whether to allow the model to run tool calls in parallel
	PreviousResponseId *string                        `json:"previous_response_id,omitempty"` // Optional: The unique ID of the previous response
	Reasoning          *model.OpenAIResponseReasoning `json:"reasoning,omitempty"`            // Optional: Configuration options for reasoning models
	ServiceTier        *string                        `json:"service_tier,omitempty"`         // Optional: Latency tier to use for processing
	Store              *bool                          `json:"store,omitempty"`                // Optional: Whether to store the generated model response
	Stream             *bool                          `json:"stream,omitempty"`               // Optional: If set to true, model response data will be streamed
	Temperature        *float64                       `json:"temperature,omitempty"`          // Optional: Sampling temperature
	Text               *ResponseTextConfig            `json:"text,omitempty"`                 // Optional: Configuration options for a text response
	ToolChoice         any                            `json:"tool_choice,omitempty"`          // Optional: How the model should select tools
	Tools              []model.Tool                   `json:"tools,omitempty"`                // Optional: Array of tools the model may call
	TopP               *float64                       `json:"top_p,omitempty"`                // Optional: Alternative to sampling with temperature
	Truncation         *string                        `json:"truncation,omitempty"`           // Optional: Truncation strategy
	User               *string                        `json:"user,omitempty"`                 // Optional: Stable identifier for end-users
}

// ResponseTextConfig represents the text configuration for Response API
type ResponseTextConfig struct {
	Type       *string           `json:"type,omitempty"`        // Optional: Type of text response (e.g., "text")
	JSONSchema *model.JSONSchema `json:"json_schema,omitempty"` // Optional: JSON schema for structured outputs
}

// ConvertChatCompletionToResponseAPI converts a ChatCompletion request to Response API format
func ConvertChatCompletionToResponseAPI(request *model.GeneralOpenAIRequest) *ResponseAPIRequest {
	responseReq := &ResponseAPIRequest{
		Model: request.Model,
		Input: make([]any, 0, len(request.Messages)),
	}

	// Convert messages to input - Response API expects messages directly in the input array
	for _, message := range request.Messages {
		responseReq.Input = append(responseReq.Input, message)
	}

	// Map other fields
	if request.MaxTokens > 0 {
		responseReq.MaxOutputTokens = &request.MaxTokens
	}
	if request.MaxCompletionTokens != nil {
		responseReq.MaxOutputTokens = request.MaxCompletionTokens
	}

	responseReq.Temperature = request.Temperature
	responseReq.TopP = request.TopP
	responseReq.Stream = &request.Stream
	responseReq.User = &request.User
	responseReq.Store = request.Store
	responseReq.Metadata = request.Metadata

	if request.ServiceTier != nil {
		responseReq.ServiceTier = request.ServiceTier
	}

	if request.ParallelTooCalls != nil {
		responseReq.ParallelToolCalls = request.ParallelTooCalls
	}

	// Handle tools
	if len(request.Tools) > 0 {
		responseReq.Tools = request.Tools
		responseReq.ToolChoice = request.ToolChoice
	}

	// Handle thinking/reasoning
	if request.ReasoningEffort != nil || request.Thinking != nil {
		if responseReq.Reasoning == nil {
			responseReq.Reasoning = &model.OpenAIResponseReasoning{}
		}

		if responseReq.Reasoning.Summary == nil {
			auto := "auto"
			responseReq.Reasoning.Summary = &auto
		}

		if request.ReasoningEffort != nil {
			responseReq.Reasoning.Effort = request.ReasoningEffort
		}
	}

	// Handle response format
	if request.ResponseFormat != nil {
		textConfig := &ResponseTextConfig{}
		if request.ResponseFormat.Type != "" {
			textConfig.Type = &request.ResponseFormat.Type
		}
		if request.ResponseFormat.JsonSchema != nil {
			textConfig.JSONSchema = request.ResponseFormat.JsonSchema
		}
		responseReq.Text = textConfig
	}

	// Handle system message as instructions
	if len(request.Messages) > 0 && request.Messages[0].Role == "system" {
		systemContent := request.Messages[0].StringContent()
		responseReq.Instructions = &systemContent

		// Remove system message from input since it's now in instructions
		responseReq.Input = responseReq.Input[1:]
	}

	return responseReq
}
