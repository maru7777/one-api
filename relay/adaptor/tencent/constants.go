package tencent

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Tencent pricing: https://cloud.tencent.com/document/product/1729/97731
var ModelRatios = map[string]adaptor.ModelConfig{
	// Hunyuan Models - Based on https://cloud.tencent.com/document/product/1729/97731
	"hunyuan-lite":          {Ratio: 0.75 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-standard":      {Ratio: 4.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-standard-256K": {Ratio: 15 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-pro":           {Ratio: 30 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-vision":        {Ratio: 18 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-embedding":     {Ratio: 0.7 * ratio.MilliTokensRmb, CompletionRatio: 1},
}
