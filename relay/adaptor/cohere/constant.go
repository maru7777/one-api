package cohere

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
var ModelRatios = map[string]adaptor.ModelPrice{
	// Command Models
	"command":         {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 2}, // $15/$30 per 1M tokens
	"command-nightly": {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 2}, // $15/$30 per 1M tokens

	// Command Light Models
	"command-light":         {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2}, // $0.3/$0.6 per 1M tokens
	"command-light-nightly": {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2}, // $0.3/$0.6 per 1M tokens

	// Command R Models
	"command-r":      {Ratio: 0.5 * ratio.MilliTokensUsd, CompletionRatio: 3}, // $0.5/$1.5 per 1M tokens
	"command-r-plus": {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},   // $3/$15 per 1M tokens

	// Internet-enabled variants
	"command-internet":               {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 2},  // $15/$30 per 1M tokens
	"command-nightly-internet":       {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 2},  // $15/$30 per 1M tokens
	"command-light-internet":         {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2}, // $0.3/$0.6 per 1M tokens
	"command-light-nightly-internet": {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2}, // $0.3/$0.6 per 1M tokens
	"command-r-internet":             {Ratio: 0.5 * ratio.MilliTokensUsd, CompletionRatio: 3}, // $0.5/$1.5 per 1M tokens
	"command-r-plus-internet":        {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5},   // $3/$15 per 1M tokens
}
