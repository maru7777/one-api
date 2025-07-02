package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

func ReturnPreConsumedQuota(ctx context.Context, preConsumedQuota int64, tokenId int) {
	if preConsumedQuota != 0 {
		go func(ctx context.Context) {
			// return pre-consumed quota
			err := model.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
			if err != nil {
				logger.Error(ctx, "error return pre-consumed quota: "+err.Error())
			}
		}(ctx)
	}
}

// PostConsumeQuota handles simple billing for Audio API (legacy compatibility)
// SAFETY: This function is preserved for backward compatibility with Audio API
// WARNING: This function logs totalQuota as promptTokens and sets completionTokens to 0
func PostConsumeQuota(ctx context.Context, tokenId int, quotaDelta int64, totalQuota int64, userId int, channelId int, modelRatio float64, groupRatio float64, modelName string, tokenName string) {
	// Input validation for safety
	if ctx == nil {
		logger.SysError("PostConsumeQuota: context is nil")
		return
	}
	if tokenId <= 0 {
		logger.Error(ctx, fmt.Sprintf("PostConsumeQuota: invalid tokenId %d", tokenId))
		return
	}
	if userId <= 0 {
		logger.Error(ctx, fmt.Sprintf("PostConsumeQuota: invalid userId %d", userId))
		return
	}
	if channelId <= 0 {
		logger.Error(ctx, fmt.Sprintf("PostConsumeQuota: invalid channelId %d", channelId))
		return
	}
	if modelName == "" {
		logger.Error(ctx, "PostConsumeQuota: modelName is empty")
		return
	}

	// quotaDelta is remaining quota to be consumed
	err := model.PostConsumeTokenQuota(tokenId, quotaDelta)
	if err != nil {
		logger.SysError("error consuming token remain quota: " + err.Error())
	}
	err = model.CacheUpdateUserQuota(ctx, userId)
	if err != nil {
		logger.SysError("error update user quota cache: " + err.Error())
	}
	// totalQuota is total quota consumed
	// Always log the request for tracking purposes, regardless of quota amount
	logContent := fmt.Sprintf("model rate %.2f, group rate %.2f", modelRatio, groupRatio)
	model.RecordConsumeLog(ctx, &model.Log{
		UserId:           userId,
		ChannelId:        channelId,
		PromptTokens:     int(totalQuota), // NOTE: For Audio API, total quota is logged as prompt tokens
		CompletionTokens: 0,               // NOTE: Audio API doesn't have separate completion tokens
		ModelName:        modelName,
		TokenName:        tokenName,
		Quota:            int(totalQuota),
		Content:          logContent,
	})

	// Only update quotas when totalQuota > 0
	if totalQuota > 0 {
		model.UpdateUserUsedQuotaAndRequestCount(userId, totalQuota)
		model.UpdateChannelUsedQuota(channelId, totalQuota)
	}
	if totalQuota <= 0 {
		logger.Error(ctx, fmt.Sprintf("totalQuota consumed is %d, something is wrong", totalQuota))
	}
}

// PostConsumeQuotaDetailed handles detailed billing for ChatCompletion and Response API requests
// This function properly logs individual prompt and completion tokens with additional metadata
// SAFETY: This function validates all inputs to prevent billing errors
func PostConsumeQuotaDetailed(ctx context.Context, tokenId int, quotaDelta int64, totalQuota int64,
	userId int, channelId int, promptTokens int, completionTokens int,
	modelRatio float64, groupRatio float64, modelName string, tokenName string,
	isStream bool, startTime time.Time, systemPromptReset bool,
	completionRatio float64, toolsCost int64) {

	// Input validation for safety
	if ctx == nil {
		logger.SysError("PostConsumeQuotaDetailed: context is nil")
		return
	}
	if tokenId <= 0 {
		logger.Error(ctx, fmt.Sprintf("PostConsumeQuotaDetailed: invalid tokenId %d", tokenId))
		return
	}
	if userId <= 0 {
		logger.Error(ctx, fmt.Sprintf("PostConsumeQuotaDetailed: invalid userId %d", userId))
		return
	}
	if channelId <= 0 {
		logger.Error(ctx, fmt.Sprintf("PostConsumeQuotaDetailed: invalid channelId %d", channelId))
		return
	}
	if promptTokens < 0 || completionTokens < 0 {
		logger.Error(ctx, fmt.Sprintf("PostConsumeQuotaDetailed: negative token counts - prompt: %d, completion: %d", promptTokens, completionTokens))
		return
	}
	if modelName == "" {
		logger.Error(ctx, "PostConsumeQuotaDetailed: modelName is empty")
		return
	}

	// quotaDelta is remaining quota to be consumed
	err := model.PostConsumeTokenQuota(tokenId, quotaDelta)
	if err != nil {
		logger.SysError("error consuming token remain quota: " + err.Error())
	}
	err = model.CacheUpdateUserQuota(ctx, userId)
	if err != nil {
		logger.SysError("error update user quota cache: " + err.Error())
	}

	// totalQuota is total quota consumed
	// Always log the request for tracking purposes, regardless of quota amount
	var logContent string
	if toolsCost == 0 {
		logContent = fmt.Sprintf("model rate %.2f, group rate %.2f, completion rate %.2f", modelRatio, groupRatio, completionRatio)
	} else {
		logContent = fmt.Sprintf("model rate %.2f, group rate %.2f, completion rate %.2f, tools cost %d", modelRatio, groupRatio, completionRatio, toolsCost)
	}
	model.RecordConsumeLog(ctx, &model.Log{
		UserId:            userId,
		ChannelId:         channelId,
		PromptTokens:      promptTokens,
		CompletionTokens:  completionTokens,
		ModelName:         modelName,
		TokenName:         tokenName,
		Quota:             int(totalQuota),
		Content:           logContent,
		IsStream:          isStream,
		ElapsedTime:       helper.CalcElapsedTime(startTime),
		SystemPromptReset: systemPromptReset,
	})

	// Only update quotas when totalQuota > 0
	if totalQuota > 0 {
		model.UpdateUserUsedQuotaAndRequestCount(userId, totalQuota)
		model.UpdateChannelUsedQuota(channelId, totalQuota)
	}
	if totalQuota <= 0 {
		logger.Error(ctx, fmt.Sprintf("totalQuota consumed is %d, something is wrong", totalQuota))
	}
}
