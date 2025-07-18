package mistral

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Mistral pricing: https://docs.mistral.ai/platform/pricing/
var ModelRatios = map[string]adaptor.ModelConfig{
	// Open Models
	"open-mistral-7b":   {Ratio: 0.25 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"open-mixtral-8x7b": {Ratio: 0.7 * ratio.MilliTokensUsd, CompletionRatio: 1},

	// Mistral Models
	"mistral-small-latest":  {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"mistral-medium-latest": {Ratio: 2.7 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"mistral-large-latest":  {Ratio: 8.0 * ratio.MilliTokensUsd, CompletionRatio: 1},

	// Embedding Models
	"mistral-embed": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
