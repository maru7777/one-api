package lingyiwanwu

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on LingYi WanWu pricing: https://platform.lingyiwanwu.com/docs
var ModelRatios = map[string]adaptor.ModelConfig{
	// LingYi WanWu Models - Based on https://platform.lingyiwanwu.com/docs
	"yi-34b-chat-0205": {Ratio: 2.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"yi-34b-chat-200k": {Ratio: 12.0 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"yi-vl-plus":       {Ratio: 6.0 * ratio.MilliTokensRmb, CompletionRatio: 1},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
