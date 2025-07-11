package tencent

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

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

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Tencent pricing: https://cloud.tencent.com/document/product/1729/97731
var ModelRatios = map[string]adaptor.ModelPrice{
	// Hunyuan Models - Based on https://cloud.tencent.com/document/product/1729/97731
	"hunyuan-lite":          {Ratio: 0.75 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-standard":      {Ratio: 4.5 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-standard-256K": {Ratio: 15 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-pro":           {Ratio: 30 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-vision":        {Ratio: 18 * ratio.MilliTokensRmb, CompletionRatio: 1},
	"hunyuan-embedding":     {Ratio: 0.7 * ratio.MilliTokensRmb, CompletionRatio: 1},
	// 新增的文本生成模型
	"hunyuan-turbo":             {Ratio: 0.015 * ratio.MilliTokensRmb, CompletionRatio: 1.0}, // ¥0.015/千tokens
	"hunyuan-large":             {Ratio: 0.004 * ratio.MilliTokensRmb, CompletionRatio: 1.0}, // ¥0.004/千tokens
	"hunyuan-large-longcontext": {Ratio: 0.006 * ratio.MilliTokensRmb, CompletionRatio: 1.0}, // ¥0.006/千tokens
	"hunyuan-translation-lite":  {Ratio: 0.005 * ratio.MilliTokensRmb, CompletionRatio: 1.0}, // ¥0.005/千tokens
	"hunyuan-role":              {Ratio: 0.004 * ratio.MilliTokensRmb, CompletionRatio: 1.0}, // ¥0.004/千tokens
	"hunyuan-functioncall":      {Ratio: 0.004 * ratio.MilliTokensRmb, CompletionRatio: 1.0}, // ¥0.004/千tokens
	"hunyuan-code":              {Ratio: 0.004 * ratio.MilliTokensRmb, CompletionRatio: 1.0}, // ¥0.004/千tokens
	// 新增的视觉模型
	"hunyuan-turbo-vision": {Ratio: 0.08 * ratio.MilliTokensRmb, CompletionRatio: 1.0}, // ¥0.08/千tokens
	// 新增的图像生成模型 - 按张计费
	"hunyuan-image":                    {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-image-chat":               {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-draw-portrait":            {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-draw-portrait-chat":       {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-generate-avatar":          {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-image-toimage":            {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-change-clothes":           {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-replace-background":       {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-sketch-to-image":          {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-refine-image":             {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-image-inpainting-removal": {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	"hunyuan-image-outpainting":        {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
	// 新增的3D模型生成
	"hunyuan-to3d": {Ratio: 0.04 * ratio.ImageRmbPerPic, CompletionRatio: 1.0}, // ¥0.04/张
}
