package ali

import (
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type Input struct {
	//Prompt   string       `json:"prompt"`
	Messages []Message `json:"messages"`
}

type Parameters struct {
	TopP              *float64     `json:"top_p,omitempty"`
	TopK              int          `json:"top_k,omitempty"`
	Seed              uint64       `json:"seed,omitempty"`
	EnableSearch      bool         `json:"enable_search,omitempty"`
	IncrementalOutput bool         `json:"incremental_output,omitempty"`
	MaxTokens         int          `json:"max_tokens,omitempty"`
	Temperature       *float64     `json:"temperature,omitempty"`
	ResultFormat      string       `json:"result_format,omitempty"`
	Tools             []model.Tool `json:"tools,omitempty"`
}

type ChatRequest struct {
	Model      string     `json:"model"`
	Input      Input      `json:"input"`
	Parameters Parameters `json:"parameters,omitempty"`
}

type AliImageRequest struct {
	Model           string           `json:"model" validate:"required"`                   // 模型名称，例如 "wanx2.1-t2i-turbo", "facechain-generation"
	Input           ImageInput       `json:"input" validate:"required"`                   // 输入参数对象
	Parameters      *ImageParameters `json:"parameters,omitempty"`                        // 处理参数对象，可选
	Resources       *[]Resource      `json:"resources,omitempty" validate:"dive"`         // 资源列表，例如 "facechain-generation"，与 input/parameters 同级
	TrainingFileIds *[]string        `json:"training_file_ids,omitempty" validate:"dive"` // 训练文件ID列表，例如 "facechain-finetune" 模型使用
	// ResponseFormat  string           `json:"response_format,omitempty"`                // 响应格式，已注释，无需修改
}

// Input 定义输入参数对象，包含所有可能的输入字段
type ImageInput struct {
	// --- 通用文本提示字段 ---
	Prompt         string `json:"prompt" validate:"required,lte=800"`           // 正向提示词，长度≤800字符
	NegativePrompt string `json:"negative_prompt,omitempty" validate:"lte=500"` // 反向提示词，长度≤500字符，可选
	FacePrompt     string `json:"face_prompt,omitempty" validate:"lte=100"`     // 人脸描述提示词，长度≤100字符，例如 "wanx-virtualmodel"
	RefPrompt      string `json:"ref_prompt,omitempty" validate:"lte=120"`      // 参考提示词，中文≤120字符，例如 "wanx-background-generation-v2"
	PromptTextZh   string `json:"prompt_text_zh,omitempty" validate:"lte=50"`   // 中文提示词，长度≤50字符，例如 "wanx-poster-generation-v1"
	PromptTextEn   string `json:"prompt_text_en,omitempty" validate:"lte=50"`   // 英文提示词，长度≤50字符，例如 "wanx-poster-generation-v1"

	// --- 图像编辑功能字段 ---
	Function string `json:"function,omitempty" validate:"oneof=stylization_all stylization_local description_edit_with_mask remove_watermark expand super_resolution colorization doodle control_cartoon_feature"` // 图像编辑功能，例如 "wanx2.1-imageedit"

	// --- 图像URL字段 ---
	BaseImageUrl       string   `json:"base_image_url,omitempty" validate:"url"`       // 基础图像URL，格式支持 JPEG/PNG等，分辨率 [256x256, 4096x4096]
	MaskImageUrl       string   `json:"mask_image_url,omitempty" validate:"url"`       // 掩码图像URL，分辨率需与 base_image_url 一致
	SketchImageUrl     string   `json:"sketch_image_url,omitempty" validate:"url"`     // 草图图像URL，分辨率 [256x256, 2048x2048]
	ImageUrl           string   `json:"image_url,omitempty" validate:"url"`            // 输入图像URL，分辨率 [256x256, 5760x3240]
	RefImageUrl        string   `json:"ref_image_url,omitempty" validate:"url"`        // 参考图像URL，分辨率 [256x256, 5760x3240]
	TemplateImageUrl   string   `json:"template_image_url,omitempty" validate:"url"`   // 模板图像URL，例如 "shoemodel-v1"
	ShoeImageUrl       []string `json:"shoe_image_url,omitempty" validate:"dive,url"`  // 鞋靴图像URL列表，长度≤3
	TopGarmentUrl      string   `json:"top_garment_url,omitempty" validate:"url"`      // 上装服饰URL，例如 "aitryon"
	BottomGarmentUrl   string   `json:"bottom_garment_url,omitempty" validate:"url"`   // 下装服饰URL，例如 "aitryon"
	PersonImageUrl     string   `json:"person_image_url,omitempty" validate:"url"`     // 人物图像URL，例如 "aitryon"
	CoarseImageUrl     string   `json:"coarse_image_url,omitempty" validate:"url"`     // 粗糙图像URL，例如 "aitryon-refiner"
	FaceImageUrl       string   `json:"face_image_url,omitempty" validate:"url"`       // 人脸图像URL，例如 "wanx-style-cosplay-v1"
	BackgroundImageUrl string   `json:"background_image_url,omitempty" validate:"url"` // 背景图像URL，例如 "wanx-virtualmodel"
	MaskUrl            string   `json:"mask_url,omitempty" validate:"url"`             // 擦除掩码URL，例如 "image-erase-completion"
	ForegroundUrl      string   `json:"foreground_url,omitempty" validate:"url"`       // 保留区域掩码URL，例如 "image-erase-completion"
	TemplateUrl        string   `json:"template_url,omitempty" validate:"url"`         // 模板URL，例如 "facechain-generation"
	UserUrls           []string `json:"user_urls,omitempty" validate:"dive,url,max=5"` // 用户图像URL列表，长度≤5，例如 "facechain-generation"
	Images             []string `json:"images,omitempty" validate:"dive,url"`          // 图像URL列表，例如 "facechain-facedetect"
	Logo               string   `json:"logo,omitempty" validate:"url"`                 // Logo URL，分辨率≤1280x1280，例如 "wanx-ast"
	InitImage          string   `json:"init_image,omitempty" validate:"url"`           // 初始参考图像URL，例如 "stable-diffusion-xl"

	// --- 风格和索引字段 ---
	StyleIndex int `json:"style_index,omitempty" validate:"min=-1,max=9"` // 风格索引，[-1, 9]，例如 "wanx-style-repaint-v1"
	ModelIndex int `json:"model_index,omitempty" validate:"min=1"`        // 模型索引，例如 "wanx-style-cosplay-v1"

	// --- 文本和标题字段 ---
	Title    any    `json:"title,omitempty"`                                                                                                                                                                                                                   // 标题，可为 string 或 []string，建议长度≤30字符
	Subtitle any    `json:"subtitle,omitempty"`                                                                                                                                                                                                                // 副标题，可为 string 或 []string，建议长度≤30字符
	Text     any    `json:"text,omitempty"`                                                                                                                                                                                                                    // 文本，可为 []string 或 *TextObject，建议长度≤30字符
	BodyText string `json:"body_text,omitempty" validate:"lte=50"`                                                                                                                                                                                             // 正文，长度≤50字符，例如 "wanx-poster-generation-v1"
	Surname  string `json:"surname,omitempty" validate:"lte=2"`                                                                                                                                                                                                // 姓氏，长度1-2字符，例如 "wordart-surnames"
	Style    string `json:"style,omitempty" validate:"oneof=diy fantasy_pavilion peerless_beauty landscape_pavilion traditional_buildings green_dragon_girl cherry_blossoms lovely_girl ink_hero anime_girl lake_pavilion tranquil_countryside dusk_splendor"` // 风格，例如 "wordart-surnames"

	// --- 海报和生成控制字段 ---
	GenerateMode string  `json:"generate_mode,omitempty" validate:"oneof=generate sr hrf"`                                                                            // 生成模式，例如 "wanx-poster-generation-v1"
	GenerateNum  int     `json:"generate_num,omitempty" validate:"min=1,max=4"`                                                                                       // 生成数量，[1, 4]
	WhRatios     string  `json:"wh_ratios,omitempty" validate:"oneof=横版 竖版"`                                                                                          // 版式，例如 "wanx-poster-generation-v1"
	LoraName     string  `json:"lora_name,omitempty" validate:"oneof='' 2D插画1 2D插画2 浩瀚星云 浓郁色彩 光线粒子 透明玻璃 剪纸工艺 折纸工艺 中国水墨 中国刺绣 真实场景 2D卡通 儿童水彩 赛博背景 浅蓝抽象 深蓝抽象 抽象点线 童话油画"` // 风格名称，例如 "wanx-poster-generation-vampire"
	LoraWeight   float64 `json:"lora_weight,omitempty" validate:"min=0,max=1"`                                                                                        // 风格权重，[0, 1]
	CtrlRatio    float64 `json:"ctrl_ratio,omitempty" validate:"min=0,max=1"`                                                                                         // 留白效果权重，[0, 1]
	CtrlStep     float64 `json:"ctrl_step,omitempty" validate:"gt=0,lte=1"`                                                                                           // 留白步数比例，(0, 1]

	// --- 边缘引导字段 ---
	ReferenceEdge *ReferenceEdge `json:"reference_edge,omitempty"` // 边缘引导元素，例如 "wanx-background-generation-v2"

	// --- 虚拟模型和高级控制字段 ---
	BgstyleScale float64 `json:"bgstyle_scale,omitempty" validate:"min=0,max=1"`                               // 背景参考权重，[0, 1]，例如 "wanx-virtualmodel"
	RealPerson   bool    `json:"realPerson,omitempty"`                                                         // 是否真人，例如 "wanx-virtualmodel"
	Seed         int     `json:"seed,omitempty" validate:"min=-1,max=10000000"`                                // 种子值，[-1, 10000000]
	AspectRatio  string  `json:"aspect_ratio,omitempty" validate:"oneof='比例不变' 2:1 16:9 4:3 1:1 3:4 9:16 1:2"` // 长宽比，例如 "wanx-virtualmodel"
	Underlay     int     `json:"underlay,omitempty" validate:"min=0,max=2"`                                    // 蒙版数量，[0, 2]，例如 "wanx-ast"

	// --- 子对象字段 ---
	TextureStyle string `json:"texture_style,omitempty" validate:"oneof=material scene lighting waterfall snow_plateau forest sky chinese_building cartoon lego flower acrylic marble felt oil_painting watercolor_painting chinese_painting claborate_style_painting city_night mountain_lake autumn_leaves green_dragon red_dragon"` // 纹理风格，例如 "wordart-texture"
}

// Parameters 定义处理参数对象，包含所有可能的参数字段
type ImageParameters struct {
	// --- 通用生成参数 ---
	Size     string  `json:"size,omitempty" validate:"oneof=512*512 512x512 512*1024 512x1024 768*512 768x512 768*1024 768x1024 1024*576 1024x576 576*1024 576x1024 1024*1024 1024x1024 720*1280 720x1280 768*1152 768x1152 1280*720 1280x720"`
	N        int     `json:"n,omitempty" validate:"min=1,max=4"`                                                                                                                                                                                                                                                                                                                                                                                                                // 生成数量，[1, 4]
	Style    string  `json:"style,omitempty" validate:"oneof='<auto>' '<photography>' '<portrait>' '<3d cartoon>' '<anime>' '<oil painting>' '<watercolor>' '<sketch>' '<chinese painting>' '<flat illustration>' train_free_portrait_url_template portrait_url_template f_idcard_male f_business_male f_idcard_female f_business_female m_springflower_female f_summersport_female f_autumnleaf_female m_winterchinese_female f_hongkongvintage_female f_lightportray_female"` // 输出风格
	Seed     int     `json:"seed,omitempty" validate:"min=0,max=2147483647"`                                                                                                                                                                                                                                                                                                                                                                                                    // 种子值，[0, 2147483647]
	Steps    int     `json:"steps,omitempty" validate:"min=1,max=500"`                                                                                                                                                                                                                                                                                                                                                                                                          // 推理步数，[1, 500]
	Guidance float64 `json:"guidance,omitempty" validate:"gte=1"`                                                                                                                                                                                                                                                                                                                                                                                                               // 指导度量值，≥1
	Cfg      float64 `json:"cfg,omitempty" validate:"min=4,max=5"`                                                                                                                                                                                                                                                                                                                                                                                                              // 提示贴合度，[4, 5]
	Scale    int     `json:"scale,omitempty" validate:"min=1,max=15"`                                                                                                                                                                                                                                                                                                                                                                                                           // 贴合程度，[1, 15]

	// --- 草图和风格参数 ---
	SketchWeight     int     `json:"sketch_weight,omitempty" validate:"min=0,max=10"`                 // 草图约束程度，[0, 10]
	SketchExtraction bool    `json:"sketch_extraction,omitempty"`                                     // 是否提取草图边缘
	SketchColor      [][]int `json:"sketch_color,omitempty" validate:"dive,len=3,dive,min=0,max=255"` // 草图颜色RGB列表

	// --- 扩图参数 ---
	Angle        int     `json:"angle,omitempty" validate:"min=0,max=359"`                         // 旋转角度，[0, 359]
	XScale       float64 `json:"x_scale,omitempty" validate:"min=1,max=3"`                         // 水平扩展比例，[1, 3]
	YScale       float64 `json:"y_scale,omitempty" validate:"min=1,max=3"`                         // 垂直扩展比例，[1, 3]
	TopScale     float64 `json:"top_scale,omitempty" validate:"min=1,max=2"`                       // 上扩展比例，[1, 2]
	BottomScale  float64 `json:"bottom_scale,omitempty" validate:"min=1,max=2"`                    // 下扩展比例，[1, 2]
	LeftScale    float64 `json:"left_scale,omitempty" validate:"min=1,max=2"`                      // 左扩展比例，[1, 2]
	RightScale   float64 `json:"right_scale,omitempty" validate:"min=1,max=2"`                     // 右扩展比例，[1, 2]
	TopOffset    int     `json:"top_offset,omitempty" validate:"min=0"`                            // 上方像素偏移，≥0
	BottomOffset int     `json:"bottom_offset,omitempty" validate:"min=0"`                         // 下方像素偏移，≥0
	LeftOffset   int     `json:"left_offset,omitempty" validate:"min=0"`                           // 左侧像素偏移，≥0
	RightOffset  int     `json:"right_offset,omitempty" validate:"min=0"`                          // 右侧像素偏移，≥0
	OutputRatio  string  `json:"output_ratio,omitempty" validate:"oneof='' 1:1 3:4 4:3 9:16 16:9"` // 输出宽高比

	// --- 图像质量和水印参数 ---
	BestQuality         bool `json:"best_quality,omitempty"`          // 最佳质量模式
	LimitImageSize      bool `json:"limit_image_size,omitempty"`      // 限制图像大小
	AddWatermark        bool `json:"add_watermark,omitempty"`         // 添加水印
	PromptExtend        bool `json:"prompt_extend,omitempty"`         // 提示智能改写
	Offload             bool `json:"offload,omitempty"`               // 计算卸载
	AddSamplingMetadata bool `json:"add_sampling_metadata,omitempty"` // 添加采样元数据

	// --- 高级控制参数 ---
	ShortSideSize   string   `json:"short_side_size,omitempty" validate:"oneof=512 1024 2048"`                                                                                                                   // 短边大小，例如 "wanx-virtualmodel"
	ModelVersion    string   `json:"model_version,omitempty" validate:"oneof=v2 v3"`                                                                                                                             // 模型版本，例如 "wanx-background-generation-v2"
	RefPromptWeight float64  `json:"ref_prompt_weight,omitempty" validate:"min=0,max=1"`                                                                                                                         // 参考提示权重，[0, 1]
	NoiseLevel      int      `json:"noise_level,omitempty" validate:"min=0,max=999"`                                                                                                                             // 噪声级别，[0, 999]
	RefStrength     float64  `json:"ref_strength,omitempty" validate:"min=0,max=1"`                                                                                                                              // 参考图相似度，[0, 1]
	RefMode         string   `json:"ref_mode,omitempty" validate:"oneof=repaint refonly"`                                                                                                                        // 参考图模式
	UpscaleFactor   int      `json:"upscale_factor,omitempty" validate:"min=1,max=4"`                                                                                                                            // 超分倍数，[1, 4]
	MaskColor       [][]int  `json:"mask_color,omitempty" validate:"dive,len=3,dive,min=0,max=255"`                                                                                                              // 掩码颜色RGB列表
	FastMode        bool     `json:"fast_mode,omitempty"`                                                                                                                                                        // 快速模式
	DilateFlag      bool     `json:"dilate_flag,omitempty"`                                                                                                                                                      // 膨胀标志
	Temperature     float64  `json:"temperature,omitempty" validate:"min=0,max=1"`                                                                                                                               // 采样温度，[0, 1]
	TopP            float64  `json:"top_p,omitempty" validate:"min=0,max=1"`                                                                                                                                     // 核采样阈值，[0, 1]
	Resolution      int      `json:"resolution,omitempty" validate:"oneof=-1 1024 1280"`                                                                                                                         // 输出分辨率，例如 "aitryon"
	RestoreFace     bool     `json:"restore_face,omitempty"`                                                                                                                                                     // 还原人脸
	Gender          string   `json:"gender,omitempty" validate:"oneof=woman man"`                                                                                                                                // 性别
	ClothesType     []string `json:"clothes_type,omitempty" validate:"dive,oneof=upper lower dress"`                                                                                                             // 服饰类型
	ImageShortSize  int      `json:"image_short_size,omitempty" validate:"min=512,max=1024"`                                                                                                                     // 图像短边长度，[512, 1024]
	AlphaChannel    bool     `json:"alpha_channel,omitempty"`                                                                                                                                                    // 是否带alpha通道
	Shift           float64  `json:"shift,omitempty" validate:"gte=0"`                                                                                                                                           // 偏移量，≥0
	FontName        string   `json:"font_name,omitempty" validate:"oneof=dongfangdakai puhuiti_m shuheiti jinbuti kuheiti kuaileti wenyiti logoti cangeryuyangti_m siyuansongti_b siyuanheiti_m fangzhengkaiti"` // 字体名称，例如 "wordart-semantic"
}

// ReferenceEdge 定义边缘引导元素对象
type ReferenceEdge struct {
	ForegroundEdge       []string `json:"foreground_edge,omitempty" validate:"dive,url,max=10"`     // 前景元素URL列表，长度≤10
	BackgroundEdge       []string `json:"background_edge,omitempty" validate:"dive,url,max=10"`     // 背景元素URL列表，长度≤10
	ForegroundEdgePrompt []string `json:"foreground_edge_prompt,omitempty" validate:"dive,lte=120"` // 前景提示词列表，长度≤120字符
	BackgroundEdgePrompt []string `json:"background_edge_prompt,omitempty" validate:"dive,lte=120"` // 背景提示词列表，长度≤120字符
}

// TextObject 定义文本对象，例如 "wordart-texture"
type TextObject struct {
	TextContent      string  `json:"text_content,omitempty" validate:"lte=6"`                                                                                                                                                                            // 文本内容，长度≤6字符
	FontName         string  `json:"font_name,omitempty" validate:"oneof=dongfangdakai puhuiti_m shuheiti jinbuti kuheiti kuaileti wenyiti logoti cangeryuyangti_m siyuansongti_b siyuanheiti_m fangzhengkaiti gufeng1 gufeng2 gufeng3 gufeng4 gufeng5"` // 字体名称
	OutputImageRatio string  `json:"output_image_ratio,omitempty" validate:"oneof=1:1 16:9 9:16"`                                                                                                                                                        // 输出图像宽高比
	TtfUrl           string  `json:"ttf_url,omitempty" validate:"url"`                                                                                                                                                                                   // TTF文件URL
	TextStrength     float64 `json:"text_strength,omitempty" validate:"min=0,max=1"`                                                                                                                                                                     // 字形强度，[0, 1]
	TextInverse      bool    `json:"text_inverse,omitempty"`                                                                                                                                                                                             // 文字亮暗
}

// Resource 定义资源对象，例如 "facechain-generation"
type Resource struct {
	ResourceId   string `json:"resource_id" validate:"required"`      // 资源ID，例如 "women_model"
	ResourceType string `json:"resource_type" validate:"eq=facelora"` // 资源类型，固定为 "facelora"
}

type TaskResponse struct {
	StatusCode int    `json:"status_code,omitempty"`
	RequestId  string `json:"request_id,omitempty"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	Output     struct {
		TaskId     string `json:"task_id,omitempty"`
		TaskStatus string `json:"task_status,omitempty"`
		Code       string `json:"code,omitempty"`
		Message    string `json:"message,omitempty"`
		Results    []struct {
			B64Image string `json:"b64_image,omitempty"`
			Url      string `json:"url,omitempty"`
			Code     string `json:"code,omitempty"`
			Message  string `json:"message,omitempty"`
		} `json:"results,omitempty"`
		TaskMetrics struct {
			Total     int `json:"TOTAL,omitempty"`
			Succeeded int `json:"SUCCEEDED,omitempty"`
			Failed    int `json:"FAILED,omitempty"`
		} `json:"task_metrics,omitempty"`
	} `json:"output,omitempty"`
	Usage Usage `json:"usage"`
}

type Header struct {
	Action       string `json:"action,omitempty"`
	Streaming    string `json:"streaming,omitempty"`
	TaskID       string `json:"task_id,omitempty"`
	Event        string `json:"event,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	Attributes   any    `json:"attributes,omitempty"`
}

type Payload struct {
	Model      string `json:"model,omitempty"`
	Task       string `json:"task,omitempty"`
	TaskGroup  string `json:"task_group,omitempty"`
	Function   string `json:"function,omitempty"`
	Parameters struct {
		SampleRate int     `json:"sample_rate,omitempty"`
		Rate       float64 `json:"rate,omitempty"`
		Format     string  `json:"format,omitempty"`
	} `json:"parameters,omitempty"`
	Input struct {
		Text string `json:"text,omitempty"`
	} `json:"input,omitempty"`
	Usage struct {
		Characters int `json:"characters,omitempty"`
	} `json:"usage,omitempty"`
}

type WSSMessage struct {
	Header  Header  `json:"header,omitempty"`
	Payload Payload `json:"payload,omitempty"`
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Parameters *struct {
		TextType string `json:"text_type,omitempty"`
	} `json:"parameters,omitempty"`
}

type Embedding struct {
	Embedding []float64 `json:"embedding"`
	TextIndex int       `json:"text_index"`
}

type EmbeddingResponse struct {
	Output struct {
		Embeddings []Embedding `json:"embeddings"`
	} `json:"output"`
	Usage Usage `json:"usage"`
	Error
}

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type Output struct {
	//Text         string                      `json:"text"`
	//FinishReason string                      `json:"finish_reason"`
	Choices []openai.TextResponseChoice `json:"choices"`
}

type ChatResponse struct {
	Output Output `json:"output"`
	Usage  Usage  `json:"usage"`
	Error
}
