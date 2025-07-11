package ai360

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
var ModelRatios = map[string]adaptor.ModelPrice{
	// AI360 Models - Based on historical pricing
	"360GPT_S2_V9":              {Ratio: 0.8572 * ratio.MilliTokensUsd, CompletionRatio: 1}, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":     {Ratio: 0.0715 * ratio.MilliTokensUsd, CompletionRatio: 1}, // ¥0.001 / 1k tokens
	"embedding_s1_v1":           {Ratio: 0.0715 * ratio.MilliTokensUsd, CompletionRatio: 1}, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1": {Ratio: 0.0715 * ratio.MilliTokensUsd, CompletionRatio: 1}, // ¥0.001 / 1k tokens
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
