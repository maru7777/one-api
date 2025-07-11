package replicate

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"io"
	"time"

	"github.com/Laisky/errors/v2"

	"github.com/songquanpeng/one-api/relay/model"
)

// toFluxRemixRequest convert OpenAI's image edit request to Flux's remix request.
//
// Note that the mask formats of OpenAI and Flux are different:
// OpenAI's mask sets the parts to be modified as transparent (0, 0, 0, 0),
// while Flux sets the parts to be modified as black (255, 255, 255, 255),
// so we need to convert the format here.
//
// Both OpenAI's Image and Mask are browser-native ImageData,
// which need to be converted to base64 dataURI format.
func Convert2FluxRemixRequest(req *model.OpenaiImageEditRequest) (*InpaintingImageByFlusReplicateRequest, error) {
	if req.ResponseFormat != "b64_json" {
		return nil, errors.New("response_format must be b64_json for replicate models")
	}

	fluxReq := &InpaintingImageByFlusReplicateRequest{
		Input: FluxInpaintingInput{
			Prompt:           req.Prompt,
			Seed:             int(time.Now().UnixNano()),
			Steps:            30,
			Guidance:         3,
			SafetyTolerance:  5,
			PromptUnsampling: false,
			OutputFormat:     "png",
		},
	}

	imgFile, err := req.Image.Open()
	if err != nil {
		return nil, errors.Wrap(err, "open image file")
	}
	defer imgFile.Close()
	imgData, err := io.ReadAll(imgFile)
	if err != nil {
		return nil, errors.Wrap(err, "read image file")
	}

	maskFile, err := req.Mask.Open()
	if err != nil {
		return nil, errors.Wrap(err, "open mask file")
	}
	defer maskFile.Close()

	// Convert image to base64
	imageBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgData)
	fluxReq.Input.Image = imageBase64

	// Convert mask data to RGBA
	maskPNG, err := png.Decode(maskFile)
	if err != nil {
		return nil, errors.Wrap(err, "decode mask file")
	}

	// convert mask to RGBA
	var maskRGBA *image.RGBA
	switch converted := maskPNG.(type) {
	case *image.RGBA:
		maskRGBA = converted
	default:
		// Convert to RGBA
		bounds := maskPNG.Bounds()
		maskRGBA = image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				maskRGBA.Set(x, y, maskPNG.At(x, y))
			}
		}
	}

	maskData := maskRGBA.Pix
	invertedMask := make([]byte, len(maskData))
	for i := 0; i+4 <= len(maskData); i += 4 {
		// If pixel is transparent (alpha = 0), make it black (255)
		if maskData[i+3] == 0 {
			invertedMask[i] = 255   // R
			invertedMask[i+1] = 255 // G
			invertedMask[i+2] = 255 // B
			invertedMask[i+3] = 255 // A
		} else {
			// Copy original pixel
			copy(invertedMask[i:i+4], maskData[i:i+4])
		}
	}

	// Convert inverted mask to base64 encoded png image
	invertedMaskRGBA := &image.RGBA{
		Pix:    invertedMask,
		Stride: maskRGBA.Stride,
		Rect:   maskRGBA.Rect,
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, invertedMaskRGBA)
	if err != nil {
		return nil, errors.Wrap(err, "encode inverted mask to png")
	}

	invertedMaskBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	fluxReq.Input.Mask = invertedMaskBase64

	return fluxReq, nil
}

// DrawImageRequest draw image by fluxpro
//
// https://replicate.com/black-forest-labs/flux-pro?prediction=kg1krwsdf9rg80ch1sgsrgq7h8&output=json
type DrawImageRequest struct {
	Input ImageInput `json:"input"`
}

