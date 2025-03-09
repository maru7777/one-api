package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/conv"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

const (
	dataPrefix       = "data: "
	done             = "[DONE]"
	dataPrefixLength = len(dataPrefix)
)

// StreamHandler processes streaming responses from OpenAI API
// It handles incremental content delivery and accumulates the final response text
// Returns error (if any), accumulated response text, and token usage information
func StreamHandler(c *gin.Context, resp *http.Response, relayMode int) (*model.ErrorWithStatusCode, string, *model.Usage) {
	// Initialize accumulators for the response
	responseText := ""
	reasoningText := ""
	var usage *model.Usage

	// Set up scanner for reading the stream line by line
	scanner := bufio.NewScanner(resp.Body)
	buffer := make([]byte, 256*1024) // 256KB buffer for large messages
	scanner.Buffer(buffer, len(buffer))
	scanner.Split(bufio.ScanLines)

	// Set response headers for SSE
	common.SetEventStreamHeaders(c)

	doneRendered := false

	// Process each line from the stream
	for scanner.Scan() {
		data := NormalizeDataLine(scanner.Text())

		// logger.Debugf(c.Request.Context(), "stream response: %s", data)

		// Skip lines that don't match expected format
		if len(data) < dataPrefixLength {
			continue // Ignore blank line or wrong format
		}

		// Verify line starts with expected prefix
		if data[:dataPrefixLength] != dataPrefix && data[:dataPrefixLength] != done {
			continue
		}

		// Check for stream termination
		if strings.HasPrefix(data[dataPrefixLength:], done) {
			render.StringData(c, data)
			doneRendered = true
			continue
		}

		// Process based on relay mode
		switch relayMode {
		case relaymode.ChatCompletions:
			var streamResponse ChatCompletionsStreamResponse

			// Parse the JSON response
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.Errorf(c.Request.Context(), "unmarshalling stream data %q got %+v", data, err)
				render.StringData(c, data) // Pass raw data to client if parsing fails
				continue
			}

			// Skip empty choices (Azure specific behavior)
			if len(streamResponse.Choices) == 0 && streamResponse.Usage == nil {
				continue
			}

			// Process each choice in the response
			for _, choice := range streamResponse.Choices {
				// Extract reasoning content from different possible fields
				currentReasoningChunk := extractReasoningContent(&choice.Delta)

				// Update accumulated reasoning text
				if currentReasoningChunk != "" {
					reasoningText += currentReasoningChunk
				}

				// Set the reasoning content in the format requested by client
				choice.Delta.SetReasoningContent(c.Query("reasoning_format"), currentReasoningChunk)

				// Accumulate response content
				responseText += conv.AsString(choice.Delta.Content)
			}

			// Send the processed data to the client
			render.StringData(c, data)

			// Update usage information if available
			if streamResponse.Usage != nil {
				usage = streamResponse.Usage
			}

		case relaymode.Completions:
			// Send the data immediately for Completions mode
			render.StringData(c, data)

			var streamResponse CompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				continue
			}

			// Accumulate text from all choices
			for _, choice := range streamResponse.Choices {
				responseText += choice.Text
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	// Ensure stream termination is sent to client
	if !doneRendered {
		render.Done(c)
	}

	// Clean up resources
	if err := resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", nil
	}

	// Return the complete response text (reasoning + content) and usage
	return nil, reasoningText + responseText, usage
}

// Helper function to extract reasoning content from message delta
func extractReasoningContent(delta *model.Message) string {
	content := ""

	// Extract reasoning from different possible fields
	if delta.Reasoning != nil {
		content += *delta.Reasoning
		delta.Reasoning = nil
	}

	if delta.ReasoningContent != nil {
		content += *delta.ReasoningContent
		delta.ReasoningContent = nil
	}

	return content
}

