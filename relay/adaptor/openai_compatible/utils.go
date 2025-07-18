package openai_compatible

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/model"
)

const (
	DataPrefix       = "data: "
	DataPrefixLength = len(DataPrefix)
	Done             = "[DONE]"
)

// ChatCompletionsStreamResponse represents the streaming response structure
type ChatCompletionsStreamResponse struct {
	Id      string                                `json:"id"`
	Object  string                                `json:"object"`
	Created int64                                 `json:"created"`
	Model   string                                `json:"model"`
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
	Usage   *model.Usage                          `json:"usage,omitempty"`
}

type ChatCompletionsStreamResponseChoice struct {
	Index        int           `json:"index"`
	Delta        model.Message `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
}

// SlimTextResponse represents the non-streaming response structure
type SlimTextResponse struct {
	Choices     []TextResponseChoice `json:"choices"`
	model.Usage `json:"usage"`
	Error       model.Error `json:"error"`
}

type TextResponseChoice struct {
	Index        int           `json:"index"`
	Message      model.Message `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// NormalizeDataLine normalizes SSE data lines
func NormalizeDataLine(data string) string {
	if strings.HasPrefix(data, "data:") {
		content := strings.TrimLeft(data[len("data:"):], " ")
		return "data: " + content
	}
	return data
}

// ErrorWrapper creates an error response
func ErrorWrapper(err error, code string, statusCode int) *model.ErrorWithStatusCode {
	logger.Error(context.TODO(), fmt.Sprintf("[%s]%+v", code, err))
	return &model.ErrorWithStatusCode{
		Error: model.Error{
			Message: err.Error(),
			Type:    "one_api_error",
			Code:    code,
		},
		StatusCode: statusCode,
	}
}

// CountTokenText estimates token count (simplified implementation)
func CountTokenText(text string, modelName string) int {
	// Simple estimation: ~4 characters per token
	return len(text) / 4
}

// GetFullRequestURL constructs the full request URL for OpenAI-compatible APIs
func GetFullRequestURL(baseURL string, requestURL string, channelType int) string {
	if channelType == channeltype.OpenAICompatible {
		return fmt.Sprintf("%s%s", strings.TrimSuffix(baseURL, "/"), strings.TrimPrefix(requestURL, "/v1"))
	}
	return fmt.Sprintf("%s%s", baseURL, requestURL)
}

// StreamHandler processes streaming responses from OpenAI-compatible APIs
func StreamHandler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseText := ""
	var usage *model.Usage

	scanner := bufio.NewScanner(resp.Body)
	buffer := make([]byte, 1024*1024) // 1MB buffer
	scanner.Buffer(buffer, len(buffer))
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)
	doneRendered := false

	for scanner.Scan() {
		data := NormalizeDataLine(scanner.Text())

		if len(data) < DataPrefixLength {
			continue
		}

		if data[:DataPrefixLength] != DataPrefix && data[:DataPrefixLength] != Done {
			continue
		}

		if strings.HasPrefix(data[DataPrefixLength:], Done) {
			render.StringData(c, data)
			doneRendered = true
			continue
		}

		// Parse the streaming chunk
		var streamResponse ChatCompletionsStreamResponse
		jsonData := data[DataPrefixLength:]
		if err := json.Unmarshal([]byte(jsonData), &streamResponse); err != nil {
			continue // Skip malformed chunks
		}

		// Accumulate response text
		for _, choice := range streamResponse.Choices {
			responseText += choice.Delta.StringContent()
		}

		// Accumulate usage information
		if streamResponse.Usage != nil {
			usage = streamResponse.Usage
		}

		// Forward the chunk to client
		render.StringData(c, data)
	}

	if err := scanner.Err(); err != nil {
		return ErrorWrapper(err, "read_stream_failed", http.StatusInternalServerError), usage
	}

	if !doneRendered {
		render.StringData(c, "data: "+Done)
	}

	if err := resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), usage
	}

	// Calculate usage if not provided
	if usage == nil {
		usage = &model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: CountTokenText(responseText, modelName),
		}
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return nil, usage
}

// Handler processes non-streaming responses from OpenAI-compatible APIs
func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}

	if err = resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	var textResponse SlimTextResponse
	if err = json.Unmarshal(responseBody, &textResponse); err != nil {
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if textResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error:      textResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}

	// Forward the response to client
	c.JSON(resp.StatusCode, textResponse)

	// Calculate usage if not provided
	usage := textResponse.Usage
	if usage.PromptTokens == 0 {
		usage.PromptTokens = promptTokens
	}
	if usage.CompletionTokens == 0 {
		// Calculate completion tokens from response text
		responseText := ""
		for _, choice := range textResponse.Choices {
			responseText += choice.Message.StringContent()
		}
		usage.CompletionTokens = CountTokenText(responseText, modelName)
	}
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return nil, &usage
}
