package anthropic

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

func stopReasonClaude2OpenAI(reason string) string {
	switch reason {
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	default:
		return reason
	}
}

func ConvertRequest(textRequest model.GeneralOpenAIRequest) *Request {
	claudeRequest := Request{
		GeneralOpenAIRequest: textRequest,
	}

	if claudeRequest.MaxTokens == 0 {
		claudeRequest.MaxTokens = 500 // max_tokens is required
	}

	// anthropic's new messages API use system to represent the system prompt
	var filteredMessages []model.Message
	for _, msg := range claudeRequest.Messages {
		if msg.Role != "system" {
			filteredMessages = append(filteredMessages, msg)
			continue
		}

		claudeRequest.System += msg.Content.(string)
	}
	claudeRequest.Messages = filteredMessages

	claudeRequest.N = 0 // anthropic's messages API not support n
	return &claudeRequest
}

func streamResponseClaude2OpenAI(claudeResponse *Response) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = claudeResponse.Delta.Text
	finishReason := stopReasonClaude2OpenAI(claudeResponse.Delta.StopReason)
	if finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	var response openai.ChatCompletionsStreamResponse
	response.Object = "chat.completion.chunk"
	// response.Model = claudeResponse.Model
	response.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
	return &response
}

func responseClaude2OpenAI(claudeResponse *Response) *openai.TextResponse {
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:    "assistant",
			Content: strings.TrimPrefix(claudeResponse.Delta.Text, " "),
			Name:    nil,
		},
		FinishReason: stopReasonClaude2OpenAI(claudeResponse.Delta.StopReason),
	}
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", helper.GetUUID()),
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

var dataRegexp = regexp.MustCompile(`^data: (\{.*\})\B`)

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, string) {
	responseText := ""
	responseId := fmt.Sprintf("chatcmpl-%s", helper.GetUUID())
	createdTime := helper.GetTimestamp()
	scanner := bufio.NewScanner(resp.Body)
	// scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 	if atEOF && len(data) == 0 {
	// 		return 0, nil, nil
	// 	}
	// 	if i := strings.Index(string(data), "\r\n\r\n"); i >= 0 {
	// 		return i + 4, data[0:i], nil
	// 	}
	// 	if atEOF {
	// 		return len(data), data, nil
	// 	}
	// 	return 0, nil, nil
	// })
	dataChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := strings.TrimSpace(scanner.Text())
			// logger.SysLog(fmt.Sprintf("stream response: %s", data))

			matched := dataRegexp.FindAllStringSubmatch(data, -1)
			for _, match := range matched {
				data = match[1]
				// logger.SysLog(fmt.Sprintf("chunk response: %s", data))
				dataChan <- data
			}
		}

		stopChan <- true
	}()
	common.SetEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			var claudeResponse Response

			err := json.Unmarshal([]byte(data), &claudeResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				return true
			}

			switch claudeResponse.Type {
			case TypeContentStart, TypePing, TypeMessageDelta:
				return true
			case TypeContentStop, TypeMessageStop:
				if claudeResponse.Delta.StopReason == "" {
					claudeResponse.Delta.StopReason = "end_turn"
				}
			case TypeContent:
				claudeResponse.Delta.StopReason = "null"
			case TypeError:
				logger.SysError("error response: " + claudeResponse.Error.Message)
				return false
			default:
				logger.SysError("unknown response type: " + string(data))
				return true
			}

			responseText += claudeResponse.Delta.Text
			response := streamResponseClaude2OpenAI(&claudeResponse)
			response.Id = responseId
			response.Created = createdTime
			jsonStr, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case <-stopChan:
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}
	return nil, responseText
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var claudeResponse Response
	err = json.Unmarshal(responseBody, &claudeResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if claudeResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: claudeResponse.Error.Message,
				Type:    claudeResponse.Error.Type,
				Param:   "",
				Code:    claudeResponse.Error.Type,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseClaude2OpenAI(&claudeResponse)
	fullTextResponse.Model = modelName
	completionTokens := openai.CountTokenText(claudeResponse.Delta.Text, modelName)
	usage := model.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}
