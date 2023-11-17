package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"one-api/common"
	"strconv"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"
	"github.com/pkoukk/tiktoken-go"
)

var stopFinishReason = "stop"

// tokenEncoderMap won't grow after initialization
var tokenEncoderMap = map[string]*tiktoken.Tiktoken{}
var defaultTokenEncoder *tiktoken.Tiktoken

func InitTokenEncoders() {
	common.SysLog("initializing token encoders")
	gpt35TokenEncoder, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get gpt-3.5-turbo token encoder: %s", err.Error()))
	}
	defaultTokenEncoder = gpt35TokenEncoder
	gpt4TokenEncoder, err := tiktoken.EncodingForModel("gpt-4")
	if err != nil {
		common.FatalLog(fmt.Sprintf("failed to get gpt-4 token encoder: %s", err.Error()))
	}
	for model := range common.ModelRatio {
		if strings.HasPrefix(model, "gpt-3.5") {
			tokenEncoderMap[model] = gpt35TokenEncoder
		} else if strings.HasPrefix(model, "gpt-4") {
			tokenEncoderMap[model] = gpt4TokenEncoder
		} else {
			tokenEncoderMap[model] = nil
		}
	}
	common.SysLog("token encoders initialized")
}

func getTokenEncoder(model string) *tiktoken.Tiktoken {
	tokenEncoder, ok := tokenEncoderMap[model]
	if ok && tokenEncoder != nil {
		return tokenEncoder
	}
	if ok {
		tokenEncoder, err := tiktoken.EncodingForModel(model)
		if err != nil {
			common.SysError(fmt.Sprintf("failed to get token encoder for model %s: %s, using encoder for gpt-3.5-turbo", model, err.Error()))
			tokenEncoder = defaultTokenEncoder
		}
		tokenEncoderMap[model] = tokenEncoder
		return tokenEncoder
	}
	return defaultTokenEncoder
}

func getTokenNum(tokenEncoder *tiktoken.Tiktoken, text string) int {
	if common.ApproximateTokenEnabled {
		return int(float64(len(text)) * 0.38)
	}
	return len(tokenEncoder.Encode(text, nil, nil))
}

// CountVisionImageToken count vision image tokens
//
// https://openai.com/pricing
func CountVisionImageToken(cnt []byte, resolution VisionImageResolution) (int, error) {
	width, height, err := imageSize(cnt)
	if err != nil {
		return 0, errors.Wrap(err, "get image size")
	}

	switch resolution {
	case VisionImageResolutionLow:
		return 85, nil // fixed price
	case VisionImageResolutionHigh:
		h := math.Ceil(float64(height) / 512)
		w := math.Ceil(float64(width) / 512)
		n := w * h
		total := 85 + 170*n
		return int(total), nil
	default:
		return 0, errors.Errorf("unsupport resolution %q", resolution)
	}
}

func imageSize(cnt []byte) (width, height int, err error) {
	contentType := http.DetectContentType(cnt)
	switch contentType {
	case "image/jpeg", "image/jpg":
		img, err := jpeg.Decode(bytes.NewReader(cnt))
		if err != nil {
			return 0, 0, errors.Wrap(err, "decode jpeg")
		}

		bounds := img.Bounds()
		return bounds.Dx(), bounds.Dy(), nil
	case "image/png":
		img, err := png.Decode(bytes.NewReader(cnt))
		if err != nil {
			return 0, 0, errors.Wrap(err, "decode png")
		}

		bounds := img.Bounds()
		return bounds.Dx(), bounds.Dy(), nil
	default:
		return 0, 0, errors.Errorf("unsupport image content type %q", contentType)
	}
}

