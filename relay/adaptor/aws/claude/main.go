// Package aws provides the AWS adaptor for the relay service.
package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"

	"github.com/songquanpeng/one-api/common/config"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	"github.com/songquanpeng/one-api/relay/adaptor/aws/utils"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
var AwsModelIDMap = map[string]string{
	"claude-instant-1.2":           "anthropic.claude-instant-v1",
	"claude-2.0":                   "anthropic.claude-v2",
	"claude-2.1":                   "anthropic.claude-v2:1",
	"claude-3-haiku-20240307":      "anthropic.claude-3-haiku-20240307-v1:0",
	"claude-3-sonnet-20240229":     "anthropic.claude-3-sonnet-20240229-v1:0",
	"claude-3-opus-20240229":       "anthropic.claude-3-opus-20240229-v1:0",
	"claude-opus-4-20250514":       "anthropic.claude-opus-4-20250514-v1:0",
	"claude-3-5-sonnet-20240620":   "anthropic.claude-3-5-sonnet-20240620-v1:0",
	"claude-3-5-sonnet-20241022":   "anthropic.claude-3-5-sonnet-20241022-v2:0",
	"claude-3-5-sonnet-latest":     "anthropic.claude-3-5-sonnet-20241022-v2:0",
	"claude-3-5-haiku-20241022":    "anthropic.claude-3-5-haiku-20241022-v1:0",
	"claude-3-7-sonnet-latest":     "anthropic.claude-3-7-sonnet-20250219-v1:0",
	"claude-3-7-sonnet-20250219":   "anthropic.claude-3-7-sonnet-20250219-v1:0",
	"claude-sonnet-4-20250514":     "anthropic.claude-sonnet-4-20250514-v1:0",
	"claude-3-7-sonnet-latest-tag": "claude-3-7-sonnet-latest-tag",
	"claude-4-sonnet-latest-tag":   "claude-4-sonnet-latest-tag",
}

func AwsModelID(requestModel string) (string, error) {
	if awsModelID, ok := AwsModelIDMap[requestModel]; ok {
		return awsModelID, nil
	}
	return "", errors.Errorf("model %s not found", requestModel)
}

func AwsClaudeModelTransArn(c *gin.Context, awsCli *bedrockruntime.Client) string {
	reqModelID := c.GetString(ctxkey.RequestModel)

	// First, try to get ARN from channel's inference profile ARN mapping
	if channelModel, ok := c.Get(ctxkey.ChannelModel); ok {
		if channel, ok := channelModel.(*model.Channel); ok {
			arnMap := channel.GetInferenceProfileArnMap()
			if arnMap != nil {
				if arn, exists := arnMap[reqModelID]; exists && arn != "" {
					logger.Debugf(c, "Using channel inference profile ARN for model %s: %s", reqModelID, arn)
					return arn
				}
			}
		}
	}

	// No ARN mapping found in channel configuration
	return ""
}

// Deprecated: FastClaudeModelTransArn is no longer used
// ARN mapping is now handled through channel configuration

func Handler(c *gin.Context, awsCli *bedrockruntime.Client, modelName string) (*relaymodel.ErrorWithStatusCode, *relaymodel.Usage) {
	awsModelID, err := AwsModelID(c.GetString(ctxkey.RequestModel))
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "AwsModelID")), nil
	}

	// Use the enhanced cross-region profile conversion with fallback testing
	awsModelID = utils.ConvertModelID2CrossRegionProfileWithFallback(c.Request.Context(), awsModelID, awsCli.Options().Region, awsCli)
	awsReq := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(awsModelID),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	if arn := AwsClaudeModelTransArn(c, awsCli); arn != "" {
		awsReq.ModelId = aws.String(arn)
		logger.Debugf(c, "final use modelID [%s]", arn)
	}

	claudeReq_, ok := c.Get(ctxkey.ConvertedRequest)
	if !ok {
		return utils.WrapErr(errors.New("request not found")), nil
	}
	claudeReq := claudeReq_.(*anthropic.Request)
	awsClaudeReq := &Request{
		AnthropicVersion: "bedrock-2023-05-31",
	}
	if err = copier.Copy(awsClaudeReq, claudeReq); err != nil {
		return utils.WrapErr(errors.Wrap(err, "copy request")), nil
	}
	if awsClaudeReq.MaxTokens == 0 {
		awsClaudeReq.MaxTokens = config.DefaultMaxToken
	}
	awsReq.Body, err = json.Marshal(awsClaudeReq)
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "marshal request")), nil
	}

	// Track metrics for the operation
	startTime := time.Now()
	awsResp, err := awsCli.InvokeModel(c.Request.Context(), awsReq)
	latency := time.Since(startTime)

	// Update region health metrics
	utils.UpdateRegionHealthMetrics(awsCli.Options().Region, err == nil, latency, err)

	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "InvokeModel")), nil
	}

	claudeResponse := new(anthropic.Response)
	err = json.Unmarshal(awsResp.Body, claudeResponse)
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "unmarshal response")), nil
	}

	openaiResp := anthropic.ResponseClaude2OpenAI(c, claudeResponse)
	openaiResp.Model = modelName
	usage := relaymodel.Usage{
		PromptTokens:     claudeResponse.Usage.InputTokens,
		CompletionTokens: claudeResponse.Usage.OutputTokens,
		TotalTokens:      claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
	}
	openaiResp.Usage = usage

	c.JSON(http.StatusOK, openaiResp)
	return nil, &usage
}

