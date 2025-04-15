package ratio

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/songquanpeng/one-api/common/logger"
)

// Constants defining currency conversion and token pricing
const (
	USD2RMB float64 = 7.3
	// QuotaPerUsd is the number of tokens per USD
	QuotaPerUsd float64 = 500000 // $0.002 / 1K tokens
	// KiloTokensUsd multiply by the USD price per 1,000 tokens to get the quota cost per token
	KiloTokensUsd float64 = QuotaPerUsd / 1000
	// MilliTokensUsd multiply by the USD price per 1 million tokens to get the quota cost per token
	MilliTokensUsd float64 = KiloTokensUsd / 1000
	// KiloRmb multiply by the RMB price per 1,000 tokens to get the quota cost per token
	KiloRmb float64 = KiloTokensUsd / USD2RMB
	// MilliRmb multiply by the RMB price per 1 million tokens to get the quota cost per token
	MilliRmb       float64 = MilliTokensUsd / USD2RMB
	ImageUsdPerPic float64 = QuotaPerUsd / 1000
)

var modelRatioLock sync.RWMutex

// ModelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
var ModelRatio = map[string]float64{
	// -------------------------------------
	// OpenAI
	// https://openai.com/pricing
	// -------------------------------------
	"gpt-4.5-preview":            75 * MilliTokensUsd,
	"gpt-4.5-preview-2025-02-27": 75 * MilliTokensUsd,
	"gpt-4":                      30 * MilliTokensUsd,
	"gpt-4-0314":                 30 * MilliTokensUsd,
	"gpt-4-0613":                 30 * MilliTokensUsd,
	"gpt-4-32k":                  60 * MilliTokensUsd,
	"gpt-4-32k-0314":             60 * MilliTokensUsd,
	"gpt-4-32k-0613":             60 * MilliTokensUsd,
	"gpt-4-1106-preview":         10 * MilliTokensUsd,
	"gpt-4-0125-preview":         10 * MilliTokensUsd,
	"gpt-4-turbo-preview":        10 * MilliTokensUsd,
	"gpt-4-turbo":                10 * MilliTokensUsd,
	"gpt-4-turbo-2024-04-09":     10 * MilliTokensUsd,
	"gpt-4o":                     2.5 * MilliTokensUsd,
	"chatgpt-4o-latest":          5 * MilliTokensUsd,
	"gpt-4o-2024-05-13":          5 * MilliTokensUsd,
	"gpt-4o-2024-08-06":          2.5 * MilliTokensUsd,
	"gpt-4o-2024-11-20":          2.5 * MilliTokensUsd,
	"gpt-4o-search-preview":      5 * MilliTokensUsd,
	"gpt-4o-mini":                0.15 * MilliTokensUsd,
	"gpt-4o-mini-2024-07-18":     0.15 * MilliTokensUsd,
	"gpt-4o-mini-search-preview": 0.15 * MilliTokensUsd,
	"gpt-4-vision-preview":       10 * MilliTokensUsd,
	// Audio billing will mix text and audio tokens, the unit price is different.
	// Here records the cost of text, the cost multiplier of audio
	// relative to text is in AudioRatio
	"gpt-4o-audio-preview":                 2.5 * MilliTokensUsd,
	"gpt-4o-audio-preview-2024-12-17":      2.5 * MilliTokensUsd,
	"gpt-4o-audio-preview-2024-10-01":      2.5 * MilliTokensUsd,
	"gpt-4o-mini-audio-preview":            0.15 * MilliTokensUsd,
	"gpt-4o-mini-audio-preview-2024-12-17": 0.15 * MilliTokensUsd,
	"gpt-3.5-turbo":                        0.5 * MilliTokensUsd,
	"gpt-3.5-turbo-0301":                   1.5 * MilliTokensUsd,
	"gpt-3.5-turbo-0613":                   1.5 * MilliTokensUsd,
	"gpt-3.5-turbo-16k":                    3 * MilliTokensUsd,
	"gpt-3.5-turbo-16k-0613":               3 * MilliTokensUsd,
	"gpt-3.5-turbo-instruct":               1.5 * MilliTokensUsd,
	"gpt-3.5-turbo-1106":                   1 * MilliTokensUsd,
	"gpt-3.5-turbo-0125":                   0.5 * MilliTokensUsd,
	"o1-pro":                               150 * MilliTokensUsd,
	"o1-pro-2025-03-19":                    150 * MilliTokensUsd,
	"o1":                                   15 * MilliTokensUsd,
	"o1-2024-12-17":                        15 * MilliTokensUsd,
	"o1-preview":                           15 * MilliTokensUsd,
	"o1-preview-2024-09-12":                15 * MilliTokensUsd,
	"o1-mini":                              1.1 * MilliTokensUsd,
	"o1-mini-2024-09-12":                   1.1 * MilliTokensUsd,
	"o3-mini":                              1.1 * MilliTokensUsd,
	"o3-mini-2025-01-31":                   1.1 * MilliTokensUsd,
	"davinci-002":                          2 * MilliTokensUsd,
	"babbage-002":                          0.4 * MilliTokensUsd,
	"text-ada-001":                         0.4 * MilliTokensUsd,
	"text-babbage-001":                     0.5 * MilliTokensUsd,
	"text-curie-001":                       2 * MilliTokensUsd,
	"text-davinci-002":                     20 * MilliTokensUsd,
	"text-davinci-003":                     20 * MilliTokensUsd,
	"text-davinci-edit-001":                20 * MilliTokensUsd,
	"code-davinci-edit-001":                20 * MilliTokensUsd,
	"whisper-1":                            30 * MilliTokensUsd,
	"gpt-4o-transcribe":                    2.5 * MilliTokensUsd,
	"gpt-4o-mini-transcribe":               1.25 * MilliTokensUsd,
	"gpt-4o-mini-tts":                      0.6 * MilliTokensUsd,
	"tts-1":                                15 * MilliTokensUsd,
	"tts-1-1106":                           15 * MilliTokensUsd,
	"tts-1-hd":                             30 * MilliTokensUsd,
	"tts-1-hd-1106":                        30 * MilliTokensUsd,
	"davinci":                              20 * MilliTokensUsd,
	"curie":                                20 * MilliTokensUsd,
	"babbage":                              20 * MilliTokensUsd,
	"ada":                                  20 * MilliTokensUsd,
	"text-embedding-ada-002":               0.1 * MilliTokensUsd,
	"text-embedding-3-small":               0.02 * MilliTokensUsd,
	"text-embedding-3-large":               0.13 * MilliTokensUsd,
	"text-search-ada-doc-001":              20 * MilliTokensUsd,
	"text-moderation-stable":               0.2 * MilliTokensUsd,
	"text-moderation-latest":               0.2 * MilliTokensUsd,
	"dall-e-2":                             0.02 * ImageUsdPerPic,
	"dall-e-3":                             0.04 * ImageUsdPerPic,
	// https://www.anthropic.com/api#pricing
	"claude-instant-1.2":         0.8 * MilliTokensUsd,
	"claude-2.0":                 8.0 * MilliTokensUsd,
	"claude-2.1":                 8.0 * MilliTokensUsd,
	"claude-3-haiku-20240307":    0.25 * MilliTokensUsd,
	"claude-3-5-haiku-20241022":  1.0 * MilliTokensUsd,
	"claude-3-5-haiku-latest":    1.0 * MilliTokensUsd,
	"claude-3-sonnet-20240229":   3.0 * MilliTokensUsd,
	"claude-3-5-sonnet-20240620": 3.0 * MilliTokensUsd,
	"claude-3-5-sonnet-20241022": 3.0 * MilliTokensUsd,
	"claude-3-5-sonnet-latest":   3.0 * MilliTokensUsd,
	"claude-3-7-sonnet-20250219": 3.0 * MilliTokensUsd,
	"claude-3-7-sonnet-latest":   3.0 * MilliTokensUsd,
	"claude-3-opus-20240229":     15.0 * MilliTokensUsd,
	// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/hlrk4akp7
	"ERNIE-4.0-8K":       0.120 * KiloRmb,
	"ERNIE-3.5-8K":       0.012 * KiloRmb,
	"ERNIE-3.5-8K-0205":  0.024 * KiloRmb,
	"ERNIE-3.5-8K-1222":  0.012 * KiloRmb,
	"ERNIE-Bot-8K":       0.024 * KiloRmb,
	"ERNIE-3.5-4K-0205":  0.012 * KiloRmb,
	"ERNIE-Speed-8K":     0.004 * KiloRmb,
	"ERNIE-Speed-128K":   0.004 * KiloRmb,
	"ERNIE-Lite-8K-0922": 0.008 * KiloRmb,
	"ERNIE-Lite-8K-0308": 0.003 * KiloRmb,
	"ERNIE-Tiny-8K":      0.001 * KiloRmb,
	"BLOOMZ-7B":          0.004 * KiloRmb,
	"Embedding-V1":       0.002 * KiloRmb,
	"bge-large-zh":       0.002 * KiloRmb,
	"bge-large-en":       0.002 * KiloRmb,
	"tao-8k":             0.002 * KiloRmb,
	// https://ai.google.dev/pricing
	// https://cloud.google.com/vertex-ai/generative-ai/pricing
	"gemma-2-2b-it":                         0,
	"gemma-2-9b-it":                         0,
	"gemma-2-27b-it":                        0,
	"gemma-3-27b-it":                        0,
	"gemini-pro":                            0.25 * MilliTokensUsd, // $0.00025 / 1k characters -> $0.001 / 1k tokens
	"gemini-1.0-pro":                        0.125 * MilliTokensUsd,
	"gemini-1.0-pro-vision":                 0.125 * MilliTokensUsd,
	"gemini-1.5-pro":                        1.25 * MilliTokensUsd,
	"gemini-1.5-pro-001":                    1.25 * MilliTokensUsd,
	"gemini-1.5-pro-002":                    1.25 * MilliTokensUsd,
	"gemini-1.5-pro-experimental":           1.25 * MilliTokensUsd,
	"gemini-1.5-flash":                      0.075 * MilliTokensUsd,
	"gemini-1.5-flash-001":                  0.075 * MilliTokensUsd,
	"gemini-1.5-flash-002":                  0.075 * MilliTokensUsd,
	"gemini-1.5-flash-8b":                   0.0375 * MilliTokensUsd,
	"gemini-2.0-flash":                      0.15 * MilliTokensUsd,
	"gemini-2.0-flash-exp":                  0.075 * MilliTokensUsd,
	"gemini-2.0-flash-001":                  0.15 * MilliTokensUsd,
	"gemini-2.0-flash-lite":                 0.075 * MilliTokensUsd,
	"gemini-2.0-flash-lite-001":             0.075 * MilliTokensUsd,
	"gemini-2.0-flash-lite-preview-02-05":   0.075 * MilliTokensUsd,
	"gemini-2.0-flash-thinking-exp-01-21":   0.075 * MilliTokensUsd,
	"gemini-2.0-flash-exp-image-generation": 0.075 * MilliTokensUsd,
	"gemini-2.0-pro-exp-02-05":              1.25 * MilliTokensUsd,
	"gemini-2.5-pro-exp-03-25":              1.25 * MilliTokensUsd,
	"aqa":                                   1,
	// https://open.bigmodel.cn/pricing
	"glm-zero-preview": 0.01 * KiloRmb,
	"glm-4-plus":       0.05 * KiloRmb,
	"glm-4-0520":       0.1 * KiloRmb,
	"glm-4-airx":       0.01 * KiloRmb,
	"glm-4-air":        0.0005 * KiloRmb,
	"glm-4-long":       0.001 * KiloRmb,
	"glm-4-flashx":     0.0001 * KiloRmb,
	"glm-4-flash":      0,
	"glm-4":            0.1 * KiloRmb,   // deprecated model, available until 2025/06
	"glm-3-turbo":      0.001 * KiloRmb, // deprecated model, available until 2025/06
	"glm-4v-plus":      0.004 * KiloRmb,
	"glm-4v":           0.05 * KiloRmb,
	"glm-4v-flash":     0,
	"cogview-3-plus":   0.06 * KiloRmb,
	"cogview-3":        0.1 * KiloRmb,
	"cogview-3-flash":  0,
	"cogviewx":         0.5 * KiloRmb,
	"cogviewx-flash":   0,
	"charglm-4":        0.001 * KiloRmb,
	"emohaa":           0.015 * KiloRmb,
	"codegeex-4":       0.0001 * KiloRmb,
	"embedding-2":      0.0005 * KiloRmb,
	"embedding-3":      0.0005 * KiloRmb,
	// https://help.aliyun.com/zh/dashscope/developer-reference/tongyi-thousand-questions-metering-and-billing
	"qwen-turbo":                  0.0003 * KiloRmb,
	"qwen-turbo-latest":           0.0003 * KiloRmb,
	"qwen-plus":                   0.0008 * KiloRmb,
	"qwen-plus-latest":            0.0008 * KiloRmb,
	"qwen-max":                    0.0024 * KiloRmb,
	"qwen-max-latest":             0.0024 * KiloRmb,
	"qwen-max-longcontext":        0.0005 * KiloRmb,
	"qwen-vl-max":                 0.003 * KiloRmb,
	"qwen-vl-max-latest":          0.003 * KiloRmb,
	"qwen-vl-plus":                0.0015 * KiloRmb,
	"qwen-vl-plus-latest":         0.0015 * KiloRmb,
	"qwen-vl-ocr":                 0.005 * KiloRmb,
	"qwen-vl-ocr-latest":          0.005 * KiloRmb,
	"qwen-audio-turbo":            1.4286,
	"qwen-math-plus":              0.004 * KiloRmb,
	"qwen-math-plus-latest":       0.004 * KiloRmb,
	"qwen-math-turbo":             0.002 * KiloRmb,
	"qwen-math-turbo-latest":      0.002 * KiloRmb,
	"qwen-coder-plus":             0.0035 * KiloRmb,
	"qwen-coder-plus-latest":      0.0035 * KiloRmb,
	"qwen-coder-turbo":            0.002 * KiloRmb,
	"qwen-coder-turbo-latest":     0.002 * KiloRmb,
	"qwen-mt-plus":                0.015 * KiloRmb,
	"qwen-mt-turbo":               0.001 * KiloRmb,
	"qwq-32b-preview":             0.002 * KiloRmb,
	"qwen2.5-72b-instruct":        0.004 * KiloRmb,
	"qwen2.5-32b-instruct":        0.03 * KiloRmb,
	"qwen2.5-14b-instruct":        0.001 * KiloRmb,
	"qwen2.5-7b-instruct":         0.0005 * KiloRmb,
	"qwen2.5-3b-instruct":         0.006 * KiloRmb,
	"qwen2.5-1.5b-instruct":       0.0003 * KiloRmb,
	"qwen2.5-0.5b-instruct":       0.0003 * KiloRmb,
	"qwen2-72b-instruct":          0.004 * KiloRmb,
	"qwen2-57b-a14b-instruct":     0.0035 * KiloRmb,
	"qwen2-7b-instruct":           0.001 * KiloRmb,
	"qwen2-1.5b-instruct":         0.001 * KiloRmb,
	"qwen2-0.5b-instruct":         0.001 * KiloRmb,
	"qwen1.5-110b-chat":           0.007 * KiloRmb,
	"qwen1.5-72b-chat":            0.005 * KiloRmb,
	"qwen1.5-32b-chat":            0.0035 * KiloRmb,
	"qwen1.5-14b-chat":            0.002 * KiloRmb,
	"qwen1.5-7b-chat":             0.001 * KiloRmb,
	"qwen1.5-1.8b-chat":           0.001 * KiloRmb,
	"qwen1.5-0.5b-chat":           0.001 * KiloRmb,
	"qwen-72b-chat":               0.02 * KiloRmb,
	"qwen-14b-chat":               0.008 * KiloRmb,
	"qwen-7b-chat":                0.006 * KiloRmb,
	"qwen-1.8b-chat":              0.006 * KiloRmb,
	"qwen-1.8b-longcontext-chat":  0.006 * KiloRmb,
	"qvq-72b-preview":             0.012 * KiloRmb,
	"qwen2.5-vl-72b-instruct":     0.016 * KiloRmb,
	"qwen2.5-vl-7b-instruct":      0.002 * KiloRmb,
	"qwen2.5-vl-3b-instruct":      0.0012 * KiloRmb,
	"qwen2-vl-7b-instruct":        0.016 * KiloRmb,
	"qwen2-vl-2b-instruct":        0.002 * KiloRmb,
	"qwen-vl-v1":                  0.002 * KiloRmb,
	"qwen-vl-chat-v1":             0.002 * KiloRmb,
	"qwen2-audio-instruct":        0.002 * KiloRmb,
	"qwen-audio-chat":             0.002 * KiloRmb,
	"qwen2.5-math-72b-instruct":   0.004 * KiloRmb,
	"qwen2.5-math-7b-instruct":    0.001 * KiloRmb,
	"qwen2.5-math-1.5b-instruct":  0.001 * KiloRmb,
	"qwen2-math-72b-instruct":     0.004 * KiloRmb,
	"qwen2-math-7b-instruct":      0.001 * KiloRmb,
	"qwen2-math-1.5b-instruct":    0.001 * KiloRmb,
	"qwen2.5-coder-32b-instruct":  0.002 * KiloRmb,
	"qwen2.5-coder-14b-instruct":  0.002 * KiloRmb,
	"qwen2.5-coder-7b-instruct":   0.001 * KiloRmb,
	"qwen2.5-coder-3b-instruct":   0.001 * KiloRmb,
	"qwen2.5-coder-1.5b-instruct": 0.001 * KiloRmb,
	"qwen2.5-coder-0.5b-instruct": 0.001 * KiloRmb,
	"text-embedding-v1":           0.0007 * KiloRmb, // ￥0.0007 / 1k tokens
	"text-embedding-v3":           0.0007 * KiloRmb,
	"text-embedding-v2":           0.0007 * KiloRmb,
	"text-embedding-async-v2":     0.0007 * KiloRmb,
	"text-embedding-async-v1":     0.0007 * KiloRmb,
	"ali-stable-diffusion-xl":     0.016 * ImageUsdPerPic,
	"ali-stable-diffusion-v1.5":   0.016 * ImageUsdPerPic,
	"wanx-v1":                     0.04 * ImageUsdPerPic, // 0.02192, 原作者的是 0.016 * ImageUsdPerPic
	// ali新增生图模型
	"wanx2.1-t2i-turbo":             0.04 * ImageUsdPerPic, // 0.01918
	"wanx2.1-t2i-plus":              0.04 * ImageUsdPerPic, // 0.02740
	"wanx2.0-t2i-turbo":             0.01 * ImageUsdPerPic, // 0.00548
	"wanx2.1-imageedit":             0.04 * ImageUsdPerPic, // 0.01918
	"wanx-sketch-to-image-lite":     0.01 * ImageUsdPerPic, // 0.00822
	"wanx-style-repaint-v1":         0.04 * ImageUsdPerPic, // 0.01644
	"wanx-background-generation-v2": 0.04 * ImageUsdPerPic, // 0.01096
	"aitryon":                       0.04 * ImageUsdPerPic, // 0.02740
	"aitryon-parsing-v1":            0.01 * ImageUsdPerPic, // 0.00055
	"aitryon-refiner":               0.05 * ImageUsdPerPic, // 0.04110
	"facechain-generation":          0.04 * ImageUsdPerPic, // 0.02466
	"wordart-texture":               0.04 * ImageUsdPerPic, // 0.01096
	"wordart-semantic":              0.04 * ImageUsdPerPic, // 0.03288
	"deepseek-r1":                   0.002 * KiloRmb,
	"deepseek-v3":                   0.001 * KiloRmb,
	"deepseek-r1-distill-qwen-1.5b": 0.001 * KiloRmb,
	"deepseek-r1-distill-qwen-7b":   0.0005 * KiloRmb,
	"deepseek-r1-distill-qwen-14b":  0.001 * KiloRmb,
	// "deepseek-r1-distill-qwen-32b":  0.002 * KiloRmb,
	"deepseek-r1-distill-llama-8b": 0.0005 * KiloRmb,
	// "deepseek-r1-distill-llama-70b": 0.004 * KiloRmb,
	"SparkDesk":                 1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v1.1":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1-128K":       1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5-32K":        1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":            1.2858, // ￥0.018 / 1k tokens
	"360GPT_S2_V9":              0.8572, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":     0.0715, // ¥0.001 / 1k tokens
	"embedding_s1_v1":           0.0715, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1": 0.0715, // ¥0.001 / 1k tokens
	// https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	"hunyuan-turbo":             0.015 * KiloRmb,
	"hunyuan-large":             0.004 * KiloRmb,
	"hunyuan-large-longcontext": 0.006 * KiloRmb,
	"hunyuan-standard":          0.0008 * KiloRmb,
	"hunyuan-standard-256K":     0.0005 * KiloRmb,
	"hunyuan-translation-lite":  0.005 * KiloRmb,
	"hunyuan-role":              0.004 * KiloRmb,
	"hunyuan-functioncall":      0.004 * KiloRmb,
	"hunyuan-code":              0.004 * KiloRmb,
	"hunyuan-turbo-vision":      0.08 * KiloRmb,
	"hunyuan-vision":            0.018 * KiloRmb,
	"hunyuan-embedding":         0.0007 * KiloRmb,
	// https://platform.moonshot.cn/pricing
	"moonshot-v1-8k":   0.012 * KiloRmb,
	"moonshot-v1-32k":  0.024 * KiloRmb,
	"moonshot-v1-128k": 0.06 * KiloRmb,
	// https://platform.baichuan-ai.com/price
	"Baichuan2-Turbo":      0.008 * KiloRmb,
	"Baichuan2-Turbo-192k": 0.016 * KiloRmb,
	"Baichuan2-53B":        0.02 * KiloRmb,
	// https://api.minimax.chat/document/price
	"abab6.5-chat":  0.03 * KiloRmb,
	"abab6.5s-chat": 0.01 * KiloRmb,
	"abab6-chat":    0.1 * KiloRmb,
	"abab5.5-chat":  0.015 * KiloRmb,
	"abab5.5s-chat": 0.005 * KiloRmb,
	// https://docs.mistral.ai/platform/pricing/
	"open-mistral-7b":       0.25 * MilliTokensUsd,
	"open-mixtral-8x7b":     0.7 * MilliTokensUsd,
	"mistral-small-latest":  2.0 * MilliTokensUsd,
	"mistral-medium-latest": 2.7 * MilliTokensUsd,
	"mistral-large-latest":  8.0 * MilliTokensUsd,
	"mistral-embed":         0.1 * MilliTokensUsd,
	// -------------------------------------
	// https://groq.com/pricing/
	// -------------------------------------
	"gemma2-9b-it":                          0.20 * MilliTokensUsd,
	"llama-3.1-8b-instant":                  0.05 * MilliTokensUsd,
	"llama-3.2-11b-text-preview":            0.18 * MilliTokensUsd,
	"llama-3.2-11b-vision-preview":          0.18 * MilliTokensUsd,
	"llama-3.2-1b-preview":                  0.04 * MilliTokensUsd,
	"llama-3.2-3b-preview":                  0.06 * MilliTokensUsd,
	"llama-3.2-90b-text-preview":            0.90 * MilliTokensUsd,
	"llama-3.2-90b-vision-preview":          0.90 * MilliTokensUsd,
	"llama-3.3-70b-versatile":               0.59 * MilliTokensUsd,
	"llama-guard-3-8b":                      0.20 * MilliTokensUsd,
	"llama3-70b-8192":                       0.59 * MilliTokensUsd,
	"llama3-8b-8192":                        0.05 * MilliTokensUsd,
	"llama3-groq-70b-8192-tool-use-preview": 0.59 * MilliTokensUsd,
	"llama3-groq-8b-8192-tool-use-preview":  0.05 * MilliTokensUsd,
	"llama-3.3-70b-specdec":                 0.59 * MilliTokensUsd,
	"mistral-saba-24b":                      0.79 * MilliTokensUsd,
	"qwen-qwq-32b":                          0.29 * MilliTokensUsd,
	"qwen-2.5-coder-32b":                    0.79 * MilliTokensUsd,
	"qwen-2.5-32b":                          0.79 * MilliTokensUsd,
	"mixtral-8x7b-32768":                    0.24 * MilliTokensUsd,
	"whisper-large-v3":                      0.111 * MilliTokensUsd,
	"whisper-large-v3-turbo":                0.04 * MilliTokensUsd,
	"distil-whisper-large-v3-en":            0.02 * MilliTokensUsd,
	"deepseek-r1-distill-qwen-32b":          0.69 * MilliTokensUsd,
	"deepseek-r1-distill-llama-70b-specdec": 0.75 * MilliTokensUsd,
	"deepseek-r1-distill-llama-70b":         0.75 * MilliTokensUsd,
	// https://platform.lingyiwanwu.com/docs#-计费单元
	"yi-34b-chat-0205": 2.5 * MilliRmb,
	"yi-34b-chat-200k": 12.0 * MilliRmb,
	"yi-vl-plus":       6.0 * MilliRmb,
	// https://platform.stepfun.com/docs/pricing/details
	"step-1-8k":    0.005 * MilliRmb,
	"step-1-32k":   0.015 * MilliRmb,
	"step-1-128k":  0.040 * MilliRmb,
	"step-1-256k":  0.095 * MilliRmb,
	"step-1-flash": 0.001 * MilliRmb,
	"step-2-16k":   0.038 * MilliRmb,
	"step-1v-8k":   0.005 * MilliRmb,
	"step-1v-32k":  0.015 * MilliRmb,
	// aws llama3 https://aws.amazon.com/cn/bedrock/pricing/
	"llama3-8b-8192(33)":  0.0003 / 0.002,  // $0.0003 / 1K tokens
	"llama3-70b-8192(33)": 0.00265 / 0.002, // $0.00265 / 1K tokens
	// https://cohere.com/pricing
	"command":               0.5,
	"command-nightly":       0.5,
	"command-light":         0.5,
	"command-light-nightly": 0.5,
	"command-r":             0.5 * MilliTokensUsd,
	"command-r-plus":        3.0 * MilliTokensUsd,
	// https://platform.deepseek.com/api-docs/pricing/
	"deepseek-chat":     0.27 * MilliTokensUsd,
	"deepseek-reasoner": 0.55 * MilliTokensUsd,
	// https://www.deepl.com/pro?cta=header-prices
	"deepl-zh": 25.0 * MilliTokensUsd,
	"deepl-en": 25.0 * MilliTokensUsd,
	"deepl-ja": 25.0 * MilliTokensUsd,
	// https://console.x.ai/
	"grok-beta": 5.0 * MilliTokensUsd,
	// vertex imagen3
	// https://cloud.google.com/vertex-ai/generative-ai/pricing#imagen-models
	"imagen-3.0-generate-001":      0.04 * ImageUsdPerPic,
	"imagen-3.0-generate-002":      0.04 * ImageUsdPerPic,
	"imagen-3.0-fast-generate-001": 0.02 * ImageUsdPerPic,
	"imagen-3.0-capability-001":    0.04 * ImageUsdPerPic,
	// -------------------------------------
	// replicate charges based on the number of generated images
	// https://replicate.com/pricing
	// -------------------------------------
	"black-forest-labs/flux-1.1-pro":                0.04 * ImageUsdPerPic,
	"black-forest-labs/flux-1.1-pro-ultra":          0.06 * ImageUsdPerPic,
	"black-forest-labs/flux-canny-dev":              0.025 * ImageUsdPerPic,
	"black-forest-labs/flux-canny-pro":              0.05 * ImageUsdPerPic,
	"black-forest-labs/flux-depth-dev":              0.025 * ImageUsdPerPic,
	"black-forest-labs/flux-depth-pro":              0.05 * ImageUsdPerPic,
	"black-forest-labs/flux-dev":                    0.025 * ImageUsdPerPic,
	"black-forest-labs/flux-dev-lora":               0.032 * ImageUsdPerPic,
	"black-forest-labs/flux-fill-dev":               0.04 * ImageUsdPerPic,
	"black-forest-labs/flux-fill-pro":               0.05 * ImageUsdPerPic,
	"black-forest-labs/flux-pro":                    0.055 * ImageUsdPerPic,
	"black-forest-labs/flux-redux-dev":              0.025 * ImageUsdPerPic,
	"black-forest-labs/flux-redux-schnell":          0.003 * ImageUsdPerPic,
	"black-forest-labs/flux-schnell":                0.003 * ImageUsdPerPic,
	"black-forest-labs/flux-schnell-lora":           0.02 * ImageUsdPerPic,
	"ideogram-ai/ideogram-v2":                       0.08 * ImageUsdPerPic,
	"ideogram-ai/ideogram-v2-turbo":                 0.05 * ImageUsdPerPic,
	"recraft-ai/recraft-v3":                         0.04 * ImageUsdPerPic,
	"recraft-ai/recraft-v3-svg":                     0.08 * ImageUsdPerPic,
	"stability-ai/stable-diffusion-3":               0.035 * ImageUsdPerPic,
	"stability-ai/stable-diffusion-3.5-large":       0.065 * ImageUsdPerPic,
	"stability-ai/stable-diffusion-3.5-large-turbo": 0.04 * ImageUsdPerPic,
	"stability-ai/stable-diffusion-3.5-medium":      0.035 * ImageUsdPerPic,
	// replicate chat models
	"anthropic/claude-3.5-haiku":                1.0 * MilliTokensUsd,
	"anthropic/claude-3.5-sonnet":               3.75 * MilliTokensUsd,
	"anthropic/claude-3.7-sonnet":               3.0 * MilliTokensUsd,
	"deepseek-ai/deepseek-r1":                   10.0 * MilliTokensUsd,
	"ibm-granite/granite-20b-code-instruct-8k":  0.100 * MilliTokensUsd,
	"ibm-granite/granite-3.0-2b-instruct":       0.030 * MilliTokensUsd,
	"ibm-granite/granite-3.0-8b-instruct":       0.050 * MilliTokensUsd,
	"ibm-granite/granite-3.1-2b-instruct":       0.030 * MilliTokensUsd,
	"ibm-granite/granite-3.1-8b-instruct":       0.030 * MilliTokensUsd,
	"ibm-granite/granite-3.2-8b-instruct":       0.030 * MilliTokensUsd,
	"ibm-granite/granite-8b-code-instruct-128k": 0.050 * MilliTokensUsd,
	"meta/llama-2-13b":                          0.100 * MilliTokensUsd,
	"meta/llama-2-13b-chat":                     0.100 * MilliTokensUsd,
	"meta/llama-2-70b":                          0.650 * MilliTokensUsd,
	"meta/llama-2-70b-chat":                     0.650 * MilliTokensUsd,
	"meta/llama-2-7b":                           0.050 * MilliTokensUsd,
	"meta/llama-2-7b-chat":                      0.050 * MilliTokensUsd,
	"meta/meta-llama-3.1-405b-instruct":         9.500 * MilliTokensUsd,
	"meta/meta-llama-3-70b":                     0.650 * MilliTokensUsd,
	"meta/meta-llama-3-70b-instruct":            0.650 * MilliTokensUsd,
	"meta/meta-llama-3-8b":                      0.050 * MilliTokensUsd,
	"meta/meta-llama-3-8b-instruct":             0.050 * MilliTokensUsd,
	"mistralai/mistral-7b-instruct-v0.2":        0.050 * MilliTokensUsd,
	"mistralai/mistral-7b-v0.1":                 0.050 * MilliTokensUsd,
	// -------------------------------------
	//https://openrouter.ai/models
	// -------------------------------------
	"01-ai/yi-large":                 1.5,
	"aetherwiing/mn-starcannon-12b":  0.6,
	"ai21/jamba-1-5-large":           4.0,
	"ai21/jamba-1-5-mini":            0.2,
	"ai21/jamba-instruct":            0.35,
	"aion-labs/aion-1.0":             6.0,
	"aion-labs/aion-1.0-mini":        1.2,
	"aion-labs/aion-rp-llama-3.1-8b": 0.1,
	"allenai/llama-3.1-tulu-3-405b":  5.0,
	"alpindale/goliath-120b":         4.6875,
	"alpindale/magnum-72b":           1.125,
	"amazon/nova-lite-v1":            0.12,
	"amazon/nova-micro-v1":           0.07,
	"amazon/nova-pro-v1":             1.6,
	"anthracite-org/magnum-v2-72b":   1.5,
	"anthracite-org/magnum-v4-72b":   1.125,
	"anthropic/claude-2":             12.0,
	"anthropic/claude-2.0":           12.0,
	"anthropic/claude-2.0:beta":      12.0,
	"anthropic/claude-2.1":           12.0,
	"anthropic/claude-2.1:beta":      12.0,
	"anthropic/claude-2:beta":        12.0,
	"anthropic/claude-3-haiku":       0.625,
	"anthropic/claude-3-haiku:beta":  0.625,
	"anthropic/claude-3-opus":        37.5,
	"anthropic/claude-3-opus:beta":   37.5,
	"anthropic/claude-3-sonnet":      7.5,
	"anthropic/claude-3-sonnet:beta": 7.5,
	// "anthropic/claude-3.5-haiku":                      2.0,
	"anthropic/claude-3.5-haiku-20241022":      2.0,
	"anthropic/claude-3.5-haiku-20241022:beta": 2.0,
	"anthropic/claude-3.5-haiku:beta":          2.0,
	// "anthropic/claude-3.5-sonnet":                     7.5,
	"anthropic/claude-3.5-sonnet-20240620":        7.5,
	"anthropic/claude-3.5-sonnet-20240620:beta":   7.5,
	"anthropic/claude-3.5-sonnet:beta":            7.5,
	"cognitivecomputations/dolphin-mixtral-8x22b": 0.45,
	"cognitivecomputations/dolphin-mixtral-8x7b":  0.25,
	"cohere/command":                                  0.95,
	"cohere/command-r":                                0.7125,
	"cohere/command-r-03-2024":                        0.7125,
	"cohere/command-r-08-2024":                        0.285,
	"cohere/command-r-plus":                           7.125,
	"cohere/command-r-plus-04-2024":                   7.125,
	"cohere/command-r-plus-08-2024":                   4.75,
	"cohere/command-r7b-12-2024":                      0.075,
	"databricks/dbrx-instruct":                        0.6,
	"deepseek/deepseek-chat":                          1.25,
	"deepseek/deepseek-chat-v2.5":                     1.0,
	"deepseek/deepseek-chat:free":                     0.0,
	"deepseek/deepseek-r1":                            7,
	"deepseek/deepseek-r1-distill-llama-70b":          0.345,
	"deepseek/deepseek-r1-distill-llama-70b:free":     0.0,
	"deepseek/deepseek-r1-distill-llama-8b":           0.02,
	"deepseek/deepseek-r1-distill-qwen-1.5b":          0.09,
	"deepseek/deepseek-r1-distill-qwen-14b":           0.075,
	"deepseek/deepseek-r1-distill-qwen-32b":           0.09,
	"deepseek/deepseek-r1:free":                       0.0,
	"eva-unit-01/eva-llama-3.33-70b":                  3.0,
	"eva-unit-01/eva-qwen-2.5-32b":                    1.7,
	"eva-unit-01/eva-qwen-2.5-72b":                    3.0,
	"google/gemini-2.0-flash-001":                     0.2,
	"google/gemini-2.0-flash-exp:free":                0.0,
	"google/gemini-2.0-flash-lite-preview-02-05:free": 0.0,
	"google/gemini-2.0-flash-thinking-exp-1219:free":  0.0,
	"google/gemini-2.0-flash-thinking-exp:free":       0.0,
	"google/gemini-2.0-pro-exp-02-05:free":            0.0,
	"google/gemini-exp-1206:free":                     0.0,
	"google/gemini-flash-1.5":                         0.15,
	"google/gemini-flash-1.5-8b":                      0.075,
	"google/gemini-flash-1.5-8b-exp":                  0.0,
	"google/gemini-pro":                               0.75,
	"google/gemini-pro-1.5":                           2.5,
	"google/gemini-pro-vision":                        0.75,
	"google/gemma-2-27b-it":                           0.135,
	"google/gemma-2-9b-it":                            0.03,
	"google/gemma-2-9b-it:free":                       0.0,
	"google/gemma-7b-it":                              0.075,
	"google/learnlm-1.5-pro-experimental:free":        0.0,
	"google/palm-2-chat-bison":                        1.0,
	"google/palm-2-chat-bison-32k":                    1.0,
	"google/palm-2-codechat-bison":                    1.0,
	"google/palm-2-codechat-bison-32k":                1.0,
	"gryphe/mythomax-l2-13b":                          0.0325,
	"gryphe/mythomax-l2-13b:free":                     0.0,
	"huggingfaceh4/zephyr-7b-beta:free":               0.0,
	"infermatic/mn-inferor-12b":                       0.6,
	"inflection/inflection-3-pi":                      5.0,
	"inflection/inflection-3-productivity":            5.0,
	"jondurbin/airoboros-l2-70b":                      0.25,
	"liquid/lfm-3b":                                   0.01,
	"liquid/lfm-40b":                                  0.075,
	"liquid/lfm-7b":                                   0.005,
	"mancer/weaver":                                   1.125,
	"meta-llama/llama-2-13b-chat":                     0.11,
	"meta-llama/llama-2-70b-chat":                     0.45,
	"meta-llama/llama-3-70b-instruct":                 0.2,
	"meta-llama/llama-3-8b-instruct":                  0.03,
	"meta-llama/llama-3-8b-instruct:free":             0.0,
	"meta-llama/llama-3.1-405b":                       1.0,
	"meta-llama/llama-3.1-405b-instruct":              0.4,
	"meta-llama/llama-3.1-70b-instruct":               0.15,
	"meta-llama/llama-3.1-8b-instruct":                0.025,
	"meta-llama/llama-3.2-11b-vision-instruct":        0.0275,
	"meta-llama/llama-3.2-11b-vision-instruct:free":   0.0,
	"meta-llama/llama-3.2-1b-instruct":                0.005,
	"meta-llama/llama-3.2-3b-instruct":                0.0125,
	"meta-llama/llama-3.2-90b-vision-instruct":        0.8,
	"meta-llama/llama-3.3-70b-instruct":               0.15,
	"meta-llama/llama-3.3-70b-instruct:free":          0.0,
	"meta-llama/llama-guard-2-8b":                     0.1,
	"microsoft/phi-3-medium-128k-instruct":            0.5,
	"microsoft/phi-3-medium-128k-instruct:free":       0.0,
	"microsoft/phi-3-mini-128k-instruct":              0.05,
	"microsoft/phi-3-mini-128k-instruct:free":         0.0,
	"microsoft/phi-3.5-mini-128k-instruct":            0.05,
	"microsoft/phi-4":                                 0.07,
	"microsoft/wizardlm-2-7b":                         0.035,
	"microsoft/wizardlm-2-8x22b":                      0.25,
	"minimax/minimax-01":                              0.55,
	"mistralai/codestral-2501":                        0.45,
	"mistralai/codestral-mamba":                       0.125,
	"mistralai/ministral-3b":                          0.02,
	"mistralai/ministral-8b":                          0.05,
	"mistralai/mistral-7b-instruct":                   0.0275,
	"mistralai/mistral-7b-instruct-v0.1":              0.1,
	"mistralai/mistral-7b-instruct-v0.3":              0.0275,
	"mistralai/mistral-7b-instruct:free":              0.0,
	"mistralai/mistral-large":                         3.0,
	"mistralai/mistral-large-2407":                    3.0,
	"mistralai/mistral-large-2411":                    3.0,
	"mistralai/mistral-medium":                        4.05,
	"mistralai/mistral-nemo":                          0.04,
	"mistralai/mistral-nemo:free":                     0.0,
	"mistralai/mistral-small":                         0.3,
	"mistralai/mistral-small-24b-instruct-2501":       0.07,
	"mistralai/mistral-small-24b-instruct-2501:free":  0.0,
	"mistralai/mistral-tiny":                          0.125,
	"mistralai/mixtral-8x22b-instruct":                0.45,
	"mistralai/mixtral-8x7b":                          0.3,
	"mistralai/mixtral-8x7b-instruct":                 0.12,
	"mistralai/pixtral-12b":                           0.05,
	"mistralai/pixtral-large-2411":                    3.0,
	"neversleep/llama-3-lumimaid-70b":                 2.25,
	"neversleep/llama-3-lumimaid-8b":                  0.5625,
	"neversleep/llama-3-lumimaid-8b:extended":         0.5625,
	"neversleep/llama-3.1-lumimaid-70b":               2.25,
	"neversleep/llama-3.1-lumimaid-8b":                0.5625,
	"neversleep/noromaid-20b":                         1.125,
	"nothingiisreal/mn-celeste-12b":                   0.6,
	"nousresearch/hermes-2-pro-llama-3-8b":            0.02,
	"nousresearch/hermes-3-llama-3.1-405b":            0.4,
	"nousresearch/hermes-3-llama-3.1-70b":             0.15,
	"nousresearch/nous-hermes-2-mixtral-8x7b-dpo":     0.3,
	"nousresearch/nous-hermes-llama2-13b":             0.085,
	"nvidia/llama-3.1-nemotron-70b-instruct":          0.15,
	"nvidia/llama-3.1-nemotron-70b-instruct:free":     0.0,
	"openai/chatgpt-4o-latest":                        7.5,
	"openai/gpt-3.5-turbo":                            0.75,
	"openai/gpt-3.5-turbo-0125":                       0.75,
	"openai/gpt-3.5-turbo-0613":                       1.0,
	"openai/gpt-3.5-turbo-1106":                       1.0,
	"openai/gpt-3.5-turbo-16k":                        2.0,
	"openai/gpt-3.5-turbo-instruct":                   1.0,
	"openai/gpt-4":                                    30.0,
	"openai/gpt-4-0314":                               30.0,
	"openai/gpt-4-1106-preview":                       15.0,
	"openai/gpt-4-32k":                                60.0,
	"openai/gpt-4-32k-0314":                           60.0,
	"openai/gpt-4-turbo":                              15.0,
	"openai/gpt-4-turbo-preview":                      15.0,
	"openai/gpt-4o":                                   5.0,
	"openai/gpt-4o-2024-05-13":                        7.5,
	"openai/gpt-4o-2024-08-06":                        5.0,
	"openai/gpt-4o-2024-11-20":                        5.0,
	"openai/gpt-4o-mini":                              0.3,
	"openai/gpt-4o-mini-2024-07-18":                   0.3,
	"openai/gpt-4o:extended":                          9.0,
	"openai/gpt-4.5-preview":                          75,
	"openai/o1":                                       30.0,
	"openai/o1-mini":                                  2.2,
	"openai/o1-mini-2024-09-12":                       2.2,
	"openai/o1-preview":                               30.0,
	"openai/o1-preview-2024-09-12":                    30.0,
	"openai/o3-mini":                                  2.2,
	"openai/o3-mini-high":                             2.2,
	"openchat/openchat-7b":                            0.0275,
	"openchat/openchat-7b:free":                       0.0,
	// "openrouter/auto":                                 -500000.0,
	"perplexity/llama-3.1-sonar-huge-128k-online":  2.5,
	"perplexity/llama-3.1-sonar-large-128k-chat":   0.5,
	"perplexity/llama-3.1-sonar-large-128k-online": 0.5,
	"perplexity/llama-3.1-sonar-small-128k-chat":   0.1,
	"perplexity/llama-3.1-sonar-small-128k-online": 0.1,
	"perplexity/sonar":                             0.5,
	"perplexity/sonar-reasoning":                   2.5,
	"pygmalionai/mythalion-13b":                    0.6,
	"qwen/qvq-72b-preview":                         0.25,
	"qwen/qwen-2-72b-instruct":                     0.45,
	"qwen/qwen-2-7b-instruct":                      0.027,
	"qwen/qwen-2-7b-instruct:free":                 0.0,
	"qwen/qwen-2-vl-72b-instruct":                  0.2,
	"qwen/qwen-2-vl-7b-instruct":                   0.05,
	"qwen/qwen-2.5-72b-instruct":                   0.2,
	"qwen/qwen-2.5-7b-instruct":                    0.025,
	"qwen/qwen-2.5-coder-32b-instruct":             0.08,
	"qwen/qwen-max":                                3.2,
	"qwen/qwen-plus":                               0.6,
	"qwen/qwen-turbo":                              0.1,
	"qwen/qwen-vl-plus:free":                       0.0,
	"qwen/qwen2.5-vl-72b-instruct:free":            0.0,
	"qwen/qwq-32b-preview":                         0.09,
	"raifle/sorcererlm-8x22b":                      2.25,
	"sao10k/fimbulvetr-11b-v2":                     0.6,
	"sao10k/l3-euryale-70b":                        0.4,
	"sao10k/l3-lunaris-8b":                         0.03,
	"sao10k/l3.1-70b-hanami-x1":                    1.5,
	"sao10k/l3.1-euryale-70b":                      0.4,
	"sao10k/l3.3-euryale-70b":                      0.4,
	"sophosympatheia/midnight-rose-70b":            0.4,
	"sophosympatheia/rogue-rose-103b-v0.2:free":    0.0,
	"teknium/openhermes-2.5-mistral-7b":            0.085,
	"thedrummer/rocinante-12b":                     0.25,
	"thedrummer/unslopnemo-12b":                    0.25,
	"undi95/remm-slerp-l2-13b":                     0.6,
	"undi95/toppy-m-7b":                            0.035,
	"undi95/toppy-m-7b:free":                       0.0,
	// https://x.ai/api
	"x-ai/grok-2-1212":        5.0,
	"x-ai/grok-2-vision-1212": 5.0,
	"x-ai/grok-beta":          7.5,
	"x-ai/grok-vision-beta":   7.5,
	"xwin-lm/xwin-lm-70b":     1.875,
}