func countTokenMessages(messages []Message, model string) int {
	tokenEncoder := getTokenEncoder(model)
	// Reference:
	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	// https://github.com/pkoukk/tiktoken-go/issues/6
	//
	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerMessage int
	var tokensPerName int
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4
		tokensPerName = -1 // If there's a name, the role is omitted
	} else {
		tokensPerMessage = 3
		tokensPerName = 1
	}
	tokenNum := 0
	for _, message := range messages {
		tokenNum += tokensPerMessage
		tokenNum += getTokenNum(tokenEncoder, message.Content)
		tokenNum += getTokenNum(tokenEncoder, message.Role)
		if message.Name != nil {
			tokenNum += tokensPerName
			tokenNum += getTokenNum(tokenEncoder, *message.Name)
		}
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum
}

func countVisonTokenMessages(messages []VisionMessage, model string) (int, error) {
	tokenEncoder := getTokenEncoder(model)
	// Reference:
	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	// https://github.com/pkoukk/tiktoken-go/issues/6
	//
	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerMessage int
	var tokensPerName int
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4
		tokensPerName = -1 // If there's a name, the role is omitted
	} else {
		tokensPerMessage = 3
		tokensPerName = 1
	}
	tokenNum := 0
	for _, message := range messages {
		tokenNum += tokensPerMessage
		for _, cnt := range message.Content {
			switch cnt.Type {
			case OpenaiVisionMessageContentTypeText:
				tokenNum += getTokenNum(tokenEncoder, cnt.Text)
			case OpenaiVisionMessageContentTypeImageUrl:
				imgblob, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(cnt.ImageUrl.URL, "data:image/jpeg;base64,"))
				if err != nil {
					return 0, errors.Wrap(err, "failed to decode base64 image")
				}

				if imgtoken, err := CountVisionImageToken(imgblob, cnt.ImageUrl.Detail); err != nil {
					return 0, errors.Wrap(err, "failed to count vision image token")
				} else {
					tokenNum += imgtoken
				}
			}
		}

		tokenNum += getTokenNum(tokenEncoder, message.Role)
		if message.Name != nil {
			tokenNum += tokensPerName
			tokenNum += getTokenNum(tokenEncoder, *message.Name)
		}
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum, nil
}

func countTokenInput(input any, model string) int {
	switch input.(type) {
	case string:
		return countTokenText(input.(string), model)
	case []string:
		text := ""
		for _, s := range input.([]string) {
			text += s
		}
		return countTokenText(text, model)
	}
	return 0
}

func countTokenText(text string, model string) int {
	tokenEncoder := getTokenEncoder(model)
	return getTokenNum(tokenEncoder, text)
}

func errorWrapper(err error, code string, statusCode int) *OpenAIErrorWithStatusCode {
	openAIError := OpenAIError{
		Message: err.Error(),
		Type:    "one_api_error",
		Code:    code,
	}
	return &OpenAIErrorWithStatusCode{
		OpenAIError: openAIError,
		StatusCode:  statusCode,
	}
}

func shouldDisableChannel(err *OpenAIError, statusCode int) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if statusCode == http.StatusUnauthorized {
		return true
	}
	if err.Type == "insufficient_quota" || err.Code == "invalid_api_key" || err.Code == "account_deactivated" {
		return true
	}
	return false
}

func setEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}

func relayErrorHandler(resp *http.Response) (openAIErrorWithStatusCode *OpenAIErrorWithStatusCode) {
	openAIErrorWithStatusCode = &OpenAIErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		OpenAIError: OpenAIError{
			Message: fmt.Sprintf("bad response status code %d", resp.StatusCode),
			Type:    "upstream_error",
			Code:    "bad_response_status_code",
			Param:   strconv.Itoa(resp.StatusCode),
		},
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}
	var textResponse TextResponse
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		return
	}
	openAIErrorWithStatusCode.OpenAIError = textResponse.Error
	return
}

func getFullRequestURL(baseURL string, requestURL string, channelType int) string {
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)
	if channelType == common.ChannelTypeOpenAI {
		if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/v1"))
		}
	}
	return fullRequestURL
}
