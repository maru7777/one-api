package togetherai

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Together AI pricing: https://docs.together.ai/docs/inference-models
var ModelRatios = map[string]adaptor.ModelPrice{
	// Together AI Models - Based on https://docs.together.ai/docs/inference-models
	"meta-llama/Llama-3-70b-chat-hf":          {Ratio: 0.9 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"deepseek-ai/deepseek-coder-33b-instruct": {Ratio: 0.8 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"mistralai/Mixtral-8x22B-Instruct-v0.1":   {Ratio: 1.2 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"Qwen/Qwen1.5-72B-Chat":                   {Ratio: 0.9 * ratio.MilliTokensUsd, CompletionRatio: 1},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