// CompletionRatio is the price ratio between completion tokens and prompt tokens
var CompletionRatio = map[string]float64{
	// -------------------------------------
	// aws llama3
	// -------------------------------------
	"llama3-8b-8192(33)":  0.0006 / 0.0003,
	"llama3-70b-8192(33)": 0.0035 / 0.00265,
	// -------------------------------------
	// whisper
	// -------------------------------------
	"whisper-1":                  0, // only count input tokens
	"whisper-large-v3":           0, // only count input tokens
	"whisper-large-v3-turbo":     0, // only count input tokens
	"distil-whisper-large-v3-en": 0, // only count input tokens
	// -------------------------------------
	// deepseek
	// -------------------------------------
	"deepseek-chat":     1.1 / 0.27,
	"deepseek-reasoner": 2.19 / 0.55,
	// -------------------------------------
	// openrouter
	// -------------------------------------
	"deepseek/deepseek-chat": 1,
	"deepseek/deepseek-r1":   1,
	// -------------------------------------
	// groq
	// -------------------------------------
	"llama-3.3-70b-versatile":               0.79 / 0.59,
	"llama-3.1-8b-instant":                  0.08 / 0.05,
	"llama3-70b-8192":                       0.79 / 0.59,
	"llama3-8b-8192":                        0.08 / 0.05,
	"gemma2-9b-it":                          1.0,
	"llama-3.2-11b-text-preview":            1.0,
	"llama-3.2-11b-vision-preview":          1.0,
	"llama-3.2-1b-preview":                  1.0,
	"llama-3.2-3b-preview":                  1.0,
	"llama-3.2-90b-text-preview":            1.0,
	"llama-3.2-90b-vision-preview":          1.0,
	"llama-guard-3-8b":                      1.0,
	"llama3-groq-70b-8192-tool-use-preview": 0.79 / 0.59,
	"llama3-groq-8b-8192-tool-use-preview":  0.08 / 0.05,
	"mixtral-8x7b-32768":                    1.0,
	"deepseek-r1-distill-qwen-32b":          1.0,
	"deepseek-r1-distill-llama-70b-specdec": 0.99 / 0.75,
	"deepseek-r1-distill-llama-70b":         0.99 / 0.75,
	"llama-3.3-70b-specdec":                 0.99 / 0.59,
	"mistral-saba-24b":                      1.0,
	"qwen-qwq-32b":                          0.39 / 0.29,
	"qwen-2.5-coder-32b":                    1.0,
	"qwen-2.5-32b":                          1.0,
	// -------------------------------------
	// Replicate
	// -------------------------------------
	"anthropic/claude-3.5-haiku":                5.0 / 1.0,
	"anthropic/claude-3.5-sonnet":               18.75 / 3.75,
	"anthropic/claude-3.7-sonnet":               15.0 / 3.0,
	"deepseek-ai/deepseek-r1":                   10.0 / 10.0,
	"ibm-granite/granite-20b-code-instruct-8k":  0.5 / 0.1,
	"ibm-granite/granite-3.0-2b-instruct":       0.25 / 0.03,
	"ibm-granite/granite-3.0-8b-instruct":       0.25 / 0.05,
	"ibm-granite/granite-3.1-2b-instruct":       0.25 / 0.03,
	"ibm-granite/granite-3.1-8b-instruct":       0.25 / 0.03,
	"ibm-granite/granite-3.2-8b-instruct":       0.25 / 0.03,
	"ibm-granite/granite-8b-code-instruct-128k": 0.25 / 0.05,
	"meta/llama-2-13b":                          0.5 / 0.1,
	"meta/llama-2-13b-chat":                     0.5 / 0.1,
	"meta/llama-2-70b":                          2.75 / 0.65,
	"meta/llama-2-70b-chat":                     2.75 / 0.65,
	"meta/llama-2-7b":                           0.25 / 0.05,
	"meta/llama-2-7b-chat":                      0.25 / 0.05,
	"meta/meta-llama-3.1-405b-instruct":         9.5 / 9.5,
	"meta/meta-llama-3-70b":                     2.75 / 0.65,
	"meta/meta-llama-3-70b-instruct":            2.75 / 0.65,
	"meta/meta-llama-3-8b":                      0.25 / 0.05,
	"meta/meta-llama-3-8b-instruct":             0.25 / 0.05,
	"mistralai/mistral-7b-instruct-v0.2":        0.25 / 0.05,
	"mistralai/mistral-7b-v0.1":                 0.25 / 0.05,
}

