package ali

// var ModelList = []string{
// 	"qwen-turbo", "qwen-turbo-latest",
// 	"qwen-plus", "qwen-plus-latest",
// 	"qwen-max", "qwen-max-latest",
// 	"qwen-max-longcontext",
// 	"qwen-vl-max", "qwen-vl-max-latest", "qwen-vl-plus", "qwen-vl-plus-latest",
// 	"qwen-vl-ocr", "qwen-vl-ocr-latest",
// 	"qwen-audio-turbo",
// 	"qwen-math-plus", "qwen-math-plus-latest", "qwen-math-turbo", "qwen-math-turbo-latest",
// 	"qwen-coder-plus", "qwen-coder-plus-latest", "qwen-coder-turbo", "qwen-coder-turbo-latest",
// 	"qwq-32b-preview", "qwen2.5-72b-instruct", "qwen2.5-32b-instruct", "qwen2.5-14b-instruct", "qwen2.5-7b-instruct", "qwen2.5-3b-instruct", "qwen2.5-1.5b-instruct", "qwen2.5-0.5b-instruct",
// 	"qwen2-72b-instruct", "qwen2-57b-a14b-instruct", "qwen2-7b-instruct", "qwen2-1.5b-instruct", "qwen2-0.5b-instruct",
// 	"qwen1.5-110b-chat", "qwen1.5-72b-chat", "qwen1.5-32b-chat", "qwen1.5-14b-chat", "qwen1.5-7b-chat", "qwen1.5-1.8b-chat", "qwen1.5-0.5b-chat",
// 	"qwen-72b-chat", "qwen-14b-chat", "qwen-7b-chat", "qwen-1.8b-chat", "qwen-1.8b-longcontext-chat",
// 	"qvq-72b-preview",
// 	"qwen2.5-vl-72b-instruct", "qwen2.5-vl-7b-instruct", "qwen2.5-vl-2b-instruct", "qwen2.5-vl-1b-instruct", "qwen2.5-vl-0.5b-instruct",
// 	"qwen2-vl-7b-instruct", "qwen2-vl-2b-instruct", "qwen-vl-v1", "qwen-vl-chat-v1",
// 	"qwen2-audio-instruct", "qwen-audio-chat",
// 	"qwen2.5-math-72b-instruct", "qwen2.5-math-7b-instruct", "qwen2.5-math-1.5b-instruct", "qwen2-math-72b-instruct", "qwen2-math-7b-instruct", "qwen2-math-1.5b-instruct",
// 	"qwen2.5-coder-32b-instruct", "qwen2.5-coder-14b-instruct", "qwen2.5-coder-7b-instruct", "qwen2.5-coder-3b-instruct", "qwen2.5-coder-1.5b-instruct", "qwen2.5-coder-0.5b-instruct",
// 	"text-embedding-v1", "text-embedding-v3", "text-embedding-v2", "text-embedding-async-v2", "text-embedding-async-v1",
// 	"ali-stable-diffusion-xl", "ali-stable-diffusion-v1.5", "wanx-v1",
// 	"qwen-mt-plus", "qwen-mt-turbo",
// 	"deepseek-r1", "deepseek-v3", "deepseek-r1-distill-qwen-1.5b", "deepseek-r1-distill-qwen-7b", "deepseek-r1-distill-qwen-14b", "deepseek-r1-distill-qwen-32b", "deepseek-r1-distill-llama-8b", "deepseek-r1-distill-llama-70b",

// 	// 新增图像模型
// 	"wanx2.1-t2i-turbo", "wanx2.1-t2i-plus", "wanx2.0-t2i-turbo",
// 	"wanx-poster-generation-v1",
// 	"stable-diffusion-xl", "stable-diffusion-v1.5", "stable-diffusion-3.5-large", "stable-diffusion-3.5-large-turbo",
// 	"flux-schnell", "flux-dev", "flux-merged",
// 	"wanx-ast",

// 	// 图像编辑模型
// 	"wanx2.1-imageedit", "wanx-sketch-to-image-lite", "wanx-x-painting",
// 	"image-instance-segmentation", "image-erase-completion",
// 	"aitryon", "aitryon-refiner", "aitryon-parsing-v1",

// 	// 风格变换模型
// 	"wanx-style-repaint-v1", "wanx-style-cosplay-v1",

// 	// 特殊图像处理
// 	"image-out-painting",
// 	"wanx-virtualmodel", "virtualmodel-v2", "shoemodel-v1",
// 	"wanx-background-generation-v2",

