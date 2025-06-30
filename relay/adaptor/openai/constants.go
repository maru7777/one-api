package openai

var ModelList = []string{
	// -------------------------------------
	// Chat
	// -------------------------------------
	"gpt-3.5-turbo", "gpt-3.5-turbo-0301", "gpt-3.5-turbo-0613", "gpt-3.5-turbo-1106", "gpt-3.5-turbo-0125",
	"gpt-3.5-turbo-16k", "gpt-3.5-turbo-16k-0613",
	"gpt-3.5-turbo-instruct",
	"gpt-4", "gpt-4-0314", "gpt-4-0613", "gpt-4-1106-preview", "gpt-4-0125-preview",
	"gpt-4-32k", "gpt-4-32k-0314", "gpt-4-32k-0613",
	"gpt-4-turbo-preview", "gpt-4-turbo", "gpt-4-turbo-2024-04-09",
	"gpt-4o", "gpt-4o-2024-05-13", "gpt-4o-2024-08-06", "gpt-4o-2024-11-20", "chatgpt-4o-latest",
	"gpt-4o-mini", "gpt-4o-mini-2024-07-18",
	"gpt-4o-mini-audio-preview", "gpt-4o-mini-audio-preview-2024-12-17",
	"gpt-4o-audio-preview", "gpt-4o-audio-preview-2024-12-17", "gpt-4o-audio-preview-2024-10-01",
	"gpt-4-vision-preview",
	"gpt-4.5-preview", "gpt-4.5-preview-2025-02-27",
	"gpt-4.1", "gpt-4.1-2025-04-14",
	"gpt-4.1-mini", "gpt-4.1-mini-2025-04-14",
	"gpt-4.1-nano", "gpt-4.1-nano-2025-04-14",
	"o1", "o1-2024-12-17",
	"o1-pro", "o1-pro-2025-03-19",
	"o1-preview", "o1-preview-2024-09-12",
	"o1-mini", "o1-mini-2024-09-12",
	"o3", "o3-2025-04-16",
	"o3-mini", "o3-mini-2025-01-31",
	"o3-pro", "o3-pro-2025-06-10",
	"o4-mini", "o4-mini-2025-04-16",
	// "computer-use-preview", "computer-use-preview-2025-03-11", // TODO: not implemented
	// -------------------------------------
	// Embeddings
	// -------------------------------------
	"text-embedding-ada-002", "text-embedding-3-small", "text-embedding-3-large",
	"text-curie-001", "text-babbage-001", "text-ada-001", "text-davinci-002", "text-davinci-003",
	"text-moderation-latest", "text-moderation-stable",
	"text-davinci-edit-001",
	"davinci-002", "babbage-002",
	// https://platform.openai.com/docs/guides/tools-web-search?api-mode=chat
	"gpt-4o-search-preview", "gpt-4o-search-preview-2025-03-11",
	"gpt-4o-mini-search-preview", "gpt-4o-mini-search-preview-2025-03-11",
	// -------------------------------------
	// Draw
	// -------------------------------------
	"dall-e-2", "dall-e-3", "gpt-image-1",
	// -------------------------------------
	// Audio
	// -------------------------------------
	"whisper-1",
	"tts-1", "tts-1-1106", "tts-1-hd", "tts-1-hd-1106",
	"gpt-4o-transcribe",
	"gpt-4o-mini-transcribe",
	"gpt-4o-mini-tts",
}