// AudioRatio represents the price ratio between audio tokens and text tokens
var AudioRatio = map[string]float64{
	"gpt-4o-audio-preview":                 16,
	"gpt-4o-audio-preview-2024-12-17":      16,
	"gpt-4o-audio-preview-2024-10-01":      40,
	"gpt-4o-mini-audio-preview":            10 / 0.15,
	"gpt-4o-mini-audio-preview-2024-12-17": 10 / 0.15,
	"gpt-4o-transcribe":                    6 / 2.5,
	"gpt-4o-mini-transcribe":               3 / 1.25,
}

// GetAudioPromptRatio returns the audio prompt ratio for the given model.
func GetAudioPromptRatio(actualModelName string) float64 {
	var v float64
	if ratio, ok := AudioRatio[actualModelName]; ok {
		v = ratio
	} else {
		v = 16
	}

	return v
}

// AudioCompletionRatio is the completion ratio for audio models.
var AudioCompletionRatio = map[string]float64{
	"whisper-1":                            0,
	"gpt-4o-audio-preview":                 2,
	"gpt-4o-audio-preview-2024-12-17":      2,
	"gpt-4o-audio-preview-2024-10-01":      2,
	"gpt-4o-mini-audio-preview":            2,
	"gpt-4o-mini-audio-preview-2024-12-17": 2,
}

