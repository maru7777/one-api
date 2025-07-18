package replicate

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Replicate pricing: https://replicate.com/pricing
var ModelRatios = map[string]adaptor.ModelConfig{
	// -------------------------------------
	// Image Generation Models
	// -------------------------------------
	"black-forest-labs/flux-kontext-pro":            {Ratio: 40.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.04 per image
	"black-forest-labs/flux-1.1-pro":                {Ratio: 40.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.04 per image
	"black-forest-labs/flux-1.1-pro-ultra":          {Ratio: 60.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.06 per image
	"black-forest-labs/flux-canny-dev":              {Ratio: 25.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.025 per image
	"black-forest-labs/flux-canny-pro":              {Ratio: 50.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.05 per image
	"black-forest-labs/flux-depth-dev":              {Ratio: 25.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.025 per image
	"black-forest-labs/flux-depth-pro":              {Ratio: 50.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.05 per image
	"black-forest-labs/flux-dev":                    {Ratio: 25.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.025 per image
	"black-forest-labs/flux-dev-lora":               {Ratio: 32.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.032 per image
	"black-forest-labs/flux-fill-dev":               {Ratio: 40.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.04 per image
	"black-forest-labs/flux-fill-pro":               {Ratio: 50.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.05 per image
	"black-forest-labs/flux-pro":                    {Ratio: 55.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.055 per image
	"black-forest-labs/flux-redux-dev":              {Ratio: 25.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.025 per image
	"black-forest-labs/flux-redux-schnell":          {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},  // $0.003 per image
	"black-forest-labs/flux-schnell":                {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},  // $0.003 per image
	"black-forest-labs/flux-schnell-lora":           {Ratio: 20.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.02 per image
	"ideogram-ai/ideogram-v2":                       {Ratio: 80.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.08 per image
	"ideogram-ai/ideogram-v2-turbo":                 {Ratio: 50.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.05 per image
	"recraft-ai/recraft-v3":                         {Ratio: 40.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.04 per image
	"recraft-ai/recraft-v3-svg":                     {Ratio: 80.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.08 per image
	"stability-ai/stable-diffusion-3":               {Ratio: 35.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.035 per image
	"stability-ai/stable-diffusion-3.5-large":       {Ratio: 65.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.065 per image
	"stability-ai/stable-diffusion-3.5-large-turbo": {Ratio: 40.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.04 per image
	"stability-ai/stable-diffusion-3.5-medium":      {Ratio: 35.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.035 per image

	// -------------------------------------
	// Language Models
	// -------------------------------------
	"anthropic/claude-3.5-haiku":                {Ratio: 1.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $1.0/$5.0 per 1M tokens
	"anthropic/claude-3.5-sonnet":               {Ratio: 3.75 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $3.75/$18.75 per 1M tokens
	"anthropic/claude-3.7-sonnet":               {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $3.0/$15.0 per 1M tokens
	"deepseek-ai/deepseek-r1":                   {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},  // $10.0 per 1M tokens
	"ibm-granite/granite-20b-code-instruct-8k":  {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $0.1/$0.5 per 1M tokens
	"ibm-granite/granite-3.0-2b-instruct":       {Ratio: 0.03 * ratio.MilliTokensUsd, CompletionRatio: 8.33}, // $0.03/$0.25 per 1M tokens
	"ibm-granite/granite-3.0-8b-instruct":       {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"ibm-granite/granite-3.1-2b-instruct":       {Ratio: 0.03 * ratio.MilliTokensUsd, CompletionRatio: 8.33}, // $0.03/$0.25 per 1M tokens
	"ibm-granite/granite-3.1-8b-instruct":       {Ratio: 0.03 * ratio.MilliTokensUsd, CompletionRatio: 8.33}, // $0.03/$0.25 per 1M tokens
	"ibm-granite/granite-3.2-8b-instruct":       {Ratio: 0.03 * ratio.MilliTokensUsd, CompletionRatio: 8.33}, // $0.03/$0.25 per 1M tokens
	"ibm-granite/granite-8b-code-instruct-128k": {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"meta/llama-2-13b":                          {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $0.1/$0.5 per 1M tokens
	"meta/llama-2-13b-chat":                     {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $0.1/$0.5 per 1M tokens
	"meta/llama-2-70b":                          {Ratio: 0.65 * ratio.MilliTokensUsd, CompletionRatio: 4.23}, // $0.65/$2.75 per 1M tokens
	"meta/llama-2-70b-chat":                     {Ratio: 0.65 * ratio.MilliTokensUsd, CompletionRatio: 4.23}, // $0.65/$2.75 per 1M tokens
	"meta/llama-2-7b":                           {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"meta/llama-2-7b-chat":                      {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"meta/meta-llama-3.1-405b-instruct":         {Ratio: 9.5 * ratio.MilliTokensUsd, CompletionRatio: 1.0},   // $9.5 per 1M tokens
	"meta/meta-llama-3-70b":                     {Ratio: 0.65 * ratio.MilliTokensUsd, CompletionRatio: 4.23}, // $0.65/$2.75 per 1M tokens
	"meta/meta-llama-3-70b-instruct":            {Ratio: 0.65 * ratio.MilliTokensUsd, CompletionRatio: 4.23}, // $0.65/$2.75 per 1M tokens
	"meta/meta-llama-3-8b":                      {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"meta/meta-llama-3-8b-instruct":             {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"mistralai/mistral-7b-instruct-v0.2":        {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"mistralai/mistral-7b-v0.1":                 {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens

	// -------------------------------------
	// Video Models (TODO: implement the adaptor)
	// -------------------------------------
	// "minimax/video-01": {Ratio: 1.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
}

// 模型映射关系
var ReplicateModelMapping = map[string]string{
	"flux-1.1-pro":                     "black-forest-labs/flux-1.1-pro",
	"flux-1.1-pro-ultra":               "black-forest-labs/flux-1.1-pro-ultra",
	"flux-canny-dev":                   "black-forest-labs/flux-canny-dev",
	"flux-canny-pro":                   "black-forest-labs/flux-canny-pro",
	"flux-depth-dev":                   "black-forest-labs/flux-depth-dev",
	"flux-depth-pro":                   "black-forest-labs/flux-depth-pro",
	"flux-dev":                         "black-forest-labs/flux-dev",
	"flux-dev-lora":                    "black-forest-labs/flux-dev-lora",
	"flux-fill-dev":                    "black-forest-labs/flux-fill-dev",
	"flux-fill-pro":                    "black-forest-labs/flux-fill-pro",
	"flux-pro":                         "black-forest-labs/flux-pro",
	"flux-redux-dev":                   "black-forest-labs/flux-redux-dev",
	"flux-redux-schnell":               "black-forest-labs/flux-redux-schnell",
	"flux-schnell":                     "black-forest-labs/flux-schnell",
	"flux-schnell-lora":                "black-forest-labs/flux-schnell-lora",
	"imagen-3":                         "google/imagen-3",
	"imagen-3-fast":                    "google/imagen-3-fast",
	"upscaler":                         "google/upscaler",
	"ideogram-v2":                      "ideogram-ai/ideogram-v2",
	"ideogram-v2a":                     "ideogram-ai/ideogram-v2a",
	"ideogram-v2a-turbo":               "ideogram-ai/ideogram-v2a-turbo",
	"ideogram-v2-turbo":                "ideogram-ai/ideogram-v2-turbo",
	"photon":                           "luma/photon",
	"photon-flash":                     "luma/photon-flash",
	"image-01":                         "minimax/image-01",
	"recraft-20b":                      "recraft-ai/recraft-20b",
	"recraft-20b-svg":                  "recraft-ai/recraft-20b-svg",
	"recraft-creative-upscale":         "recraft-ai/recraft-creative-upscale",
	"recraft-crisp-upscale":            "recraft-ai/recraft-crisp-upscale",
	"recraft-v3":                       "recraft-ai/recraft-v3",
	"recraft-v3-svg":                   "recraft-ai/recraft-v3-svg",
	"stable-diffusion-3":               "stability-ai/stable-diffusion-3",
	"stable-diffusion-3.5-large":       "stability-ai/stable-diffusion-3.5-large",
	"stable-diffusion-3.5-large-turbo": "stability-ai/stable-diffusion-3.5-large-turbo",
	"stable-diffusion-3.5-medium":      "stability-ai/stable-diffusion-3.5-medium",
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
