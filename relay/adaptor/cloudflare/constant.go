package cloudflare

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Cloudflare Workers AI pricing - most models are free or very low cost
var ModelRatios = map[string]adaptor.ModelConfig{
	// Meta Llama Models
	"@cf/meta/llama-3.1-8b-instruct":         {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/meta/llama-2-7b-chat-fp16":          {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/meta/llama-2-7b-chat-int8":          {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/meta/llama-3-8b-instruct":           {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/meta-llama/llama-2-7b-chat-hf-lora": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},

	// Mistral Models
	"@cf/mistral/mistral-7b-instruct-v0.1":      {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/mistral/mistral-7b-instruct-v0.2-lora": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},

	// DeepSeek Models
	"@hf/thebloke/deepseek-coder-6.7b-base-awq":     {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/thebloke/deepseek-coder-6.7b-instruct-awq": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/deepseek-ai/deepseek-math-7b-base":         {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/deepseek-ai/deepseek-math-7b-instruct":     {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},

	// Other Models
	"@cf/thebloke/discolm-german-7b-v1-awq":      {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/tiiuae/falcon-7b-instruct":              {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/google/gemma-2b-it-lora":                {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/google/gemma-7b-it":                     {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/google/gemma-7b-it-lora":                {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/nousresearch/hermes-2-pro-mistral-7b":   {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/thebloke/llama-2-13b-chat-awq":          {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/thebloke/llamaguard-7b-awq":             {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/thebloke/mistral-7b-instruct-v0.1-awq":  {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/mistralai/mistral-7b-instruct-v0.2":     {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/thebloke/neural-chat-7b-v3-1-awq":       {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/openchat/openchat-3.5-0106":             {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/thebloke/openhermes-2.5-mistral-7b-awq": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/microsoft/phi-2":                        {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},

	// Qwen Models
	"@cf/qwen/qwen1.5-0.5b-chat":    {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/qwen/qwen1.5-1.8b-chat":    {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/qwen/qwen1.5-14b-chat-awq": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/qwen/qwen1.5-7b-chat-awq":  {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},

	// Specialized Models
	"@cf/defog/sqlcoder-7b-2":                {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/nexusflow/starling-lm-7b-beta":      {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@cf/tinyllama/tinyllama-1.1b-chat-v1.0": {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"@hf/thebloke/zephyr-7b-beta-awq":        {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
