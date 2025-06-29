package groq

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Groq pricing: https://groq.com/pricing/
var ModelRatios = map[string]adaptor.ModelPrice{
	// Regular Models
	"distil-whisper-large-v3-en": {Ratio: 0.111 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"gemma2-9b-it":               {Ratio: 0.20 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama-3.3-70b-versatile":    {Ratio: 0.59 * ratio.MilliTokensUsd, CompletionRatio: 0.79 / 0.59},
	"llama-3.1-8b-instant":       {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 0.08 / 0.05},
	"llama-guard-3-8b":           {Ratio: 0.20 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama3-70b-8192":            {Ratio: 0.59 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama3-8b-8192":             {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"mixtral-8x7b-32768":         {Ratio: 0.24 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"whisper-large-v3":           {Ratio: 0.111 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"whisper-large-v3-turbo":     {Ratio: 0.04 * ratio.MilliTokensUsd, CompletionRatio: 1},

	// Preview Models
	"qwen-qwq-32b":                          {Ratio: 0.29 * ratio.MilliTokensUsd, CompletionRatio: 0.39 / 0.29},
	"mistral-saba-24b":                      {Ratio: 0.79 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"qwen-2.5-coder-32b":                    {Ratio: 0.79 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"qwen-2.5-32b":                          {Ratio: 0.79 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"deepseek-r1-distill-qwen-32b":          {Ratio: 0.29 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"deepseek-r1-distill-llama-70b-specdec": {Ratio: 0.75 * ratio.MilliTokensUsd, CompletionRatio: 0.99 / 0.75},
	"deepseek-r1-distill-llama-70b":         {Ratio: 0.75 * ratio.MilliTokensUsd, CompletionRatio: 0.99 / 0.75},
	"llama-3.2-1b-preview":                  {Ratio: 0.04 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama-3.2-3b-preview":                  {Ratio: 0.06 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama-3.2-11b-vision-preview":          {Ratio: 0.18 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama-3.2-90b-vision-preview":          {Ratio: 0.90 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"llama-3.3-70b-specdec":                 {Ratio: 0.59 * ratio.MilliTokensUsd, CompletionRatio: 0.99 / 0.59},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
