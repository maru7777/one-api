package tencent

type Message struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

type ChatRequest struct {
	// Model name, optional values include hunyuan-lite, hunyuan-standard, hunyuan-standard-256K, hunyuan-pro.
	// For descriptions of each model, please read the [Product Overview](https://cloud.tencent.com/document/product/1729/104753).
	//
	// Note:
	// Different models have different pricing. Please refer to the [Purchase Guide](https://cloud.tencent.com/document/product/1729/97731) for details.
	Model *string `json:"Model"`
	// Chat context information.
	// Description:
	// 1. The maximum length is 40, arranged in the array in chronological order from oldest to newest.
	// 2. Message.Role optional values: system, user, assistant.
	//    Among them, the system role is optional. If it exists, it must be at the beginning of the list.
	//    User and assistant must alternate (one question and one answer), starting and ending with user,
	//    and Content cannot be empty. The order of roles is as follows: [system (optional) user assistant user assistant user ...].
	// 3. The total length of Content in Messages cannot exceed the model's length limit
	//    (refer to the [Product Overview](https://cloud.tencent.com/document/product/1729/104753) document).
	//    If it exceeds, the earliest content will be truncated, leaving only the latest content.
	Messages []*Message `json:"Messages"`
	// Stream call switch.
	// Description:
	// 1. If not provided, the default is non-streaming call (false).
	// 2. In streaming calls, results are returned incrementally using the SSE protocol
	//    (the return value is taken from Choices[n].Delta, and incremental data needs to be concatenated to obtain the complete result).
	// 3. In non-streaming calls:
	// The call method is the same as a regular HTTP request.
	// The interface response time is relatively long. **If lower latency is  it is recommended to set this to true**.
	// Only the final result is returned once (the return value is taken from Choices[n].Message).
	//
	// Note:
	// When calling through the SDK, different methods are required to obtain return values for streaming and non-streaming calls.
	// Refer to the comments or examples in the SDK (in the examples/hunyuan/v20230901/ directory of each language SDK code repository).
	Stream *bool `json:"Stream"`
	// Description:
	// 1. Affects the diversity of the output text. The larger the value, the more diverse the generated text.
	// 2. The value range is [0.0, 1.0]. If not provided, the recommended value for each model is used.
	// 3. It is not recommended to use this unless necessary, as unreasonable values can affect the results.
	TopP *float64 `json:"TopP,omitempty"`
	// Description:
	// 1. Higher values make the output more random, while lower values make it more focused and deterministic.
	// 2. The value range is [0.0, 2.0]. If not provided, the recommended value for each model is used.
	// 3. It is not recommended to use this unless necessary, as unreasonable values can affect the results.
	Temperature *float64 `json:"Temperature,omitempty"`
}

type Error struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type Usage struct {
	PromptTokens     int `json:"PromptTokens"`
	CompletionTokens int `json:"CompletionTokens"`
	TotalTokens      int `json:"TotalTokens"`
}

type ResponseChoices struct {
	FinishReason string  `json:"FinishReason,omitempty"` // Stream end flag, "stop" indicates the end packet
	Messages     Message `json:"Message,omitempty"`      // Content, returned in synchronous mode, null in stream mode. The total content supports up to 1024 tokens.
	Delta        Message `json:"Delta,omitempty"`        // Content, returned in stream mode, null in synchronous mode. The total content supports up to 1024 tokens.
}

type ChatResponse struct {
	Choices []ResponseChoices `json:"Choices,omitempty"`   // Results
	Created int64             `json:"Created,omitempty"`   // Unix timestamp string
	Id      string            `json:"Id,omitempty"`        // Session id
	Usage   Usage             `json:"Usage,omitempty"`     // Token count
	Error   Error             `json:"Error,omitempty"`     // Error message. Note: this field may return null, indicating no valid value can be obtained
	Note    string            `json:"Note,omitempty"`      // Comment
	ReqID   string            `json:"RequestId,omitempty"` // Unique request Id, returned with each request. Used for feedback interface parameters
}

type ChatResponseP struct {
	Response ChatResponse `json:"Response,omitempty"`
}

type EmbeddingRequest struct {
	InputList []string `json:"InputList"`
}

type EmbeddingData struct {
	Embedding []float64 `json:"Embedding"`
	Index     int       `json:"Index"`
	Object    string    `json:"Object"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"PromptTokens"`
	TotalTokens  int `json:"TotalTokens"`
}

type EmbeddingResponse struct {
	Data           []EmbeddingData `json:"Data"`
	EmbeddingUsage EmbeddingUsage  `json:"Usage,omitempty"`
	RequestId      string          `json:"RequestId,omitempty"`
	Error          Error           `json:"Error,omitempty"`
}

type EmbeddingResponseP struct {
	Response EmbeddingResponse `json:"Response,omitempty"`
}