// ImageInput is input of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-1.1-pro/api/schema
type ImageInput struct {
	Steps  int    `json:"steps" binding:"required,min=1"`
	Prompt string `json:"prompt" binding:"required,min=5"`
	// ImagePrompt is the image prompt, only works for flux-1.1-pro
	ImagePrompt *string `json:"image_prompt,omitempty"`
	// InputImage is the input image, only works for flux-kontext-pro
	InputImage      *string `json:"input_image,omitempty"`
	Guidance        int     `json:"guidance" binding:"required,min=2,max=5"`
	Interval        int     `json:"interval" binding:"required,min=1,max=4"`
	AspectRatio     string  `json:"aspect_ratio" binding:"required,oneof=1:1 16:9 2:3 3:2 4:5 5:4 9:16"`
	SafetyTolerance int     `json:"safety_tolerance" binding:"required,min=1,max=5"`
	Seed            int     `json:"seed"`
	NImages         int     `json:"n_images" binding:"required,min=1,max=8"`
	Width           int     `json:"width" binding:"required,min=256,max=1440"`
	Height          int     `json:"height" binding:"required,min=256,max=1440"`
}

// replicateImageRequest 代表一个通用的图像生成请求。
type replicateImageRequest struct {
	// Model string `json:"model" binding:"required"`
	// Input 包含图像生成的所有输入参数。
	Input Input `json:"input"`
}

