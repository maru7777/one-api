package ali

var ModelList = []string{
	"qwen-turbo", "qwen-turbo-latest",
	"qwen-plus", "qwen-plus-latest",
	"qwen-max", "qwen-max-latest",
	"qwen-max-longcontext",
	"qwen-vl-max", "qwen-vl-max-latest", "qwen-vl-plus", "qwen-vl-plus-latest",
	"qwen-vl-ocr", "qwen-vl-ocr-latest",
	"qwen-audio-turbo",
	"qwen-math-plus", "qwen-math-plus-latest", "qwen-math-turbo", "qwen-math-turbo-latest",
	"qwen-coder-plus", "qwen-coder-plus-latest", "qwen-coder-turbo", "qwen-coder-turbo-latest",
	"qwq-32b-preview", "qwen2.5-72b-instruct", "qwen2.5-32b-instruct", "qwen2.5-14b-instruct", "qwen2.5-7b-instruct", "qwen2.5-3b-instruct", "qwen2.5-1.5b-instruct", "qwen2.5-0.5b-instruct",
	"qwen2-72b-instruct", "qwen2-57b-a14b-instruct", "qwen2-7b-instruct", "qwen2-1.5b-instruct", "qwen2-0.5b-instruct",
	"qwen1.5-110b-chat", "qwen1.5-72b-chat", "qwen1.5-32b-chat", "qwen1.5-14b-chat", "qwen1.5-7b-chat", "qwen1.5-1.8b-chat", "qwen1.5-0.5b-chat",
	"qwen-72b-chat", "qwen-14b-chat", "qwen-7b-chat", "qwen-1.8b-chat", "qwen-1.8b-longcontext-chat",
	"qvq-72b-preview",
	"qwen2.5-vl-72b-instruct", "qwen2.5-vl-7b-instruct", "qwen2.5-vl-2b-instruct", "qwen2.5-vl-1b-instruct", "qwen2.5-vl-0.5b-instruct",
	"qwen2-vl-7b-instruct", "qwen2-vl-2b-instruct", "qwen-vl-v1", "qwen-vl-chat-v1",
	"qwen2-audio-instruct", "qwen-audio-chat",
	"qwen2.5-math-72b-instruct", "qwen2.5-math-7b-instruct", "qwen2.5-math-1.5b-instruct", "qwen2-math-72b-instruct", "qwen2-math-7b-instruct", "qwen2-math-1.5b-instruct",
	"qwen2.5-coder-32b-instruct", "qwen2.5-coder-14b-instruct", "qwen2.5-coder-7b-instruct", "qwen2.5-coder-3b-instruct", "qwen2.5-coder-1.5b-instruct", "qwen2.5-coder-0.5b-instruct",
	"text-embedding-v1", "text-embedding-v3", "text-embedding-v2", "text-embedding-async-v2", "text-embedding-async-v1",
	"ali-stable-diffusion-xl", "ali-stable-diffusion-v1.5", "wanx-v1",
	"qwen-mt-plus", "qwen-mt-turbo",
	"deepseek-r1", "deepseek-v3", "deepseek-r1-distill-qwen-1.5b", "deepseek-r1-distill-qwen-7b", "deepseek-r1-distill-qwen-14b", "deepseek-r1-distill-qwen-32b", "deepseek-r1-distill-llama-8b", "deepseek-r1-distill-llama-70b",

	// 新增图像模型
	"wanx2.1-t2i-turbo", "wanx2.1-t2i-plus", "wanx2.0-t2i-turbo",
	"wanx-poster-generation-v1",
	"stable-diffusion-xl", "stable-diffusion-v1.5", "stable-diffusion-3.5-large", "stable-diffusion-3.5-large-turbo",
	"flux-schnell", "flux-dev", "flux-merged",
	"wanx-ast",

	// 图像编辑模型
	"wanx2.1-imageedit", "wanx-sketch-to-image-lite", "wanx-x-painting",
	"image-instance-segmentation", "image-erase-completion",
	"aitryon", "aitryon-refiner", "aitryon-parsing-v1",

	// 风格变换模型
	"wanx-style-repaint-v1", "wanx-style-cosplay-v1",

	// 特殊图像处理
	"image-out-painting",
	"wanx-virtualmodel", "virtualmodel-v2", "shoemodel-v1",
	"wanx-background-generation-v2",

	// 人像和文字艺术
	"facechain-generation",
	"wordart-semantic", "wordart-texture", "wordart-surnames",
}

var AliModelMapping = map[string]string{
	"ali-stable-diffusion-xl":   "stable-diffusion-xl",
	"ali-stable-diffusion-v1.5": "stable-diffusion-v1.5",
}
