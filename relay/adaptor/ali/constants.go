package ali

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
var ModelRatios = map[string]adaptor.ModelConfig{
	// Qwen Turbo Models
	"qwen-turbo":        {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-turbo-latest": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen Plus Models
	"qwen-plus":        {Ratio: 0.8 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-plus-latest": {Ratio: 0.8 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen Max Models
	"qwen-max":             {Ratio: 2.4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-max-latest":      {Ratio: 2.4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-max-longcontext": {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen Vision Models
	"qwen-vl-max":         {Ratio: 3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-vl-max-latest":  {Ratio: 3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-vl-plus":        {Ratio: 1.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-vl-plus-latest": {Ratio: 1.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-vl-ocr":         {Ratio: 5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-vl-ocr-latest":  {Ratio: 5 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen Audio Models
	"qwen-audio-turbo": {Ratio: 1.4286, CompletionRatio: 1},

	// Qwen Math Models
	"qwen-math-plus":         {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-math-plus-latest":  {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-math-turbo":        {Ratio: 2 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-math-turbo-latest": {Ratio: 2 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen Coder Models
	"qwen-coder-plus":         {Ratio: 3.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-coder-plus-latest":  {Ratio: 3.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-coder-turbo":        {Ratio: 2 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-coder-turbo-latest": {Ratio: 2 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen MT Models
	"qwen-mt-plus":  {Ratio: 15 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-mt-turbo": {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// QwQ Models
	"qwq-32b-preview": {Ratio: 2 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen 2.5 Models
	"qwen2.5-72b-instruct":  {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-32b-instruct":  {Ratio: 30 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-14b-instruct":  {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-7b-instruct":   {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-3b-instruct":   {Ratio: 6 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-1.5b-instruct": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-0.5b-instruct": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen 2 Models
	"qwen2-72b-instruct":      {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2-57b-a14b-instruct": {Ratio: 3.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2-7b-instruct":       {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2-1.5b-instruct":     {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2-0.5b-instruct":     {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen 1.5 Models
	"qwen1.5-110b-chat": {Ratio: 8 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen1.5-72b-chat":  {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen1.5-32b-chat":  {Ratio: 2 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen1.5-14b-chat":  {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen1.5-7b-chat":   {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen1.5-1.8b-chat": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen1.5-0.5b-chat": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen 1 Models
	"qwen-72b-chat":              {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-14b-chat":              {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-7b-chat":               {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-1.8b-chat":             {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-1.8b-longcontext-chat": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// QVQ Models
	"qvq-72b-preview": {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen 2.5 VL Models
	"qwen2.5-vl-72b-instruct":  {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-vl-7b-instruct":   {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-vl-2b-instruct":   {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-vl-1b-instruct":   {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-vl-0.5b-instruct": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen 2 VL Models
	"qwen2-vl-7b-instruct": {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2-vl-2b-instruct": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-vl-v1":           {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-vl-chat-v1":      {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen Audio Models
	"qwen2-audio-instruct": {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen-audio-chat":      {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen Math Models (additional)
	"qwen2.5-math-72b-instruct":  {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-math-7b-instruct":   {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-math-1.5b-instruct": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2-math-72b-instruct":    {Ratio: 4 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2-math-7b-instruct":     {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2-math-1.5b-instruct":   {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Qwen Coder Models (additional)
	"qwen2.5-coder-32b-instruct":  {Ratio: 2 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-coder-14b-instruct":  {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-coder-7b-instruct":   {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-coder-3b-instruct":   {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-coder-1.5b-instruct": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"qwen2.5-coder-0.5b-instruct": {Ratio: 0.3 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// DeepSeek Models (hosted on Ali)
	"deepseek-r1":                   {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 8},
	"deepseek-v3":                   {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 2},
	"deepseek-r1-distill-qwen-1.5b": {Ratio: 0.07 * ratio.MilliTokensRmb, CompletionRatio: 0.28},
	"deepseek-r1-distill-qwen-7b":   {Ratio: 0.14 * ratio.MilliTokensRmb, CompletionRatio: 0.28},
	"deepseek-r1-distill-qwen-14b":  {Ratio: 0.28 * ratio.MilliTokensRmb, CompletionRatio: 0.28},
	"deepseek-r1-distill-qwen-32b":  {Ratio: 0.42 * ratio.MilliTokensRmb, CompletionRatio: 0.28},
	"deepseek-r1-distill-llama-8b":  {Ratio: 0.14 * ratio.MilliTokensRmb, CompletionRatio: 0.28},
	"deepseek-r1-distill-llama-70b": {Ratio: 1 * ratio.MilliTokensRmb, CompletionRatio: 2},

	// Embedding Models
	"text-embedding-v1":       {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"text-embedding-v3":       {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"text-embedding-v2":       {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"text-embedding-async-v2": {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"text-embedding-async-v1": {Ratio: 0.5 * ratio.MilliTokensRmb, CompletionRatio: 1},

	// Image Generation Models
	"ali-stable-diffusion-xl":   {Ratio: 8 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"ali-stable-diffusion-v1.5": {Ratio: 8 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"wanx-v1":                   {Ratio: 8 * ratio.MilliTokensRmb, CompletionRatio: 1},
}
