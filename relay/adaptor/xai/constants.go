package xai

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on X.AI pricing: https://console.x.ai/
var ModelRatios = map[string]adaptor.ModelConfig{
	// Grok Models - Based on https://console.x.ai/
	"grok-2":               {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"grok-2-latest":        {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"grok-2-1212":          {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"grok-vision-beta":     {Ratio: 7.5 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"grok-2-vision-1212":   {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"grok-2-vision":        {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"grok-2-vision-latest": {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"grok-beta":            {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 1},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
