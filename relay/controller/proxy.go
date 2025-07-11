// Package controller is a package for handling the relay controller
package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// RelayProxyHelper is a helper function to proxy the request to the upstream service
func RelayProxyHelper(c *gin.Context, relayMode int) *relaymodel.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := metalib.GetByContext(c)

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	resp, err := adaptor.DoRequest(c, meta, c.Request.Body)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	// do response
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		return respErr
	}

	// log proxy request with zero quota
	quotaId := c.GetInt(ctxkey.Id)
	requestId := c.GetString(ctxkey.RequestId)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Log the proxy request with zero quota
		model.RecordConsumeLog(ctx, &model.Log{
			UserId:           meta.UserId,
			ChannelId:        meta.ChannelId,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			ModelName:        "proxy",
			TokenName:        meta.TokenName,
			Quota:            0,
			Content:          "proxy request, no quota consumption",
			IsStream:         meta.IsStream,
			ElapsedTime:      helper.CalcElapsedTime(meta.StartTime),
		})
		model.UpdateUserUsedQuotaAndRequestCount(meta.UserId, 0)
		model.UpdateChannelUsedQuota(meta.ChannelId, 0)

		// also update user request cost
		docu := model.NewUserRequestCost(
			quotaId,
			requestId,
			0,
		)
		if err = docu.Insert(); err != nil {
			logger.Errorf(ctx, "insert user request cost failed: %+v", err)
		}
	}()

	return nil
}
