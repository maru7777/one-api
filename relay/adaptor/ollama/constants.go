package ollama

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Ollama is typically free for local usage, but we set minimal pricing for consistency
var ModelRatios = map[string]adaptor.ModelPrice{
	// Ollama Models - typically free for local usage
	"codellama:7b-instruct": {Ratio: 0.01 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama2:7b":             {Ratio: 0.01 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama2:latest":         {Ratio: 0.01 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama3:latest":         {Ratio: 0.01 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"phi3:latest":           {Ratio: 0.01 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"qwen:0.5b-chat":        {Ratio: 0.005 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"qwen:7b":               {Ratio: 0.01 * ratio.MilliTokensUsd, CompletionRatio: 1},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