// Input 结构体包含所有可能的输入参数。
type Input struct {
	// === 基本参数 ===
	// Prompt 是图像生成的文本提示。
	// 在大多数模型中是必需的，除非使用了 redux_image（例如 flux-redux-dev, flux-redux-schnell）。
	Prompt string `json:"prompt" validate:"omitempty" description:"图像生成的文本提示。除非提供 redux_image，否则是必需的。"`

	// Seed 是用于可重现生成的随机种子。
	// 可选，整数，在某些模型中最大值为 2147483647（例如 ideogram-ai）。
	Seed int `json:"seed,omitempty" validate:"omitempty,max=2147483647" description:"用于可重现的随机种子。在某些模型中最大值为 2147483647。"`

	// NegativePrompt 用于避免图像中出现某些元素。
	// 可选，字符串，用于 google/imagen-3, ideogram-ai 等模型。
	NegativePrompt string `json:"negative_prompt,omitempty" description:"用于避免图像中出现某些元素的文本提示。"`

	// === 图像尺寸和比例 ===
	// Width 是生成图像的宽度。
	// 可选，整数，范围 256-2048，在某些模型中必须是 32 的倍数（例如 flux-1.1-pro）。
	Width int `json:"width,omitempty" validate:"omitempty,min=256,max=2048" description:"图像宽度（像素），256-2048，在某些模型中必须是 32 的倍数。"`

	// Height 是生成图像的高度。
	// 可选，整数，范围 256-2048，在某些模型中必须是 32 的倍数（例如 flux-1.1-pro）。
	Height int `json:"height,omitempty" validate:"omitempty,min=256,max=2048" description:"图像高度（像素），256-2048，在某些模型中必须是 32 的倍数。"`

	// AspectRatio 是生成图像的宽高比。
	// 可选，字符串，涵盖所有模型的枚举值。
	AspectRatio string `json:"aspect_ratio,omitempty" validate:"omitempty,oneof=custom 1:1 16:9 21:9 3:2 2:3 4:5 5:4 9:16 3:4 4:3 9:21 16:10 10:16 3:1 1:3 7:5 5:7" description:"宽高比，因模型而异。"`

	// Size 指定确切的宽度 x 高度。
	// 可选，字符串，recraft-ai 模型的枚举值。
	Size string `json:"size,omitempty" validate:"omitempty,oneof=1024x1024 1365x1024 1024x1365 1536x1024 1024x1536 1820x1024 1024x1820 1024x2048 2048x1024 1434x1024 1024x1434 1024x1280 1280x1024 1024x1707 1707x1024" description:"确切的宽度 x 高度，在某些模型中覆盖 aspect_ratio。"`

	// Resolution 指定确切的分辨率。
	// 可选，字符串，ideogram-ai 模型的枚举值。
	Resolution string `json:"resolution,omitempty" validate:"omitempty,oneof=None 512x1536 576x1408 576x1472 576x1536 640x1344 640x1408 640x1472 640x1536 704x1152 704x1216 704x1280 704x1344 704x1408 704x1472 736x1312 768x1088 768x1216 768x1280 768x1344 832x960 832x1024 832x1088 832x1152 832x1216 832x1248 864x1152 896x960 896x1024 896x1088 896x1120 896x1152 960x832 960x896 960x1024 960x1088 1024x832 1024x896 1024x960 1024x1024 1088x768 1088x832 1088x896 1088x960 1120x896 1152x704 1152x832 1152x864 1152x896 1216x704 1216x768 1216x832 1248x832 1280x704 1280x768 1280x800 1312x736 1344x640 1344x704 1344x768 1408x576 1408x640 1408x704 1472x576 1472x640 1472x704 1536x512 1536x576 1536x640" description:"确切的分辨率，在 ideogram-ai 模型中覆盖 aspect_ratio。"`

	// Megapixels 指定近似的百万像素。
	// 可选，字符串，flux 模型的枚举值。
	Megapixels string `json:"megapixels,omitempty" validate:"omitempty,oneof=1 0.25 match_input" description:"近似的百万像素，例如 1、0.25 或 match_input。"`

	// === 图像提示和参考 ===
	// Image 是用于 img2img 或 inpainting 的输入图像。
	// 可选，字符串（URI），在某些模型中是必需的（例如 flux-fill-pro）。
	Image string `json:"image,omitempty" validate:"omitempty,url" description:"用于 img2img 或 inpainting 的输入图像 URI。"`

	// ImagePrompt 是用于引导生成的图像 URI。
	// 可选，字符串（URI），在 flux 模型中使用。
	ImagePrompt string `json:"image_prompt,omitempty" validate:"omitempty,url" description:"用于与文本提示一起引导生成的图像 URI。"`

	// ReduxImage 是用于 redux 模式的输入图像。
	// 可选，字符串（URI），在 flux-redux 模型中是必需的。
	ReduxImage string `json:"redux_image,omitempty" validate:"omitempty,url" description:"用于 redux 模式的输入图像 URI，在某些模型中替代 prompt。"`

	// ControlImage 是用于边缘引导或深度感知生成的控制图像。
	// 可选，字符串（URI），在 flux-canny 和 flux-depth 模型中是必需的。
	ControlImage string `json:"control_image,omitempty" validate:"omitempty,url" description:"用于边缘引导或深度感知生成的控制图像 URI。"`

	// Mask 是用于 inpainting 的蒙版图像。
	// 可选，字符串（URI），在 flux-fill 和 ideogram-ai 模型中使用。
	Mask string `json:"mask,omitempty" validate:"omitempty,url" description:"用于 inpainting 的黑白蒙版 URI。黑色保留，白色 inpaint。"`

	// ImageReferenceUrl 是参考图像的 URL。
	// 可选，字符串（URI），在 luma/photon 模型中使用。
	ImageReferenceUrl string `json:"image_reference_url,omitempty" validate:"omitempty,url" description:"用于引导生成的参考图像 URI。"`

	// StyleReferenceUrl 是风格参考图像的 URL。
	// 可选，字符串（URI），在 luma/photon 模型中使用。
	StyleReferenceUrl string `json:"style_reference_url,omitempty" validate:"omitempty,url" description:"风格参考图像 URI。"`

	// CharacterReferenceUrl 是角色参考图像的 URL。
	// 可选，字符串（URI），在 luma/photon 模型中使用。
	CharacterReferenceUrl string `json:"character_reference_url,omitempty" validate:"omitempty,url" description:"角色参考图像 URI。"`

	// === 输出控制 ===
	// OutputFormat 是输出图像的格式。
	// 可选，字符串，涵盖所有枚举值（webp, jpg, png）。
	OutputFormat string `json:"output_format,omitempty" validate:"omitempty,oneof=webp jpg png" description:"输出图像格式：webp, jpg, 或 png。"`

	// OutputQuality 是输出图像的质量。
	// 可选，整数，范围 0-100，默认值因模型而异。
	OutputQuality int `json:"output_quality,omitempty" validate:"omitempty,min=0,max=100" description:"输出质量，0-100，100 最佳。"`

	// NumOutputs 是要生成的图像数量。
	// 可选，整数，范围 1-9，在大多数模型中默认值为 1。
	NumOutputs int `json:"num_outputs,omitempty" validate:"omitempty,min=1,max=9" description:"要生成的图像数量，1-9。"`

	// NumberOfImages 是要生成的图像数量（替代命名）。
	// 可选，整数，范围 1-9，在 minimax/image-01 中使用。
	NumberOfImages int `json:"number_of_images,omitempty" validate:"omitempty,min=1,max=9" description:"要生成的图像数量，1-9（num_outputs 的替代）。"`

	// === 模型特定参数 ===
	// Cfg 是引导比例。
	// 可选，float64，范围 0-20，默认值因模型而异。
	Cfg float64 `json:"cfg,omitempty" validate:"omitempty,min=0,max=20" description:"引导比例，0-20，控制提示的遵循程度。"`

	// Steps 是扩散/推理步数。
	// 可选，整数，范围 1-50，默认值因模型而异。
	Steps int `json:"steps,omitempty" validate:"omitempty,min=1,max=50" description:"扩散步数，1-50。"`

	// Guidance 控制提示的遵循程度与创造力。
	// 可选，float64，范围 0-100，默认值因模型而异。
	Guidance float64 `json:"guidance,omitempty" validate:"omitempty,min=0,max=100" description:"引导提示的遵循程度，0-100。"`

	// PromptUpsampling 启用自动提示增强。
	// 可选，布尔值，在大多数模型中默认为 false。
	PromptUpsampling bool `json:"prompt_upsampling,omitempty" description:"启用提示增强以提高创造力。"`

	// SafetyTolerance 控制内容安全级别。
	// 可选，整数，范围 1-6，在 flux 模型中默认为 2。
	SafetyTolerance int `json:"safety_tolerance,omitempty" validate:"omitempty,min=1,max=6" description:"安全容忍度，1-6，1 最严格。"`

	// DisableSafetyChecker 禁用安全检查器。
	// 可选，布尔值，在 flux 模型中默认为 false。
	DisableSafetyChecker bool `json:"disable_safety_checker,omitempty" description:"禁用生成图像的安全检查器。"`

	// PromptStrength 是 img2img 模式中的提示强度。
	// 可选，float64，范围 0-1，在许多模型中默认为 0.85。
	PromptStrength float64 `json:"prompt_strength,omitempty" validate:"omitempty,min=0,max=1" description:"img2img 模式中的提示强度，0-1。"`

	// NumInferenceSteps 是去噪步数。
	// 可选，整数，范围 1-50，默认值因模型而异。
	NumInferenceSteps int `json:"num_inference_steps,omitempty" validate:"omitempty,min=1,max=50" description:"去噪步数，1-50。"`

	// GoFast 启用更快的预测模式。
	// 可选，布尔值，在某些 flux 模型中默认为 true。
	GoFast bool `json:"go_fast,omitempty" description:"启用更快的预测模式。"`

	// LoraScale 确定 LoRA 应用的强度。
	// 可选，float64，范围 -1 到 3，在 flux-lora 模型中默认为 1。
	LoraScale float64 `json:"lora_scale,omitempty" validate:"omitempty,min=-1,max=3" description:"LoRA 应用强度，-1 到 3。"`

	// LoraWeights 指定 LoRA 权重的 URI 或标识符。
	// 可选，字符串，在 flux-lora 模型中使用。
	LoraWeights string `json:"lora_weights,omitempty" description:"LoRA 权重的 URI 或标识符。"`

	// Interval 增加输出方差。
	// 可选，float64，范围 1-4，在 flux-pro 中默认为 2。
	Interval float64 `json:"interval,omitempty" validate:"omitempty,min=1,max=4" description:"输出方差控制，1-4。"`

	// Outpaint 指定外延选项。
	// 可选，字符串，flux-fill-pro 的枚举值。
	Outpaint string `json:"outpaint,omitempty" validate:"omitempty,oneof=None 'Zoom out 1.5x' 'Zoom out 2x' 'Make square' 'Left outpaint' 'Right outpaint' 'Top outpaint' 'Bottom outpaint'" description:"外延选项。"`

	// Style 指定生成风格。
	// 可选，字符串，recraft-ai 模型的枚举值。
	Style string `json:"style,omitempty" validate:"omitempty,oneof=realistic_image 'realistic_image/b_and_w' 'realistic_image/enterprise' 'realistic_image/hard_flash' 'realistic_image/hdr' 'realistic_image/motion_blur' 'realistic_image/natural_light' 'realistic_image/studio_portrait' digital_illustration 'digital_illustration/2d_art_poster' 'digital_illustration/2d_art_poster_2' 'digital_illustration/3d' 'digital_illustration/80s' 'digital_illustration/engraving_color' 'digital_illustration/glow' 'digital_illustration/grain' 'digital_illustration/hand_drawn' 'digital_illustration/hand_drawn_outline' 'digital_illustration/handmade_3d' 'digital_illustration/infantile_sketch' 'digital_illustration/kawaii' 'digital_illustration/pixel_art' 'digital_illustration/psychedelic' 'digital_illustration/seamless' 'digital_illustration/voxel' 'digital_illustration/watercolor' vector_illustration 'vector_illustration/cartoon' 'vector_illustration/doodle_line_art' engraving 'vector_illustration/flat_2' 'vector_illustration/kawaii' line_art line_circuit linocut seamless icon 'icon/broken_line' 'icon/colored_outline' 'icon/colored_shapes' 'icon/colored_shapes_gradient' 'icon/doodle_fill' 'icon/doodle_offset_fill' 'icon/offset_fill' outline 'icon/outline_gradient' 'icon/uneven_fill' any" description:"生成风格，因模型而异。"`

	// StyleType 指定美学风格。
	// 可选，字符串，ideogram-ai 模型的枚举值。
	StyleType string `json:"style_type,omitempty" validate:"omitempty,oneof=None Auto General Realistic Design 'Render 3D' Anime" description:"ideogram-ai 模型的美学风格。"`

	// MagicPromptOption 控制提示优化。
	// 可选，字符串，ideogram-ai 模型的枚举值。
	MagicPromptOption string `json:"magic_prompt_option,omitempty" validate:"omitempty,oneof=Auto On Off" description:"提示优化模式。"`

	// PromptOptimizer 启用提示优化。
	// 可选，布尔值，在 minimax/image-01 中默认为 true。
	PromptOptimizer bool `json:"prompt_optimizer,omitempty" description:"启用提示优化。"`

	// ImagePromptStrength 混合文本和图像提示。
	// 可选，float64，范围 0-1，在 flux-1.1-pro-ultra 中默认为 0.1。
	ImagePromptStrength float64 `json:"image_prompt_strength,omitempty" validate:"omitempty,min=0,max=1" description:"文本和图像提示的混合强度，0-1。"`

	// ImageReferenceWeight 控制参考图像的影响。
	// 可选，float64，范围 0-1，在 luma/photon 中默认为 0.85。
	ImageReferenceWeight float64 `json:"image_reference_weight,omitempty" validate:"omitempty,min=0,max=1" description:"参考图像的影响，0-1。"`

	// StyleReferenceWeight 控制风格参考的影响。
	// 可选，float64，范围 0-1，在 luma/photon 中默认为 0.85。
	StyleReferenceWeight float64 `json:"style_reference_weight,omitempty" validate:"omitempty,min=0,max=1" description:"风格参考的影响，0-1。"`

	// Raw 生成较少处理的图像。
	// 可选，布尔值，在 flux-1.1-pro-ultra 中默认为 false。
	Raw bool `json:"raw,omitempty" description:"生成较少处理、自然外观的图像。"`

	// SafetyFilterLevel 控制安全过滤级别。
	// 可选，字符串，google/imagen-3 的枚举值。
	SafetyFilterLevel string `json:"safety_filter_level,omitempty" validate:"omitempty,oneof=block_low_and_above block_medium_and_above block_only_high" description:"安全过滤级别：从严格到宽松。"`
}