// Handler processes non-streaming responses from OpenAI API
// Returns error (if any) and token usage information
func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	// Read the entire response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}

	// Close the original response body
	if err = resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	// Parse the response JSON
	var textResponse SlimTextResponse
	if err = json.Unmarshal(responseBody, &textResponse); err != nil {
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	// Check for API errors
	if textResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error:      textResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}

	// Process reasoning content in each choice
	for _, msg := range textResponse.Choices {
		reasoningContent := processReasoningContent(&msg)

		// Set reasoning in requested format if content exists
		if reasoningContent != "" {
			msg.SetReasoningContent(c.Query("reasoning_format"), reasoningContent)
		}
	}

	// Reset response body for forwarding to client
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	logger.Debugf(c.Request.Context(), "handler response: %s", string(responseBody))

	// Forward all response headers (not just first value of each)
	for k, values := range resp.Header {
		for _, v := range values {
			c.Writer.Header().Add(k, v)
		}
	}

	// Set response status and copy body to client
	c.Writer.WriteHeader(resp.StatusCode)
	if _, err = io.Copy(c.Writer, resp.Body); err != nil {
		return ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}

	// Close the reset body
	if err = resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	// Calculate token usage if not provided by API
	calculateTokenUsage(&textResponse, promptTokens, modelName)

	return nil, &textResponse.Usage
}

// processReasoningContent is a helper function to extract and process reasoning content from the message
func processReasoningContent(msg *TextResponseChoice) string {
	var reasoningContent string

	// Check different locations for reasoning content
	switch {
	case msg.Reasoning != nil:
		reasoningContent = *msg.Reasoning
		msg.Reasoning = nil
	case msg.ReasoningContent != nil:
		reasoningContent = *msg.ReasoningContent
		msg.ReasoningContent = nil
	case msg.Message.Reasoning != nil:
		reasoningContent = *msg.Message.Reasoning
		msg.Message.Reasoning = nil
	case msg.Message.ReasoningContent != nil:
		reasoningContent = *msg.Message.ReasoningContent
		msg.Message.ReasoningContent = nil
	}

	return reasoningContent
}

// Helper function to calculate token usage
func calculateTokenUsage(response *SlimTextResponse, promptTokens int, modelName string) {
	// Calculate tokens if not provided by the API
	if response.Usage.TotalTokens == 0 ||
		(response.Usage.PromptTokens == 0 && response.Usage.CompletionTokens == 0) {

		completionTokens := 0
		for _, choice := range response.Choices {
			// Count content tokens
			completionTokens += CountTokenText(choice.Message.StringContent(), modelName)

			// Count reasoning tokens in all possible locations
			if choice.Message.Reasoning != nil {
				completionTokens += CountToken(*choice.Message.Reasoning)
			}
			if choice.Message.ReasoningContent != nil {
				completionTokens += CountToken(*choice.Message.ReasoningContent)
			}
			if choice.Reasoning != nil {
				completionTokens += CountToken(*choice.Reasoning)
			}
			if choice.ReasoningContent != nil {
				completionTokens += CountToken(*choice.ReasoningContent)
			}
		}

		// Set usage values
		response.Usage = model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	} else if hasAudioTokens(response) {
		// Handle audio tokens conversion
		calculateAudioTokens(response, modelName)
	}
}

// Helper function to check if response has audio tokens
func hasAudioTokens(response *SlimTextResponse) bool {
	return (response.PromptTokensDetails != nil && response.PromptTokensDetails.AudioTokens > 0) ||
		(response.CompletionTokensDetails != nil && response.CompletionTokensDetails.AudioTokens > 0)
}

// Helper function to calculate audio token usage
func calculateAudioTokens(response *SlimTextResponse, modelName string) {
	// Convert audio tokens for prompt
	if response.PromptTokensDetails != nil {
		response.Usage.PromptTokens = response.PromptTokensDetails.TextTokens +
			int(math.Ceil(
				float64(response.PromptTokensDetails.AudioTokens)*
					ratio.GetAudioPromptRatio(modelName),
			))
	}

	// Convert audio tokens for completion
	if response.CompletionTokensDetails != nil {
		response.Usage.CompletionTokens = response.CompletionTokensDetails.TextTokens +
			int(math.Ceil(
				float64(response.CompletionTokensDetails.AudioTokens)*
					ratio.GetAudioPromptRatio(modelName)*ratio.GetAudioCompletionRatio(modelName),
			))
	}

	// Calculate total tokens
	response.Usage.TotalTokens = response.Usage.PromptTokens + response.Usage.CompletionTokens
}