// TenentImageRequest 统一请求结构体，涵盖所有混元生图相关接口的输入参数
type TenentImageRequest struct {
	Model              string     `json:"model,omitempty" binding:"omitempty,oneof=hunyuan-image hunyuan-image-chat hunyuan-draw-portrait hunyuan-draw-portrait-chat hunyuan-to3d"`                                                                                                                                                                                                                                                                                                                                                                                // 模型，必选
	Action             string     `json:"Action,omitempty" binding:"omitempty,oneof=SubmitHunyuanImageJob QueryHunyuanImageJob SubmitHunyuanImageChatJob QueryHunyuanImageChatJob TextToImageLite SubmitHunyuanTo3DJob QueryHunyuanTo3DJob ImageToImage GenerateAvatar UploadTrainPortraitImages SubmitDrawPortraitJob QueryDrawPortraitJob SubmitTrainPortraitModelJob QueryTrainPortraitModelJob SubmitMemeJob QueryMemeJob SubmitGlamPicJob QueryGlamPicJob ChangeClothes ReplaceBackground SketchToImage RefineImage ImageInpaintingRemoval ImageOutpainting"` // 接口动作，必选
	Version            string     `json:"Version,omitempty" binding:"omitempty,eq=2023-09-01|eq=2022-12-29"`                                                                                                                                                                                                                                                                                                                                                                                                                                                       // 接口版本，必选，支持 2023-09-01 和 2022-12-29
	Region             string     `json:"Region,omitempty" binding:"omitempty,eq=ap-guangzhou" default:"ap-guangzhou"`                                                                                                                                                                                                                                                                                                                                                                                                                                             // 地域，必选，固定为 ap-guangzhou
	Prompt             string     `json:"Prompt,omitempty" binding:"omitempty,max=1024"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                           // 文本描述，可选，最大 1024 字符（混元生图等接口）
	NegativePrompt     string     `json:"NegativePrompt,omitempty" binding:"omitempty,max=1024"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                   // 反向提示词，可选，最大 1024 字符（混元生图等接口）
	Style              string     `json:"Style,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     // 绘画风格，可选，风格编号由外部提供
	Styles             []string   `json:"Styles,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    // 风格数组，可选（图像风格化等接口）
	Resolution         string     `json:"Resolution,omitempty" binding:"omitempty,oneof=origin 768:768 768:1024 1024:768 1024:1024 720:1280 1280:720 768:1280 1280:768 1080:1920 1920:1080 512:640 1024:1280 2048:2560 1280:1280"`                                                                                                                                                                                                                                                                                                                                 // 分辨率，可选，支持所有接口的分辨率选项
	Num                int        `json:"Num,omitempty" binding:"omitempty,gte=1,lte=4"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                           // 生成数量，可选，1~4
	Clarity            string     `json:"Clarity,omitempty" binding:"omitempty,oneof=x2 x4"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                       // 超分选项，可选，x2 或 x4
	ContentImage       *Image     `json:"ContentImage,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                              // 参考图，可选
	Revise             int        `json:"Revise,omitempty" binding:"omitempty,oneof=0 1"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                          // Prompt 扩写开关，可选，0 或 1
	Seed               int        `json:"Seed,omitempty" binding:"omitempty,gte=0"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                // 随机种子，可选，非负整数
	LogoAdd            int        `json:"LogoAdd,omitempty" binding:"omitempty,oneof=0 1"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                         // 水印标识开关，可选，0 或 1
	LogoParam          *LogoParam `json:"LogoParam,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 // 水印参数，可选
	ChatId             string     `json:"ChatId,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    // 对话 ID，可选
	JobId              string     `json:"JobId,omitempty" binding:"omitempty,uuid"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                // 任务 ID，可选，UUID 格式
	ImageBase64        string     `json:"ImageBase64,omitempty" binding:"omitempty,max=200"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                       // 输入图 Base64，可选，最大 200 字符
	ImageUrl           string     `json:"ImageUrl,omitempty" binding:"omitempty,uri,max=200"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                      // 输入图 URL，可选，URL 格式，最大 200 字符
	RspImgType         string     `json:"RspImgType,omitempty" binding:"omitempty,oneof=base64 url"`                                                                                                                                                                                                                                                                                                                                                                                                                                                               // 返回图像方式，可选，base64 或 url
	InputImage         string     `json:"InputImage,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                // 输入图片 Base64，可选（多接口支持）
	InputUrl           string     `json:"InputUrl,omitempty" binding:"omitempty,uri"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                              // 输入图片 URL，可选（多接口支持）
	ModelId            string     `json:"ModelId,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   // 模型 ID，可选（AI 写真相关）
	StyleId            string     `json:"StyleId,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   // 风格 ID，可选（AI 写真生成）
	ImageNum           int        `json:"ImageNum,omitempty" binding:"omitempty,gte=1,lte=4"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                      // 生成图片数量，可选，1~4（AI 写真）
	Definition         string     `json:"Definition,omitempty" binding:"omitempty,oneof=sd hd hdpro uhd"`                                                                                                                                                                                                                                                                                                                                                                                                                                                          // 清晰度，可选（AI 写真）
	Pose               string     `json:"Pose,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      // 表情动作，可选（表情动图）
	Text               string     `json:"Text,omitempty" binding:"omitempty,max=10"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                               // 自定义文案，可选，最大 10 字符（表情动图）
	Haircut            bool       `json:"Haircut,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   // 头发遮罩，可选（表情动图）
	TemplateUrl        string     `json:"TemplateUrl,omitempty" binding:"omitempty,uri"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                           // 模板图 URL，可选（AI 美照）
	FaceInfos          []FaceInfo `json:"FaceInfos,omitempty" binding:"omitempty,dive,len=1-5"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                    // 人脸信息，可选，1~5 个（AI 美照）
	Similarity         float64    `json:"Similarity,omitempty" binding:"omitempty,gte=0,lte=1"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                    // 相似度，可选，0~1（AI 美照）
	Type               string     `json:"Type,omitempty" binding:"omitempty,oneof=human pet"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                      // 图像类型，可选，human 或 pet（百变头像）
	Filter             int        `json:"Filter,omitempty" binding:"omitempty,oneof=0 1"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                          // 质量检测，可选，0 或 1（百变头像）
	TrainMode          int        `json:"TrainMode,omitempty" binding:"omitempty,oneof=0 1 2"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                     // 训练模式，可选，0、1、2（AI 写真）
	BaseUrl            string     `json:"BaseUrl,omitempty" binding:"omitempty,uri"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                               // 训练基础 URL，可选（AI 写真）
	Urls               []string   `json:"Urls,omitempty" binding:"omitempty,dive,len=19-24"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                       // 训练图片 URL 数组，可选，19~24 个（AI 写真）
	ProductUrl         string     `json:"ProductUrl,omitempty" binding:"omitempty,uri"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                            // 商品 URL，可选（商品背景生成）
	Product            string     `json:"Product,omitempty" binding:"omitempty,max=50"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                            // 商品描述，可选，最大 50 字符（商品背景生成）
	BackgroundTemplate string     `json:"BackgroundTemplate,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                        // 背景模板，可选（商品背景生成）
	MaskUrl            string     `json:"MaskUrl,omitempty" binding:"omitempty,uri"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                               // 遮罩 URL，可选（局部消除、商品背景生成）
	Mask               string     `json:"Mask,omitempty" binding:"omitempty"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      // 遮罩 Base64，可选（局部消除、商品背景生成）
	ClothesUrl         string     `json:"ClothesUrl,omitempty" binding:"omitempty,uri"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                            // 服装 URL，可选（模特换装）
	ClothesType        string     `json:"ClothesType,omitempty" binding:"omitempty,oneof=Upper-body Lower-body Dress"`                                                                                                                                                                                                                                                                                                                                                                                                                                             // 服装类型，可选（模特换装）
	ModelUrl           string     `json:"ModelUrl,omitempty" binding:"omitempty,uri"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                              // 模特 URL，可选（模特换装）
	Strength           float64    `json:"Strength,omitempty" binding:"omitempty,gt=0,lte=1"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                       // 风格化自由度，可选，0~1（图像风格化）
	EnhanceImage       int        `json:"EnhanceImage,omitempty" binding:"omitempty,oneof=0 1"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                    // 画质增强，可选，0 或 1（图像风格化）
	RestoreFace        int        `json:"RestoreFace,omitempty" binding:"omitempty,gte=0,lte=6"`                                                                                                                                                                                                                                                                                                                                                                                                                                                                   // 面部优化，可选，0~6（图像风格化）
	Ratio              string     `json:"Ratio,omitempty" binding:"omitempty,oneof=1:1 4:3 3:4 16:9 9:16"`                                                                                                                                                                                                                                                                                                                                                                                                                                                         // 扩图比例，可选（扩图）
}

