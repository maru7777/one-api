package moonshot

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Moonshot pricing: https://platform.moonshot.cn/docs/pricing
var ModelRatios = map[string]adaptor.ModelConfig{
	// Moonshot Models - Based on https://platform.moonshot.cn/docs/pricing
	"moonshot-v1-8k":   {Ratio: 12 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"moonshot-v1-32k":  {Ratio: 24 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"moonshot-v1-128k": {Ratio: 60 * ratio.MilliTokensRmb, CompletionRatio: 1},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