// GetAudioCompletionRatio returns the completion ratio for audio models.
func GetAudioCompletionRatio(actualModelName string) float64 {
	var v float64
	if ratio, ok := AudioCompletionRatio[actualModelName]; ok {
		v = ratio
	} else {
		v = 2
	}

	return v
}

// AudioTokensPerSecond is the number of audio tokens per second for each model.
var AudioPromptTokensPerSecond = map[string]float64{
	// Whisper API price is $0.0001/sec. One-api's historical ratio is 15,
	// corresponding to $0.03/kilo_tokens.
	// After conversion, tokens per second should be 0.0001/0.03*1000 = 3.3333.
	"whisper-1": 0.0001 / 0.03 * 1000,
	// gpt-4o-audio series processes 10 tokens per second
	"gpt-4o-audio-preview":                 10,
	"gpt-4o-audio-preview-2024-12-17":      10,
	"gpt-4o-audio-preview-2024-10-01":      10,
	"gpt-4o-mini-audio-preview":            10,
	"gpt-4o-mini-audio-preview-2024-12-17": 10,
	"gpt-4o-transcribe":                    10,
	"gpt-4o-mini-transcribe":               10,
	"gpt-4o-mini-tts":                      10,
}

// GetAudioPromptTokensPerSecond returns the number of audio tokens per second
// for the given model.
func GetAudioPromptTokensPerSecond(actualModelName string) float64 {
	var v float64
	if tokensPerSecond, ok := AudioPromptTokensPerSecond[actualModelName]; ok {
		v = tokensPerSecond
	} else {
		v = 10
	}

	return v
}