// InpaintingImageByFlusReplicateRequest is request to inpainting image by flux pro
//
// https://replicate.com/black-forest-labs/flux-fill-pro/api/schema
type InpaintingImageByFlusReplicateRequest struct {
	Input FluxInpaintingInput `json:"input"`
}

// FluxInpaintingInput is input of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-fill-pro/api/schema
type FluxInpaintingInput struct {
	Mask             string `json:"mask" binding:"required"`
	Image            string `json:"image" binding:"required"`
	Seed             int    `json:"seed"`
	Steps            int    `json:"steps" binding:"required,min=1"`
	Prompt           string `json:"prompt" binding:"required,min=5"`
	Guidance         int    `json:"guidance" binding:"required,min=2,max=5"`
	OutputFormat     string `json:"output_format"`
	SafetyTolerance  int    `json:"safety_tolerance" binding:"required,min=1,max=5"`
	PromptUnsampling bool   `json:"prompt_unsampling"`
}

// replicateImageResponse is response of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-pro?prediction=kg1krwsdf9rg80ch1sgsrgq7h8&output=json
type replicateImageResponse struct {
	CompletedAt time.Time        `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	DataRemoved bool             `json:"data_removed"`
	Error       string           `json:"error"`
	ID          string           `json:"id"`
	Input       DrawImageRequest `json:"input"`
	Logs        string           `json:"logs"`
	Metrics     FluxMetrics      `json:"metrics"`
	// Output could be `string` or `[]string`
	Output    any       `json:"output"`
	StartedAt time.Time `json:"started_at"`
	Status    string    `json:"status"`
	URLs      FluxURLs  `json:"urls"`
	Version   string    `json:"version"`
}

func (r *replicateImageResponse) GetOutput() ([]string, error) {
	switch v := r.Output.(type) {
	case string:
		return []string{v}, nil
	case []string:
		return v, nil
	case nil:
		return nil, nil
	case []interface{}:
		// convert []interface{} to []string
		ret := make([]string, len(v))
		for idx, vv := range v {
			if vvv, ok := vv.(string); ok {
				ret[idx] = vvv
			} else {
				return nil, errors.Errorf("unknown output type: [%T]%v", vv, vv)
			}
		}

		return ret, nil
	default:
		return nil, errors.Errorf("unknown output type: [%T]%v", r.Output, r.Output)
	}
}

// FluxMetrics is metrics of ImageResponse
type FluxMetrics struct {
	ImageCount  int     `json:"image_count"`
	PredictTime float64 `json:"predict_time"`
	TotalTime   float64 `json:"total_time"`
}

// FluxURLs is urls of ImageResponse
type FluxURLs struct {
	Get    string `json:"get"`
	Cancel string `json:"cancel"`
}

type ReplicateChatRequest struct {
	Input ChatInput `json:"input" form:"input" binding:"required"`
}

// ChatInput is input of ChatByReplicateRequest
//
// https://replicate.com/meta/meta-llama-3.1-405b-instruct/api/schema
type ChatInput struct {
	TopK             int     `json:"top_k"`
	TopP             float64 `json:"top_p"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	MinTokens        int     `json:"min_tokens"`
	Temperature      float64 `json:"temperature"`
	SystemPrompt     string  `json:"system_prompt"`
	StopSequences    string  `json:"stop_sequences"`
	PromptTemplate   string  `json:"prompt_template"`
	PresencePenalty  float64 `json:"presence_penalty"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
}

// ChatResponse is response of ChatByReplicateRequest
//
// https://replicate.com/meta/meta-llama-3.1-405b-instruct/examples?input=http&output=json
type ChatResponse struct {
	CompletedAt time.Time   `json:"completed_at"`
	CreatedAt   time.Time   `json:"created_at"`
	DataRemoved bool        `json:"data_removed"`
	Error       string      `json:"error"`
	ID          string      `json:"id"`
	Input       ChatInput   `json:"input"`
	Logs        string      `json:"logs"`
	Metrics     FluxMetrics `json:"metrics"`
	// Output could be `string` or `[]string`
	Output    []string        `json:"output"`
	StartedAt time.Time       `json:"started_at"`
	Status    string          `json:"status"`
	URLs      ChatResponseUrl `json:"urls"`
	Version   string          `json:"version"`
}

// ChatResponseUrl is task urls of ChatResponse
type ChatResponseUrl struct {
	Stream string `json:"stream"`
	Get    string `json:"get"`
	Cancel string `json:"cancel"`
}
