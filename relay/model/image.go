package model

// LogoRect 标识图片的坐标参数
type LogoRect struct {
	X      int `json:"X,omitempty" form:"X"`           // 左上角X坐标
	Y      int `json:"Y,omitempty" form:"Y"`           // 左上角Y坐标
	Width  int `json:"Width,omitempty" form:"Width"`   // 方框宽度
	Height int `json:"Height,omitempty" form:"Height"` // 方框高度
}

// LogoParam 标识内容设置
type LogoParam struct {
	LogoUrl   string    `json:"LogoUrl,omitempty" form:"LogoUrl"`     // 水印URL
	LogoImage string    `json:"LogoImage,omitempty" form:"LogoImage"` // 水印Base64
	LogoRect  *LogoRect `json:"LogoRect,omitempty" form:"LogoRect"`   // 水印坐标
}

// FaceInfo 结构体（假设已定义）
type FaceInfo struct {
	// 假设包含人脸信息
	FaceId string `json:"FaceId,omitempty"`
}

// Image 图片信息结构体
type Image struct {
	ImageUrl    string `json:"ImageUrl,omitempty" validate:"omitempty,url,endswith=jpg|jpeg|png|bmp|tiff|webp"`
	ImageBase64 string `json:"ImageBase64,omitempty" validate:"omitempty,base64"`
}