var (
	DefaultModelRatio      map[string]float64
	DefaultCompletionRatio map[string]float64
)

func init() {
	DefaultModelRatio = make(map[string]float64)
	for k, v := range ModelRatio {
		DefaultModelRatio[k] = v
	}
	DefaultCompletionRatio = make(map[string]float64)
	for k, v := range CompletionRatio {
		DefaultCompletionRatio[k] = v
	}
}

func AddNewMissingRatio(oldRatio string) string {
	newRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(oldRatio), &newRatio)
	if err != nil {
		logger.SysError("error unmarshalling old ratio: " + err.Error())
		return oldRatio
	}
	for k, v := range DefaultModelRatio {
		if _, ok := newRatio[k]; !ok {
			newRatio[k] = v
		}
	}
	jsonBytes, err := json.Marshal(newRatio)
	if err != nil {
		logger.SysError("error marshalling new ratio: " + err.Error())
		return oldRatio
	}
	return string(jsonBytes)
}

func ModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(ModelRatio)
	if err != nil {
		logger.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

// UpdateModelRatioByJSONString updates the ModelRatio map with the given JSON string.
func UpdateModelRatioByJSONString(jsonStr string) error {
	modelRatioLock.Lock()
	defer modelRatioLock.Unlock()
	ModelRatio = make(map[string]float64)
	err := json.Unmarshal([]byte(jsonStr), &ModelRatio)
	if err != nil {
		logger.SysError("error unmarshalling model ratio: " + err.Error())
		return err
	}

	// f3300f08e25e212f1b32ae1f678eb7ec2dec6a8c change the ratio of image models,
	// so we need to multiply the ratio by 1000 for legacy settings.
	// for name, ratio := range ModelRatio {
	// 	switch name {
	// 	case "dall-e-2",
	// 		"dall-e-3",
	// 		"imagen-3.0-generate-001",
	// 		"imagen-3.0-generate-002",
	// 		"imagen-3.0-fast-generate-001",
	// 		"imagen-3.0-capability-001",
	// 		"ali-stable-diffusion-xl",
	// 		"ali-stable-diffusion-v1.5",
	// 		"black-forest-labs/flux-1.1-pro",
	// 		"black-forest-labs/flux-1.1-pro-ultra",
	// 		"black-forest-labs/flux-canny-dev",
	// 		"black-forest-labs/flux-canny-pro",
	// 		"black-forest-labs/flux-depth-dev",
	// 		"black-forest-labs/flux-depth-pro",
	// 		"black-forest-labs/flux-dev",
	// 		"black-forest-labs/flux-dev-lora",
	// 		"black-forest-labs/flux-fill-dev",
	// 		"black-forest-labs/flux-fill-pro",
	// 		"black-forest-labs/flux-pro",
	// 		"black-forest-labs/flux-redux-dev",
	// 		"black-forest-labs/flux-redux-schnell",
	// 		"black-forest-labs/flux-schnell",
	// 		"black-forest-labs/flux-schnell-lora",
	// 		"ideogram-ai/ideogram-v2",
	// 		"ideogram-ai/ideogram-v2-turbo",
	// 		"recraft-ai/recraft-v3",
	// 		"recraft-ai/recraft-v3-svg",
	// 		"stability-ai/stable-diffusion-3",
	// 		"stability-ai/stable-diffusion-3.5-large",
	// 		"stability-ai/stable-diffusion-3.5-large-turbo",
	// 		"stability-ai/stable-diffusion-3.5-medium":
	// 		if ratio < 1000 {
	// 			logger.SysWarnf("the model ratio of %s is less than 1000, please check it", name)
	// 			ModelRatio[name] = ratio * 1000
	// 		}
	// 	}
	// }

	return nil
}

func GetModelRatio(name string, channelType int) float64 {
	modelRatioLock.RLock()
	defer modelRatioLock.RUnlock()
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	if strings.HasPrefix(name, "command-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}

	model := fmt.Sprintf("%s(%d)", name, channelType)

	for _, targetName := range []string{model, name} {
		for _, ratioMap := range []map[string]float64{
			ModelRatio,
			DefaultModelRatio,
			AudioRatio,
		} {
			if ratio, ok := ratioMap[targetName]; ok {
				return ratio
			}
		}
	}

	logger.SysError("model ratio not found: " + name)
	return 2.5 * MilliTokensUsd
}

// CompletionRatio2JSONString returns the CompletionRatio map as a JSON string.
func CompletionRatio2JSONString() string {
	jsonBytes, err := json.Marshal(CompletionRatio)
	if err != nil {
		logger.SysError("error marshalling completion ratio: " + err.Error())
	}
	return string(jsonBytes)
}

// completionRatioLock is a mutex for synchronizing access to the CompletionRatio map.
var completionRatioLock sync.RWMutex

// UpdateCompletionRatioByJSONString updates the CompletionRatio map with the given JSON string.
func UpdateCompletionRatioByJSONString(jsonStr string) error {
	completionRatioLock.Lock()
	defer completionRatioLock.Unlock()
	CompletionRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &CompletionRatio)
}

// GetCompletionRatio returns the completion ratio for the given model name and channel type.
func GetCompletionRatio(name string, channelType int) float64 {
	completionRatioLock.RLock()
	defer completionRatioLock.RUnlock()
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	model := fmt.Sprintf("%s(%d)", name, channelType)

	name = strings.TrimPrefix(name, "openai/")
	for _, targetName := range []string{model, name} {
		for _, ratioMap := range []map[string]float64{
			CompletionRatio,
			DefaultCompletionRatio,
			AudioCompletionRatio,
		} {
			// first try the model name
			if ratio, ok := ratioMap[targetName]; ok {
				return ratio
			}

			// then try the model name without some special prefix
			normalizedTargetName := strings.TrimPrefix(targetName, "openai/")
			if ratio, ok := ratioMap[normalizedTargetName]; ok {
				return ratio
			}
		}
	}

	// openai
	switch {
	case strings.HasPrefix(name, "gpt-3.5"):
		switch {
		case name == "gpt-3.5-turbo" || strings.HasSuffix(name, "0125"):
			// https://openai.com/blog/new-embedding-models-and-api-updates
			// Updated GPT-3.5 Turbo model and lower pricing
			return 3
		case strings.HasSuffix(name, "1106"):
			return 2
		default:
			return 4.0 / 3.0
		}
	case name == "chatgpt-4o-latest":
		return 3
	case strings.HasPrefix(name, "gpt-4"):
		switch {
		case strings.HasPrefix(name, "gpt-4o"):
			if name == "gpt-4o-2024-05-13" {
				return 3
			}
			return 4
		case strings.HasPrefix(name, "gpt-4-"):
			return 3
		default:
			return 2
		}
	// including o1/o1-preview/o1-mini
	case strings.HasPrefix(name, "o1") ||
		strings.HasPrefix(name, "o3"):
		return 4
	}

	if strings.HasPrefix(name, "claude-3") {
		return 5
	}
	if strings.HasPrefix(name, "claude-") {
		return 3
	}
	if strings.HasPrefix(name, "mistral-") {
		return 3
	}
	if strings.HasPrefix(name, "gemini-") {
		return 4
	}

	switch name {
	case "llama2-70b-4096":
		return 0.8 / 0.64
	case "llama3-8b-8192":
		return 2
	case "llama3-70b-8192":
		return 0.79 / 0.59
	case "command", "command-light", "command-nightly", "command-light-nightly":
		return 2
	case "command-r":
		return 3
	case "command-r-plus":
		return 5
	case "grok-beta":
		return 3
	// Replicate Models
	// https://replicate.com/pricing
	case "ibm-granite/granite-20b-code-instruct-8k":
		return 5
	case "ibm-granite/granite-3.0-2b-instruct":
		return 8.333333333333334
	case "ibm-granite/granite-3.0-8b-instruct",
		"ibm-granite/granite-8b-code-instruct-128k":
		return 5
	case "meta/llama-2-13b",
		"meta/llama-2-13b-chat",
		"meta/llama-2-7b",
		"meta/llama-2-7b-chat",
		"meta/meta-llama-3-8b",
		"meta/meta-llama-3-8b-instruct":
		return 5
	case "meta/llama-2-70b",
		"meta/llama-2-70b-chat",
		"meta/meta-llama-3-70b",
		"meta/meta-llama-3-70b-instruct":
		return 2.750 / 0.650 // ≈4.230769
	case "meta/meta-llama-3.1-405b-instruct":
		return 1
	case "mistralai/mistral-7b-instruct-v0.2",
		"mistralai/mistral-7b-v0.1":
		return 5
	case "mistralai/mixtral-8x7b-instruct-v0.1":
		return 1.000 / 0.300 // ≈3.333333
	}

	logger.SysWarn(fmt.Sprintf("completion ratio not found for model: %s (channel type: %d), using default value 1", name, channelType))
	return 1
}