// 	// 人像和文字艺术
// 	"facechain-generation",
// 	"wordart-semantic", "wordart-texture", "wordart-surnames",
// }

var ModelList = []string{
	// 商业版模型
	"qwen-turbo",
	"qwen-turbo-latest",
	"qwen-plus",
	"qwen-plus-latest",
	"qwen-max",
	"qwen-max-latest",
	//通义千问VL是具有视觉（图像）理解能力的文本生成模型，不仅能进行OCR（图片文字识别），还能进一步总结和推理，例如从商品照片中提取属性，根据习题图进行解题等
	"qwen-vl-max",
	"qwen-vl-max-latest",
	"qwen-vl-plus",
	"qwen-vl-plus-latest",
	"qwen-vl-ocr",
	"qwen-vl-ocr-latest",
	//通义千问Audio是音频理解模型，支持输入多种音频（人类语音、自然音、音乐、歌声）和文本，并输出文本。该模型不仅能对输入的音频进行转录，还具备更深层次的语义理解、情感分析、音频事件检测、语音聊天等能力
	"qwen-audio-turbo",
	//通义千问数学模型是专门用于数学解题的语言模型
	"qwen-math-plus",
	"qwen-math-plus-latest",
	"qwen-math-turbo",
	"qwen-math-turbo-latest",
	// 通义千问代码模型
	"qwen-coder-plus",
	"qwen-coder-plus-latest",
	"qwen-coder-turbo",
	"qwen-coder-turbo-latest",
	//基于通义千问模型优化的机器翻译大语言模型，擅长中英互译、中文与小语种互译、英文与小语种互译，小语种包括日、韩、法、西、德、葡（巴西）、泰、印尼、越、阿等26种。在多语言互译的基础上，提供术语干预、领域提示、记忆库等能力，提升模型在复杂应用场景下的翻译效果
	"qwen-mt-plus",
	"qwen-mt-turbo",
	//文本生成-通义千问-开源版
	"qwq-32b-preview",
	"qwen2.5-72b-instruct",
	"qwen2.5-32b-instruct",
	"qwen2-72b-instruct",
	"qwen2-57b-a14b-instruct",
	"qwen2.5-vl-72b-instruct",
	"qwen2.5-vl-32b-instruct",
	"qwen2.5-math-72b-instruct", //基于Qwen模型构建的专门用于数学解题的语言模型。
	"qvq-72b-preview",           // qvq-72b-preview模型专注于提升视觉推理能力，尤其在数学推理领域
	"qwen2.5-coder-32b-instruct",
	// DeepSeek-R1 在后训练阶段大规模使用了强化学习技术，在仅有极少标注数据的情况下，极大提升了模型推理能力，尤其在数学、代码、自然语言推理等任务上；DeepSeek-V3 为 MoE 模型，671B 参数，激活 37B，在 14.8T token 上进行了预训练，在长文本、代码、数学、百科、中文能力上表现优秀
	"deepseek-r1",
	"deepseek-v3",
	"deepseek-r1-distill-llama-32b",
	"deepseek-r1-distill-qwen-32b",
	//文本向量模型用于将文本转换成一组可以代表文字的数字，适用于搜索、聚类、推荐、分类任务。模型根据输入Token数计费。
	"text-embedding-v1", // ￥0.0007 / 1k tokens
	"text-embedding-v3",
	"text-embedding-v2",
	"text-embedding-async-v2",
	"text-embedding-async-v1",
	// 生图模型
	"ali-stable-diffusion-xl",
	"ali-stable-diffusion-v1.5",
	"wanx-v1",                       // 0.02192, 原作者的是 0.016 * ImageUsdPerPic
	"wanx2.1-t2i-turbo",             // 0.01918
	"wanx2.1-t2i-plus",              // 0.02740
	"wanx2.0-t2i-turbo",             // 0.00548
	"wanx2.1-imageedit",             // 0.01918
	"wanx-sketch-to-image-lite",     // 0.00822
	"wanx-style-repaint-v1",         // 0.01644
	"wanx-background-generation-v2", // 0.01096
	"aitryon",                       // 0.02740
	"aitryon-parsing-v1",            // 0.00055
	"aitryon-refiner",               // 0.04110
	"facechain-generation",          // 0.02466
	"wordart-texture",               // 0.01096
	"wordart-semantic",              // 0.03288
}
var AliModelMapping = map[string]string{
	"ali-stable-diffusion-xl":   "stable-diffusion-xl",
	"ali-stable-diffusion-v1.5": "stable-diffusion-v1.5",
}
