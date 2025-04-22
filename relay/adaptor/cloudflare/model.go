package cloudflare

import "github.com/songquanpeng/one-api/relay/model"

type Request struct {
	Messages    []model.Message `json:"messages,omitempty"`
	Lora        string          `json:"lora,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Prompt      string          `json:"prompt,omitempty"`
	Raw         bool            `json:"raw,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
}

type CloudflareImageRequest struct {
	// Model 是图像生成模型，
	Model string `json:"model,omitempty" validate:"oneof=flux-1-schnell stable-diffusion-xl-lightning stable-diffusion-v1-5-img2img dreamshaper-8-lcm stable-diffusion-v1-5-inpainting stable-diffusion-xl-base-1.0"`
	// Prompt 是图像生成的文本描述，必填，最小长度为 1
	Prompt string `json:"prompt,omitempty" validate:"min=1,max=1024"`
	// Steps 是扩散步数，仅用于 flux-1-schnell，可选，默认 4，范围 1-8
	Steps int `json:"steps,omitempty" validate:"min=1,max=8"`
	// NumSteps 是扩散步数，用于其他模型，可选，默认 20，范围 1-20
	NumSteps int `json:"num_steps,omitempty" validate:"min=1,max=20"`
	// NegativePrompt 是要避免的元素描述，可选
	NegativePrompt *string `json:"negative_prompt,omitempty"`
	// Height 是生成图像的高度（像素），可选，范围 256-2048
	Height int `json:"height,omitempty" validate:"min=256,max=2048"`
	// Width 是生成图像的宽度（像素），可选，范围 256-2048
	Width int `json:"width,omitempty" validate:"min=256,max=2048"`
	// Image 是 img2img 任务的输入图像数据，可选，使用 8 位无符号整数数组表示
	Image []uint8 `json:"image,omitempty"`
	// ImageB64 是 img2img 任务的 base64 编码图像，可选
	ImageB64 string `json:"image_b64,omitempty"`
	// Mask 是 inpainting 任务的掩码数据，可选，使用 8 位无符号整数数组表示
	Mask []uint8 `json:"mask,omitempty"`
	// Strength 是 img2img 任务的变换强度，可选，默认 1，范围 0-1
	Strength float64 `json:"strength,omitempty" validate:"min=0,max=1"`
	// Guidance 是生成图像与提示的贴合程度，可选，默认 7.5，无特定范围限制
	Guidance float64 `json:"guidance,omitempty"`
	// Seed 是随机种子，用于重现生成结果，可选
	Seed int `json:"seed,omitempty"`
}

// ImageResponse 用于解析 flux-1-schnell 的 JSON 响应
type CloudflareImageResponse struct {
	Image    string            `json:"image,omitempty"`
	Errors   []CloudflareError `json:"errors,omitempty"`
	Success  bool              `json:"success,omitempty"`
	Result   map[string]string `json:"result,omitempty"`
	Messages any               `json:"messages,omitempty"`
}

// CloudflareError 表示 Cloudflare API 错误的详细信息
type CloudflareError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
