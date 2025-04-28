package tencent

var ModelList = []string{
	"hunyuan-turbo",
	"hunyuan-large",
	"hunyuan-large-longcontext",
	"hunyuan-standard",
	"hunyuan-standard-256K",
	"hunyuan-translation-lite",
	"hunyuan-role",
	"hunyuan-functioncall",
	"hunyuan-code",
	"hunyuan-turbo-vision",
	"hunyuan-vision",
	// 生图模型
	"hunyuan-image",
	"hunyuan-generate-avatar",
	"hunyuan-image-chat",
	"hunyuan-image-toimage",
	"hunyuan-draw-portrait",
	"hunyuan-change-clothes",
	"hunyuan-replace-background",
	"hunyuan-sketch-to-image",
	"hunyuan-refine-image",
	"hunyuan-image-inpainting-removal",
	"hunyuan-image-outpainting",
	// 3D模型
	"hunyuan-to3d",
}

// TaskActionMapping 定义了提交任务和查询任务的映射关系
var TaskActionMapping = map[string]string{
	"SubmitHunyuanImageJob":       "QueryHunyuanImageJob",
	"SubmitHunyuanImageChatJob":   "QueryHunyuanImageChatJob",
	"SubmitHunyuanTo3DJob":        "QueryHunyuanTo3DJob",
	"SubmitTrainPortraitModelJob": "QueryTrainPortraitModelJob",
	"SubmitMemeJob":               "QueryMemeJob",
	"SubmitGlamPicJob":            "QueryGlamPicJob",
}

// modelToActionMap 定义了模型和动作的映射关系
var modelToActionMap = map[string]string{
	// 生图模型
	"hunyuan-image":                    "SubmitHunyuanImageJob",     // 文本生成图像
	"hunyuan-generate-avatar":          "SubmitMemeJob",             // 推测为表情动图生成（头像/表情包）
	"hunyuan-image-chat":               "SubmitHunyuanImageChatJob", // 多轮对话生成图像
	"hunyuan-image-to-image":           "TextToImageLite",           // 轻量版文本生成图像
	"hunyuan-draw-portrait":            "SubmitGlamPicJob",          // AI 美照生成
	"hunyuan-change-clothes":           "ChangeClothes",             // 模特换装
	"hunyuan-replace-background":       "ReplaceBackground",         // 商品背景替换
	"hunyuan-sketch-to-image":          "SketchToImage",             // 线稿生成图像
	"hunyuan-refine-image":             "RefineImage",               // 图像变清晰
	"hunyuan-image-inpainting-removal": "ImageInpaintingRemoval",    // 局部消除
	"hunyuan-image-outpainting":        "ImageOutpainting",          // 图像扩图
	// 3D 模型
	"hunyuan-to3d": "SubmitHunyuanTo3DJob", // 3D 模型生成
}
