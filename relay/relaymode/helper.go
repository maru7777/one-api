package relaymode

import "strings"

func GetByPath(path string) int {
	relayMode := Unknown
	if strings.HasPrefix(path, "/v1/oneapi/proxy") {
		relayMode = Proxy
	} else if strings.HasPrefix(path, "/v1/responses") {
		relayMode = ResponseAPI
	} else if strings.HasPrefix(path, "/v1/chat/completions") {
		relayMode = ChatCompletions
	} else if strings.HasPrefix(path, "/v1/completions") {
		relayMode = Completions
	} else if strings.HasPrefix(path, "/v1/embeddings") {
		relayMode = Embeddings
	} else if strings.HasPrefix(path, "/v1/rerank") {
		relayMode = Rerank
	} else if strings.HasSuffix(path, "/rerank") {
		relayMode = Rerank
	} else if strings.HasSuffix(path, "/rerankers") {
		relayMode = Rerank
	} else if strings.HasSuffix(path, "embeddings") {
		relayMode = Embeddings
	} else if strings.HasPrefix(path, "/v1/moderations") {
		relayMode = Moderations
	} else if strings.HasPrefix(path, "/v1/images/generations") {
		relayMode = ImagesGenerations
	} else if strings.HasPrefix(path, "/v1/edits") {
		relayMode = Edits
	} else if strings.HasPrefix(path, "/v1/audio/speech") {
		relayMode = AudioSpeech
	} else if strings.HasPrefix(path, "/v1/audio/transcriptions") {
		relayMode = AudioTranscription
	} else if strings.HasPrefix(path, "/v1/audio/translations") {
		relayMode = AudioTranslation
	} else if strings.HasPrefix(path, "/v1/images/edits") {
		relayMode = ImagesEdits
	} else if strings.HasPrefix(path, "/v1/oneapi/proxy") {
		relayMode = Proxy
	}
	// else if strings.Contains(path, "/aigc/text2image/") || strings.Contains(path, "/aigc/image2image/") || strings.Contains(path, "/aigc/image-generation/generation") || strings.Contains(path, "/aigc/background-generation/generation") || strings.Contains(path, "/aigc/virtualmodel/generation") || strings.Contains(path, "/vision/image-process/process") || strings.Contains(path, "/aigc/album/") || strings.Contains(path, "/aigc/wordart/") {
	// 	relayMode = ImagesGenerations // 阿里图片。本打算同时支持纯原生接口.但是确定从通用请求转专用请求的方案确定后，恐怕暂时不支持了
	// }
	return relayMode
}