// FaceInfo 用于 AI 美照中的人脸信息
type FaceInfo struct {
	ImageUrls []string `json:"ImageUrls" binding:"required,dive,len=1-6"` // 人脸图片 URL 数组，必选，1~6 个
}

// Image 图片信息结构体
type Image struct {
	ImageUrl    string `json:"ImageUrl,omitempty" validate:"omitempty,url,endswith=jpg|jpeg|png|bmp|tiff|webp"`
	ImageBase64 string `json:"ImageBase64,omitempty" validate:"omitempty,base64"`
}

// LogoParam 水印参数结构体
type LogoParam struct {
	LogoUrl   string    `json:"LogoUrl,omitempty" validate:"omitempty,url"`
	LogoImage string    `json:"LogoImage,omitempty" validate:"omitempty,base64"`
	LogoRect  *LogoRect `json:"LogoRect,omitempty" validate:"omitempty"`
}

// LogoRect 水印坐标结构体
type LogoRect struct {
	X      int `json:"X,omitempty" validate:"gte=0"`
	Y      int `json:"Y,omitempty" validate:"gte=0"`
	Width  int `json:"Width,omitempty" validate:"gte=0"`
	Height int `json:"Height,omitempty" validate:"gte=0"`
}

// TencentTaskResponse 统一响应结构体，涵盖所有混元生图相关接口的输出参数
type TencentTaskResponse struct {
	JobId         string    `json:"JobId,omitempty" binding:"omitempty,uuid"`                                          // 任务 ID，可选，UUID 格式
	RequestId     string    `json:"RequestId" binding:"required,uuid"`                                                 // 唯一请求 ID，必选，UUID 格式
	JobStatusCode string    `json:"JobStatusCode,omitempty" binding:"omitempty,oneof=1 2 4 5 INIT WAIT RUN FAIL DONE"` // 任务状态码，可选，支持数字和字符串状态
	JobStatusMsg  string    `json:"JobStatusMsg,omitempty" binding:"omitempty,max=256"`                                // 任务状态描述，可选，最大 256 字符
	JobErrorCode  string    `json:"JobErrorCode,omitempty" binding:"omitempty"`                                        // 错误码，可选
	JobErrorMsg   string    `json:"JobErrorMsg,omitempty" binding:"omitempty,max=256"`                                 // 错误信息，可选，最大 256 字符
	ResultImage   any       `json:"ResultImage,omitempty" binding:"omitempty"`                                         // 生成图，可选，string 或 []string，需断言
	ResultUrls    []string  `json:"ResultUrls,omitempty" binding:"omitempty,dive,uri"`                                 // 生成图 URL 数组，可选（AI 写真等接口）
	MaskImage     string    `json:"MaskImage,omitempty" binding:"omitempty,uri"`                                       // Mask 图 URL，可选（商品背景生成）
	ResultDetails []string  `json:"ResultDetails,omitempty" binding:"omitempty,dive,max=256"`                          // 结果详情，可选，最大 256 字符
	RevisedPrompt []string  `json:"RevisedPrompt,omitempty" binding:"omitempty,dive,max=1024"`                         // 扩写后的 Prompt，可选，最大 1024 字符
	ChatId        string    `json:"ChatId,omitempty" binding:"omitempty"`                                              // 对话 ID，可选
	History       []History `json:"History,omitempty" binding:"omitempty"`                                             // 多轮对话历史，可选
	ResultFile3Ds []File3Ds `json:"ResultFile3Ds,omitempty" binding:"omitempty"`                                       // 3D 文件列表，可选
	Status        string    `json:"Status,omitempty" binding:"omitempty,oneof=WAIT RUN FAIL DONE"`                     // 3D 任务状态，可选
	ErrorCode     string    `json:"ErrorCode,omitempty" binding:"omitempty"`                                           // 3D 任务错误码，可选
	ErrorMessage  string    `json:"ErrorMessage,omitempty" binding:"omitempty,max=256"`                                // 3D 任务错误信息，可选，最大 256 字符
	Error         *ErrorMsg `json:"Error,omitempty" binding:"omitempty"`                                               // 错误信息，可选
}

type History struct {
	ChatId        string `json:"ChatId,omitempty" binding:"omitempty"`                 // 对话 ID，可选
	Prompt        string `json:"Prompt,omitempty" binding:"omitempty,max=1024"`        // 原始 Prompt，可选，最大 1024 字符
	RevisedPrompt string `json:"RevisedPrompt,omitempty" binding:"omitempty,max=1024"` // 扩写 Prompt，可选，最大 1024 字符
	Seed          int    `json:"Seed,omitempty" binding:"omitempty,gte=0"`             // 随机种子，可选，非负整数
}

type File3Ds struct {
	File3D []File3D `json:"File3D,omitempty" binding:"omitempty"` // 3D 文件列表，可选
}

type File3D struct {
	Type string `json:"Type,omitempty" binding:"omitempty,oneof=GIF OBJ"` // 文件格式，可选，值为 GIF 或 OBJ
	Url  string `json:"Url,omitempty" binding:"omitempty,uri"`            // 文件 URL，可选，URL 格式
}

type ErrorMsg struct {
	Code    string `json:"Code" binding:"required"`   // 错误码，必选
	Message string `json:"Message" binding:"max=256"` // 错误信息，必选，最大 256 字符
}
