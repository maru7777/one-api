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
	Tools              []ResponseAPITool              `json:"tools,omitempty"`                // Optional: Array of tools the model may call
	TopP               *float64                       `json:"top_p,omitempty"`                // Optional: Alternative to sampling with temperature
	Truncation         *string                        `json:"truncation,omitempty"`           // Optional: Truncation strategy
	User               *string                        `json:"user,omitempty"`                 // Optional: Stable identifier for end-users
}

// ResponseAPITool represents the tool format for Response API requests
// This differs from the ChatCompletion tool format where function properties are nested
type ResponseAPITool struct {
	Type        string                 `json:"type"`                  // Required: "function"
	Name        string                 `json:"name"`                  // Required: Function name
	Description string                 `json:"description,omitempty"` // Optional: Function description
	Parameters  map[string]interface{} `json:"parameters,omitempty"`  // Optional: Function parameters schema
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

	// Handle tools (modern format)
	if len(request.Tools) > 0 {
		// Convert ChatCompletion tools to Response API tool format
		responseAPITools := make([]ResponseAPITool, 0, len(request.Tools))
		for _, tool := range request.Tools {
			responseAPITool := ResponseAPITool{
				Type:        tool.Type,
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			}
			responseAPITools = append(responseAPITools, responseAPITool)
		}
		responseReq.Tools = responseAPITools
		responseReq.ToolChoice = request.ToolChoice
	} else if len(request.Functions) > 0 {
		// Handle legacy functions format by converting to Response API tool format
		responseAPITools := make([]ResponseAPITool, 0, len(request.Functions))
		for _, function := range request.Functions {
			responseAPITool := ResponseAPITool{
				Type:        "function",
				Name:        function.Name,
				Description: function.Description,
				Parameters:  function.Parameters,
			}
			responseAPITools = append(responseAPITools, responseAPITool)
		}
		responseReq.Tools = responseAPITools
		responseReq.ToolChoice = request.FunctionCall
	}

	// Handle thinking/reasoning
	if request.ReasoningEffort != nil || request.Thinking != nil {
		if responseReq.Reasoning == nil {
			responseReq.Reasoning = &model.OpenAIResponseReasoning{}
		}

		if responseReq.Reasoning.Summary == nil {
			reasoningSummary := "detailed"
			responseReq.Reasoning.Summary = &reasoningSummary
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

// ResponseAPIResponse represents the OpenAI Response API response structure
// https://platform.openai.com/docs/api-reference/responses
type ResponseAPIResponse struct {
	Id                 string                         `json:"id"`                             // Unique identifier for this Response
	Object             string                         `json:"object"`                         // The object type of this resource - always set to "response"
	CreatedAt          int64                          `json:"created_at"`                     // Unix timestamp (in seconds) of when this Response was created
	Status             string                         `json:"status"`                         // The status of the response generation
	Model              string                         `json:"model"`                          // Model ID used to generate the response
	Output             []OutputItem                   `json:"output"`                         // An array of content items generated by the model
	Usage              *model.Usage                   `json:"usage,omitempty"`                // Token usage details
	Instructions       *string                        `json:"instructions,omitempty"`         // System message as the first item in the model's context
	MaxOutputTokens    *int                           `json:"max_output_tokens,omitempty"`    // Upper bound for the number of tokens
	Metadata           any                            `json:"metadata,omitempty"`             // Set of 16 key-value pairs
	ParallelToolCalls  bool                           `json:"parallel_tool_calls"`            // Whether to allow the model to run tool calls in parallel
	PreviousResponseId *string                        `json:"previous_response_id,omitempty"` // The unique ID of the previous response
	Reasoning          *model.OpenAIResponseReasoning `json:"reasoning,omitempty"`            // Configuration options for reasoning models
	ServiceTier        *string                        `json:"service_tier,omitempty"`         // Latency tier used for processing
	Temperature        *float64                       `json:"temperature,omitempty"`          // Sampling temperature used
	Text               *ResponseTextConfig            `json:"text,omitempty"`                 // Configuration options for text response
	ToolChoice         any                            `json:"tool_choice,omitempty"`          // How the model selected tools
	Tools              []model.Tool                   `json:"tools,omitempty"`                // Array of tools the model may call
	TopP               *float64                       `json:"top_p,omitempty"`                // Alternative to sampling with temperature
	Truncation         *string                        `json:"truncation,omitempty"`           // Truncation strategy
	User               *string                        `json:"user,omitempty"`                 // Stable identifier for end-users
	Error              *model.Error                   `json:"error,omitempty"`                // Error object if the response failed
	IncompleteDetails  *IncompleteDetails             `json:"incomplete_details,omitempty"`   // Details about why the response is incomplete
}

// OutputItem represents an item in the response output array
type OutputItem struct {
	Type    string          `json:"type"`              // Type of output item (e.g., "message", "reasoning", "function_call")
	Id      string          `json:"id,omitempty"`      // Unique identifier for this item
	Status  string          `json:"status,omitempty"`  // Status of this item (e.g., "completed")
	Role    string          `json:"role,omitempty"`    // Role of the message (e.g., "assistant")
	Content []OutputContent `json:"content,omitempty"` // Array of content items
	Summary []OutputContent `json:"summary,omitempty"` // Array of summary items (for reasoning)
	// Function call fields
	CallId    string `json:"call_id,omitempty"`   // Call ID for function calls
	Name      string `json:"name,omitempty"`      // Function name for function calls
	Arguments string `json:"arguments,omitempty"` // Function arguments for function calls
}

// OutputContent represents content within an output item
type OutputContent struct {
	Type        string `json:"type"`                  // Type of content (e.g., "output_text", "summary_text")
	Text        string `json:"text,omitempty"`        // Text content
	Annotations []any  `json:"annotations,omitempty"` // Annotations for the content
}

// IncompleteDetails provides details about why a response is incomplete
type IncompleteDetails struct {
	Reason string `json:"reason,omitempty"` // Reason why the response is incomplete
}

// ConvertResponseAPIToChatCompletion converts a Response API response back to ChatCompletion format
// This function follows the same pattern as ResponseClaude2OpenAI in the anthropic adaptor
func ConvertResponseAPIToChatCompletion(responseAPIResp *ResponseAPIResponse) *TextResponse {
	var responseText string
	var reasoningText string
	tools := make([]model.Tool, 0)

	// Extract content from output array
	for _, outputItem := range responseAPIResp.Output {
		switch outputItem.Type {
		case "message":
			if outputItem.Role == "assistant" {
				for _, content := range outputItem.Content {
					switch content.Type {
					case "output_text":
						responseText += content.Text
					case "reasoning":
						reasoningText += content.Text
					default:
						// Handle other content types if needed
					}
				}
			}
		case "reasoning":
			// Handle reasoning items separately
			for _, summaryContent := range outputItem.Summary {
				if summaryContent.Type == "summary_text" {
					reasoningText += summaryContent.Text
				}
			}
		case "function_call":
			// Handle function call items
			if outputItem.CallId != "" && outputItem.Name != "" {
				tool := model.Tool{
					Id:   outputItem.CallId,
					Type: "function",
					Function: model.Function{
						Name:      outputItem.Name,
						Arguments: outputItem.Arguments,
					},
				}
				tools = append(tools, tool)
			}
		}
	}

	// Handle reasoning content from reasoning field if present
	if responseAPIResp.Reasoning != nil {
		// Reasoning content would be handled here if needed
	}

	// Convert status to finish reason
	finishReason := "stop"
	switch responseAPIResp.Status {
	case "completed":
		finishReason = "stop"
	case "failed":
		finishReason = "stop"
	case "incomplete":
		finishReason = "length"
	case "cancelled":
		finishReason = "stop"
	default:
		finishReason = "stop"
	}

	choice := TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:      "assistant",
			Content:   responseText,
			Name:      nil,
			ToolCalls: tools,
		},
		FinishReason: finishReason,
	}

	if reasoningText != "" {
		choice.Message.Reasoning = &reasoningText
	}

	// Create the chat completion response
	fullTextResponse := TextResponse{
		Id:      responseAPIResp.Id,
		Model:   responseAPIResp.Model,
		Object:  "chat.completion",
		Created: responseAPIResp.CreatedAt,
		Choices: []TextResponseChoice{choice},
	}

	// Set usage if available
	if responseAPIResp.Usage != nil {
		fullTextResponse.Usage = *responseAPIResp.Usage
	}

	return &fullTextResponse
}

// ConvertResponseAPIStreamToChatCompletion converts a Response API streaming response chunk back to ChatCompletion streaming format
// This function handles individual streaming chunks from the Response API
func ConvertResponseAPIStreamToChatCompletion(responseAPIChunk *ResponseAPIResponse) *ChatCompletionsStreamResponse {
	var deltaContent string
	var reasoningText string
	var finishReason *string
	var toolCalls []model.Tool

	// Extract content from output array
	for _, outputItem := range responseAPIChunk.Output {
		switch outputItem.Type {
		case "message":
			if outputItem.Role == "assistant" {
				for _, content := range outputItem.Content {
					switch content.Type {
					case "output_text":
						deltaContent += content.Text
					case "reasoning":
						reasoningText += content.Text
					default:
						// Handle other content types if needed
					}
				}
			}
		case "reasoning":
			// Handle reasoning items separately
			for _, summaryContent := range outputItem.Summary {
				if summaryContent.Type == "summary_text" {
					reasoningText += summaryContent.Text
				}
			}
		case "function_call":
			// Handle function call items
			if outputItem.CallId != "" && outputItem.Name != "" {
				tool := model.Tool{
					Id:   outputItem.CallId,
					Type: "function",
					Function: model.Function{
						Name:      outputItem.Name,
						Arguments: outputItem.Arguments,
					},
				}
				toolCalls = append(toolCalls, tool)
			}
		}
	}

	// Convert status to finish reason for final chunks
	if responseAPIChunk.Status == "completed" {
		reason := "stop"
		finishReason = &reason
	} else if responseAPIChunk.Status == "failed" {
		reason := "stop"
		finishReason = &reason
	} else if responseAPIChunk.Status == "incomplete" {
		reason := "length"
		finishReason = &reason
	}

	// Create the streaming choice
	choice := ChatCompletionsStreamResponseChoice{
		Index: 0,
		Delta: model.Message{
			Role:    "assistant",
			Content: deltaContent,
		},
		FinishReason: finishReason,
	}

	// Set tool calls if present
	if len(toolCalls) > 0 {
		choice.Delta.ToolCalls = toolCalls
	}

	// Set reasoning content if present
	if reasoningText != "" {
		choice.Delta.Reasoning = &reasoningText
	}

	// Create the streaming response
	streamResponse := ChatCompletionsStreamResponse{
		Id:      responseAPIChunk.Id,
		Object:  "chat.completion.chunk",
		Created: responseAPIChunk.CreatedAt,
		Model:   responseAPIChunk.Model,
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}

	// Add usage if available (typically only in the final chunk)
	if responseAPIChunk.Usage != nil {
		streamResponse.Usage = responseAPIChunk.Usage
	}

	return &streamResponse
}
