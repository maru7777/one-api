package veo

// CreateVideoRequest is the request body for the Veo API.
//
// https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/veo-video-generation
type CreateVideoRequest struct {
	Instances  []CreateVideoInstance `json:"instances" binding:"required,min=1"`
	Parameters CreateVideoParameters `json:"parameters" binding:"required"`
}

type CreateVideoInstance struct {
	Prompt string                    `json:"prompt" binding:"required"`
	Image  *CreateVideoInstanceImage `json:"image,omitempty"` // Optional image to be used as a prompt
}

type CreateVideoInstanceImage struct {
	BytesBase64Encoded string  `json:"bytesBase64Encoded" binding:"required"`
	MimeType           *string `json:"mimeType,omitempty" binding:"omitempty,oneof=image/jpeg image/png"`
}

type CreateVideoParameters struct {
	SampleCount      int     `json:"sampleCount" binding:"required,min=1"`
	AspectRatio      *string `json:"aspectRatio,omitempty"`
	NegativePrompt   *string `json:"negativePrompt,omitempty"`
	PersonGeneration *string `json:"personGeneration,omitempty"`
	Seed             *uint32 `json:"seed,omitempty"`
	StorageUri       *string `json:"storageUri,omitempty"`
	// DurationSeconds specifies the duration of the video in seconds, default is 8 seconds.
	DurationSeconds *int  `json:"durationSeconds,omitempty"`
	EnhancePrompt   *bool `json:"enhancePrompt,omitempty"`
	// GenerateAudio specifies whether to generate audio for the video.
	// not support for veo-2.0-generate-001.
	GenerateAudio *bool `json:"generateAudio,omitempty"`
}

type CreateVideoTaskResponse struct {
	// Name is the unique identifier for the video generation task.
	Name string `json:"name"`
}

type PollVideoTaskRequest struct {
	OperationName string `json:"operationName" binding:"required"`
}

type PollVideoTaskResponse struct {
	Name     string                    `json:"name"`
	Done     bool                      `json:"done"`
	Response PollVideoTaskResponseData `json:"response"`
}

type PollVideoTaskResponseData struct {
	Type             string            `json:"@type"`
	GeneratedSamples []GeneratedSample `json:"generatedSamples"`
}

type GeneratedSample struct {
	Video VideoData `json:"video"`
}

type VideoData struct {
	URI      string `json:"uri"`
	Encoding string `json:"encoding"`
}