func StreamHandler(c *gin.Context, awsCli *bedrockruntime.Client) (*relaymodel.ErrorWithStatusCode, *relaymodel.Usage) {
	createdTime := helper.GetTimestamp()
	awsModelID, err := AwsModelID(c.GetString(ctxkey.RequestModel))
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "AwsModelID")), nil
	}

	// Use the enhanced cross-region profile conversion with fallback testing
	awsModelID = utils.ConvertModelID2CrossRegionProfileWithFallback(c.Request.Context(), awsModelID, awsCli.Options().Region, awsCli)
	awsReq := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(awsModelID),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	if arn := AwsClaudeModelTransArn(c, awsCli); arn != "" {
		awsReq.ModelId = aws.String(arn)
		logger.Debugf(c, "final use modelID [%s]", arn)
	}

	claudeReq_, ok := c.Get(ctxkey.ConvertedRequest)
	if !ok {
		return utils.WrapErr(errors.New("request not found")), nil
	}
	claudeReq := claudeReq_.(*anthropic.Request)

	awsClaudeReq := &Request{
		AnthropicVersion: "bedrock-2023-05-31",
	}
	if err = copier.Copy(awsClaudeReq, claudeReq); err != nil {
		return utils.WrapErr(errors.Wrap(err, "copy request")), nil
	}
	if awsClaudeReq.MaxTokens == 0 {
		awsClaudeReq.MaxTokens = config.DefaultMaxToken
	}
	awsReq.Body, err = json.Marshal(awsClaudeReq)
	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "marshal request")), nil
	}

	// Track metrics for the operation
	startTime := time.Now()
	awsResp, err := awsCli.InvokeModelWithResponseStream(c.Request.Context(), awsReq)
	latency := time.Since(startTime)

	// Update region health metrics
	utils.UpdateRegionHealthMetrics(awsCli.Options().Region, err == nil, latency, err)

	if err != nil {
		return utils.WrapErr(errors.Wrap(err, "InvokeModelWithResponseStream")), nil
	}
	stream := awsResp.GetStream()
	defer stream.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	var usage relaymodel.Usage
	var id string
	var lastToolCallChoice openai.ChatCompletionsStreamResponseChoice

	c.Stream(func(w io.Writer) bool {
		event, ok := <-stream.Events()
		if !ok {
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}

		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			claudeResp := new(anthropic.StreamResponse)
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(claudeResp)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				return false
			}

			response, meta := anthropic.StreamResponseClaude2OpenAI(c, claudeResp)
			if meta != nil {
				usage.PromptTokens += meta.Usage.InputTokens
				usage.CompletionTokens += meta.Usage.OutputTokens
				if len(meta.Id) > 0 { // only message_start has an id, otherwise it's a finish_reason event.
					id = fmt.Sprintf("chatcmpl-%s", meta.Id)
					return true
				} else { // finish_reason case
					if len(lastToolCallChoice.Delta.ToolCalls) > 0 {
						lastArgs := &lastToolCallChoice.Delta.ToolCalls[len(lastToolCallChoice.Delta.ToolCalls)-1].Function
						if len(lastArgs.Arguments.(string)) == 0 { // compatible with OpenAI sending an empty object `{}` when no arguments.
							lastArgs.Arguments = "{}"
							response.Choices[len(response.Choices)-1].Delta.Content = nil
							response.Choices[len(response.Choices)-1].Delta.ToolCalls = lastToolCallChoice.Delta.ToolCalls
						}
					}
				}
			}
			if response == nil {
				return true
			}
			response.Id = id
			response.Model = c.GetString(ctxkey.OriginalModel)
			response.Created = createdTime

			for _, choice := range response.Choices {
				if len(choice.Delta.ToolCalls) > 0 {
					lastToolCallChoice = choice
				}
			}
			jsonStr, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
			return false
		default:
			fmt.Println("union is nil or unknown type")
			return false
		}
	})

	return nil, &usage
}