// ImageRequest 是一个通用的图像请求结构体，用于接收来自不同渠道的图像请求数据。
type ImageRequest struct {
	// --- 通用参数 ---
	Model   string `json:"model,omitempty" form:"model"`     // 模型，必选
	Action  string `json:"Action,omitempty" form:"Action"`   // 接口动作，必选
	Version string `json:"Version,omitempty" form:"Version"` // 接口版本，必选，支持 2023-09-01 和 2022-12-29
	Region  string `json:"Region,omitempty" form:"Region"`   // 地域，必选，固定为 ap-guangzhou

	// --- 提示词相关 ---
	Prompt         string `json:"Prompt,omitempty" form:"Prompt"`                 // 文本描述，可选，最大 1024 字符（混元生图等接口）
	NegativePrompt string `json:"NegativePrompt,omitempty" form:"NegativePrompt"` // 反向提示词，可选，最大 1024 字符（混元生图等接口）
	RefPrompt      string `json:"ref_prompt,omitempty" form:"ref_prompt"`         // 引导文本提示词，支持中英文
	NegRefPrompt   string `json:"neg_ref_prompt,omitempty" form:"neg_ref_prompt"` // 负向引导提示词
	FacePrompt     string `json:"face_prompt,omitempty" form:"face_prompt"`       // 人像面部描述，适用于虚拟模型
	PromptTextZh   string `json:"prompt_text_zh,omitempty" form:"prompt_text_zh"` // 中文提示词，适用于海报生成
	PromptTextEn   string `json:"prompt_text_en,omitempty" form:"prompt_text_en"` // 英文提示词，适用于海报生成
	Surname        string `json:"surname,omitempty" form:"surname"`               // 姓氏，适用于百家姓生成
	Title          any    `json:"title,omitempty" form:"title"`                   // 主标题，文本数组
	Subtitle       any    `json:"subtitle,omitempty" form:"subtitle"`             // 副标题，文本数组
	BodyText       string `json:"body_text,omitempty" form:"body_text"`           // 正文，适用于海报生成

	// --- 图像输入相关 ---
	ContentImage          *Image    `json:"ContentImage,omitempty" form:"ContentImage"`                       // 参考图，可选
	ImageBase64           string    `json:"ImageBase64,omitempty" form:"ImageBase64"`                         // 输入图 Base64，可选，最大 200 字符
	ImageUrl              string    `json:"ImageUrl,omitempty" form:"ImageUrl"`                               // 输入图 URL，可选，URL 格式，最大 200 字符
	InputImage            string    `json:"InputImage,omitempty" form:"InputImage"`                           // 输入图片 Base64，可选（多接口支持）
	InputUrl              string    `json:"InputUrl,omitempty" form:"InputUrl"`                               // 输入图片 URL，可选（多接口支持）
	BaseImageURL          string    `json:"base_image_url,omitempty" form:"base_image_url"`                   // 基础图像 URL
	MaskImageURL          string    `json:"mask_image_url,omitempty" form:"mask_image_url"`                   // 掩码图像 URL
	SketchImageURL        string    `json:"sketch_image_url,omitempty" form:"sketch_image_url"`               // 草图图像 URL
	RefImageURL           string    `json:"ref_image_url,omitempty" form:"ref_image_url"`                     // 参考图像 URL
	FaceImageURL          string    `json:"face_image_url,omitempty" form:"face_image_url"`                   // 人脸图像 URL
	PersonImageURL        string    `json:"person_image_url,omitempty" form:"person_image_url"`               // 人物图像 URL
	TopGarmentURL         string    `json:"top_garment_url,omitempty" form:"top_garment_url"`                 // 上装服饰 URL
	BottomGarmentURL      string    `json:"bottom_garment_url,omitempty" form:"bottom_garment_url"`           // 下装服饰 URL
	CoarseImageURL        string    `json:"coarse_image_url,omitempty" form:"coarse_image_url"`               // 粗糙图像 URL
	Images                *[]string `json:"images,omitempty" form:"images"`                                   // 图像 URL 列表
	TrainingFileIDs       *[]string `json:"training_file_ids,omitempty" form:"training_file_ids"`             // 训练文件 ID 列表
	UserURLs              *[]string `json:"user_urls,omitempty" form:"user_urls"`                             // 用户图像 URL 列表
	ForegroundURL         string    `json:"foreground_url,omitempty" form:"foreground_url"`                   // 保留区域掩码 URL
	StyleRefURL           string    `json:"style_ref_url,omitempty" form:"style_ref_url"`                     // 风格参考图像 URL
	ImagePrompt           string    `json:"image_prompt,omitempty" form:"image_prompt"`                       // 图像提示 URL
	ControlImage          string    `json:"control_image,omitempty" form:"control_image"`                     // 控制图像 URL
	ReduxImage            string    `json:"redux_image,omitempty" form:"redux_image"`                         // Redux 图像 URL
	ShoeImageURL          *[]string `json:"shoe_image_url,omitempty" form:"shoe_image_url"`                   // 鞋靴图像 URL 列表
	BackgroundImageURL    string    `json:"background_image_url,omitempty" form:"background_image_url"`       // 背景参考图像 URL
	ImageReferenceURL     string    `json:"image_reference_url,omitempty" form:"image_reference_url"`         // 图像参考 URL
	StyleReferenceURL     string    `json:"style_reference_url,omitempty" form:"style_reference_url"`         // 风格参考 URL
	CharacterReferenceURL string    `json:"character_reference_url,omitempty" form:"character_reference_url"` // 角色参考 URL
	InitImage             string    `json:"init_image,omitempty" form:"init_image"`                           // 初始参考图像 URL
	TemplateUrl           string    `json:"TemplateUrl,omitempty" form:"TemplateUrl"`                         // 模板图 URL，可选（AI 美照）
	ProductUrl            string    `json:"ProductUrl,omitempty" form:"ProductUrl"`                           // 商品 URL，可选（商品背景生成）
	MaskUrl               string    `json:"MaskUrl,omitempty" form:"MaskUrl"`                                 // 遮罩 URL，可选（局部消除、商品背景生成）
	Mask                  string    `json:"Mask,omitempty" form:"Mask"`                                       // 遮罩 Base64，可选（局部消除、商品背景生成）
	ClothesUrl            string    `json:"ClothesUrl,omitempty" form:"ClothesUrl"`                           // 服装 URL，可选（模特换装）
	ModelUrl              string    `json:"ModelUrl,omitempty" form:"ModelUrl"`                               // 模特 URL，可选（模特换装）

	// --- 边缘引导相关 ---
	ReferenceEdge *map[string][]string `json:"reference_edge,omitempty" form:"reference_edge"` // 边缘引导元素

	// --- 图像尺寸和比例 ---
	Resolution       string `json:"Resolution,omitempty" form:"Resolution"`                 // 分辨率，可选，支持所有接口的分辨率选项
	Size             string `json:"size,omitempty" form:"size"`                             // 输出图像分辨率
	ShortSideSize    string `json:"short_side_size,omitempty" form:"short_side_size"`       // 短边大小
	OutputImageRatio string `json:"output_image_ratio,omitempty" form:"output_image_ratio"` // 输出图像宽高比
	WhRatios         string `json:"wh_ratios,omitempty" form:"wh_ratios"`                   // 海报版式
	AspectRatio      string `json:"aspect_ratio,omitempty" form:"aspect_ratio"`             // 图像宽高比
	Width            int    `json:"width,omitempty" form:"width"`                           // 宽度
	Height           int    `json:"height,omitempty" form:"height"`                         // 高度
	ImageShortSize   int    `json:"image_short_size,omitempty" form:"image_short_size"`     // 短边长度

	// --- 生成数量和随机性 ---
	Num            int `json:"Num,omitempty" form:"Num"`                           // 生成数量，可选，1~4
	N              int `json:"n,omitempty" form:"n"`                               // 生成图片数量
	GenerateNum    int `json:"generate_num,omitempty" form:"generate_num"`         // 海报生成数量
	NumOutputs     int `json:"num_outputs,omitempty" form:"num_outputs"`           // 输出数量
	NumberOfImages int `json:"number_of_images,omitempty" form:"number_of_images"` // 图像生成数量
	Seed           int `json:"Seed,omitempty" form:"Seed"`                         // 随机种子，可选，非负整数

	// --- 风格和模式 ---
	Style        string   `json:"Style,omitempty" form:"Style"`                 // 绘画风格，可选，风格编号由外部提供
	Styles       []string `json:"Styles,omitempty" form:"Styles"`               // 风格数组，可选（图像风格化等接口）
	StyleIndex   int      `json:"style_index,omitempty" form:"style_index"`     // 人像风格索引
	ModelIndex   int      `json:"model_index,omitempty" form:"model_index"`     // 风格类型索引
	TextureStyle string   `json:"texture_style,omitempty" form:"texture_style"` // 纹理风格
	ModelVersion string   `json:"model_version,omitempty" form:"model_version"` // 模型版本
	GenerateMode string   `json:"generate_mode,omitempty" form:"generate_mode"` // 生成模式
	RefMode      string   `json:"ref_mode,omitempty" form:"ref_mode"`           // 参考图模式
	Function     string   `json:"function,omitempty" form:"function"`           // 图像编辑功能
	StyleType    string   `json:"style_type,omitempty" form:"style_type"`       // 风格类型

	// --- 控制参数 ---
	Clarity           string `json:"Clarity,omitempty" form:"Clarity"`                         // 超分选项，可选，x2 或 x4
	SketchWeight      int    `json:"sketch_weight,omitempty" form:"sketch_weight"`             // 草图约束程度
	NoiseLevel        int    `json:"noise_level,omitempty" form:"noise_level"`                 // 噪声级别
	UpscaleFactor     int    `json:"upscale_factor,omitempty" form:"upscale_factor"`           // 超分放大倍数
	Angle             int    `json:"angle,omitempty" form:"angle"`                             // 旋转角度
	Underlay          int    `json:"underlay,omitempty" form:"underlay"`                       // 蒙版数量
	Steps             int    `json:"steps,omitempty" form:"steps"`                             // 推理步数
	NumInferenceSteps int    `json:"num_inference_steps,omitempty" form:"num_inference_steps"` // 去噪步数

	// --- 浮点控制参数 ---
	Strength             float64 `json:"Strength,omitempty" form:"Strength"`                             // 风格化自由度，可选，0~1（图像风格化）
	XScale               float64 `json:"x_scale,omitempty" form:"x_scale"`                               // 水平扩展比例
	YScale               float64 `json:"y_scale,omitempty" form:"y_scale"`                               // 垂直扩展比例
	TopScale             float64 `json:"top_scale,omitempty" form:"top_scale"`                           // 向上扩展比例
	BottomScale          float64 `json:"bottom_scale,omitempty" form:"bottom_scale"`                     // 向下扩展比例
	LeftScale            float64 `json:"left_scale,omitempty" form:"left_scale"`                         // 向左扩展比例
	RightScale           float64 `json:"right_scale,omitempty" form:"right_scale"`                       // 向右扩展比例
	RefStrength          float64 `json:"ref_strength,omitempty" form:"ref_strength"`                     // 参考图相似度
	RefPromptWeight      float64 `json:"ref_prompt_weight,omitempty" form:"ref_prompt_weight"`           // 引导文本权重
	Temperature          float64 `json:"temperature,omitempty" form:"temperature"`                       // 采样温度
	TopP                 float64 `json:"top_p,omitempty" form:"top_p"`                                   // 核采样概率阈值
	LoraWeight           float64 `json:"lora_weight,omitempty" form:"lora_weight"`                       // LoRA 权重
	CtrlRatio            float64 `json:"ctrl_ratio,omitempty" form:"ctrl_ratio"`                         // 留白效果权重
	CtrlStep             float64 `json:"ctrl_step,omitempty" form:"ctrl_step"`                           // 留白步数比例
	TextStrength         float64 `json:"text_strength,omitempty" form:"text_strength"`                   // 文字强度
	PromptStrength       float64 `json:"prompt_strength,omitempty" form:"prompt_strength"`               // 提示词强度
	ImageReferenceWeight float64 `json:"image_reference_weight,omitempty" form:"image_reference_weight"` // 图像参考权重
	StyleReferenceWeight float64 `json:"style_reference_weight,omitempty" form:"style_reference_weight"` // 风格参考权重
	Guidance             float64 `json:"guidance,omitempty" form:"guidance"`                             // 引导度
	CFG                  float64 `json:"cfg,omitempty" form:"cfg"`                                       // 提示贴合度
	LoraScale            float64 `json:"lora_scale,omitempty" form:"lora_scale"`                         // LoRA 应用强度
	ImagePromptStrength  float64 `json:"image_prompt_strength,omitempty" form:"image_prompt_strength"`   // 图像提示强度
	Interval             float64 `json:"interval,omitempty" form:"interval"`                             // 输出多样性
	Shift                float64 `json:"shift,omitempty" form:"shift"`                                   // 偏移量

	// --- 布尔控制参数 ---
	Revise               int  `json:"Revise,omitempty" form:"Revise"`                                 // Prompt 扩写开关，可选，0 或 1
	LogoAdd              int  `json:"LogoAdd,omitempty" form:"LogoAdd"`                               // 水印标识开关，可选，0 或 1
	BestQuality          bool `json:"best_quality,omitempty" form:"best_quality"`                     // 是否开启最佳质量模式
	LimitImageSize       bool `json:"limit_image_size,omitempty" form:"limit_image_size"`             // 是否限制图像大小
	AddWatermark         bool `json:"add_watermark,omitempty" form:"add_watermark"`                   // 是否添加水印
	PromptExtend         bool `json:"prompt_extend,omitempty" form:"prompt_extend"`                   // 是否智能改写提示词
	Watermark            bool `json:"watermark,omitempty" form:"watermark"`                           // 是否添加 AI 水印
	SketchExtraction     bool `json:"sketch_extraction,omitempty" form:"sketch_extraction"`           // 是否提取草图边缘
	FastMode             bool `json:"fast_mode,omitempty" form:"fast_mode"`                           // 是否快速模式
	DilateFlag           bool `json:"dilate_flag,omitempty" form:"dilate_flag"`                       // 是否膨胀掩码
	RealPerson           bool `json:"real_person,omitempty" form:"real_person"`                       // 输入是否真人
	AlphaChannel         bool `json:"alpha_channel,omitempty" form:"alpha_channel"`                   // 是否返回 alpha 通道
	TextInverse          bool `json:"text_inverse,omitempty" form:"text_inverse"`                     // 文字区域亮暗
	CreativeTitleLayout  bool `json:"creative_title_layout,omitempty" form:"creative_title_layout"`   // 是否创意标题排版
	PromptUpsampling     bool `json:"prompt_upsampling,omitempty" form:"prompt_upsampling"`           // 是否提示词增强
	GoFast               bool `json:"go_fast,omitempty" form:"go_fast"`                               // 是否加速预测
	Offload              bool `json:"offload,omitempty" form:"offload"`                               // 是否卸载到 CPU
	AddSamplingMetadata  bool `json:"add_sampling_metadata,omitempty" form:"add_sampling_metadata"`   // 是否嵌入元数据
	DisableSafetyChecker bool `json:"disable_safety_checker,omitempty" form:"disable_safety_checker"` // 是否禁用安全检查
	PromptOptimizer      bool `json:"prompt_optimizer,omitempty" form:"prompt_optimizer"`             // 是否使用提示优化
	SkinRetouch          bool `json:"skin_retouch,omitempty" form:"skin_retouch"`                     // 是否自动美颜
	EnhanceImage         int  `json:"EnhanceImage,omitempty" form:"EnhanceImage"`                     // 画质增强，可选，0 或 1（图像风格化）
	RestoreFace          int  `json:"RestoreFace,omitempty" form:"RestoreFace"`                       // 面部优化，可选，0~6（图像风格化）

	// --- 任务相关 ---
	ChatId  string `json:"ChatId,omitempty" form:"ChatId"`   // 对话 ID，可选
	JobId   string `json:"JobId,omitempty" form:"JobId"`     // 任务 ID，可选，UUID 格式
	ModelId string `json:"ModelId,omitempty" form:"ModelId"` // 模型 ID，可选（AI 写真相关）

	// --- AI 写真相关 ---
	StyleId    string     `json:"StyleId,omitempty" form:"StyleId"`       // 风格 ID，可选（AI 写真生成）
	ImageNum   int        `json:"ImageNum,omitempty" form:"ImageNum"`     // 生成图片数量，可选，1~4（AI 写真）
	Definition string     `json:"Definition,omitempty" form:"Definition"` // 清晰度，可选（AI 写真）
	FaceInfos  []FaceInfo `json:"FaceInfos,omitempty" form:"FaceInfos"`   // 人脸信息，可选，1~5 个（AI 美照）
	Similarity float64    `json:"Similarity,omitempty" form:"Similarity"` // 相似度，可选，0~1（AI 美照）
	TrainMode  int        `json:"TrainMode,omitempty" form:"TrainMode"`   // 训练模式，可选，0、1、2（AI 写真）
	BaseUrl    string     `json:"BaseUrl,omitempty" form:"BaseUrl"`       // 训练基础 URL，可选（AI 写真）
	Urls       []string   `json:"Urls,omitempty" form:"Urls"`             // 训练图片 URL 数组，可选，19~24 个（AI 写真）

	// --- 表情动图相关 ---
	Pose    string `json:"Pose,omitempty" form:"Pose"`       // 表情动作，可选（表情动图）
	Text    string `json:"Text,omitempty" form:"Text"`       // 自定义文案，可选，最大 10 字符（表情动图）
	Haircut bool   `json:"Haircut,omitempty" form:"Haircut"` // 头发遮罩，可选（表情动图）

	// --- 百变头像相关 ---
	Type   string `json:"Type,omitempty" form:"Type"`     // 图像类型，可选，human 或 pet（百变头像）
	Filter int    `json:"Filter,omitempty" form:"Filter"` // 质量检测，可选，0 或 1（百变头像）

	// --- 商品背景生成相关 ---
	Product            string `json:"Product,omitempty" form:"Product"`                       // 商品描述，可选，最大 50 字符（商品背景生成）
	BackgroundTemplate string `json:"BackgroundTemplate,omitempty" form:"BackgroundTemplate"` // 背景模板，可选（商品背景生成）

	// --- 模特换装相关 ---
	ClothesType string `json:"ClothesType,omitempty" form:"ClothesType"` // 服装类型，可选（模特换装）

	// --- 扩图相关 ---
	Ratio string `json:"Ratio,omitempty" form:"Ratio"` // 扩图比例，可选（扩图）

	// --- 其他参数 ---
	LogoParam           *LogoParam `json:"LogoParam,omitempty" form:"LogoParam"`                       // 水印参数，可选
	RspImgType          string     `json:"RspImgType,omitempty" form:"RspImgType"`                     // 返回图像方式，可选，base64 或 url
	Gender              string     `json:"gender,omitempty" form:"gender"`                             // 性别
	ClothesTypes        *[]string  `json:"clothes_type,omitempty" form:"clothes_type"`                 // 服饰类型
	FontName            string     `json:"font_name,omitempty" form:"font_name"`                       // 字体名称
	TTFURL              string     `json:"ttf_url,omitempty" form:"ttf_url"`                           // TTF 文件 URL
	Logo                string     `json:"logo,omitempty" form:"logo"`                                 // Logo URL
	AuxiliaryParameters string     `json:"auxiliary_parameters,omitempty" form:"auxiliary_parameters"` // 辅助参数
	LoraName            string     `json:"lora_name,omitempty" form:"lora_name"`                       // LoRA 风格名称
	LoraWeights         string     `json:"lora_weights,omitempty" form:"lora_weights"`                 // LoRA 权重 URL
	OutputFormat        string     `json:"output_format,omitempty" form:"output_format"`               // 输出格式
	OutputQuality       int        `json:"output_quality,omitempty" form:"output_quality"`             // 输出质量
	SafetyTolerance     int        `json:"safety_tolerance,omitempty" form:"safety_tolerance"`         // 安全容忍度
	MagicPromptOption   string     `json:"magic_prompt_option,omitempty" form:"magic_prompt_option"`   // 魔法提示选项
	Megapixels          string     `json:"megapixels,omitempty" form:"megapixels"`                     // 像素数量
	SafetyFilterLevel   string     `json:"safety_filter_level,omitempty" form:"safety_filter_level"`   // 安全过滤级别
	Resources           any        `json:"resources,omitempty" form:"resources"`                       // 资源列表

	// --- 扩展偏移参数 ---
	TopOffset    int    `json:"top_offset,omitempty" form:"top_offset"`       // 上方添加像素
	BottomOffset int    `json:"bottom_offset,omitempty" form:"bottom_offset"` // 下方添加像素
	LeftOffset   int    `json:"left_offset,omitempty" form:"left_offset"`     // 左侧添加像素
	RightOffset  int    `json:"right_offset,omitempty" form:"right_offset"`   // 右侧添加像素
	MaskColor    *[]int `json:"mask_color,omitempty" form:"mask_color"`       // 掩码颜色 RGB 列表
	SketchColor  *[]int `json:"sketch_color,omitempty" form:"sketch_color"`   // 草图画笔颜色 RGB 列表

	// --- 新增字段 ---
	ResponseFormat string `json:"response_format,omitempty" form:"response_format"` // 响应格式，解决 vertexai/imagen 错误
	User           string `json:"user,omitempty" form:"user"`                       // 用户标识，解决 zhipu 错误
	Quality        string `json:"quality,omitempty" form:"quality"`                 // 图像质量
	UserId         string `json:"user_id,omitempty" form:"user_id"`                 // 用户标识，zhipu 标记人
}

// // ToFormData converts the ImageRequest to form data
// func (r *ImageRequest) ToFormData() ([]byte, error) {

// }
