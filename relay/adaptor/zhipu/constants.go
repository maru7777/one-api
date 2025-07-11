package zhipu

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Zhipu pricing: https://open.bigmodel.cn/pricing
var ModelRatios = map[string]adaptor.ModelPrice{
	// GLM Zero Models
	"glm-zero-preview": {Ratio: 0.7 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// GLM-4 Models
	"glm-4-plus":   {Ratio: 0.1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4-0520":   {Ratio: 0.1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4":        {Ratio: 0.1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4-airx":   {Ratio: 0.01 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4-air":    {Ratio: 0.001 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4-long":   {Ratio: 0.001 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4-flashx": {Ratio: 0.001 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4-flash":  {Ratio: 0.001 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// GLM-3 Models
	"glm-3-turbo": {Ratio: 0.005 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// GLM Vision Models
	"glm-4v-plus":  {Ratio: 0.1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4v":       {Ratio: 0.05 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"glm-4v-flash": {Ratio: 0.001 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// CogView Image Models
	"cogview-3-plus":  {Ratio: 0.08 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"cogview-3":       {Ratio: 0.04 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"cogview-3-flash": {Ratio: 0.008 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"cogviewx":        {Ratio: 0.04 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"cogviewx-flash":  {Ratio: 0.008 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Character and Code Models
	"charglm-4":  {Ratio: 0.1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"emohaa":     {Ratio: 0.1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"codegeex-4": {Ratio: 0.001 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Embedding Models
	"embedding-3": {Ratio: 0.0005 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"embedding-2": {Ratio: 0.0005 * ratio.MilliTokensRmb, CompletionRatio: 1},
}
