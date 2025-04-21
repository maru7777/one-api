package zhipu

import (
	"time"

	"github.com/songquanpeng/one-api/relay/model"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Prompt      []Message `json:"prompt"`
	Temperature *float64  `json:"temperature,omitempty"`
	TopP        *float64  `json:"top_p,omitempty"`
	RequestId   string    `json:"request_id,omitempty"`
	Incremental bool      `json:"incremental,omitempty"`
}

type ResponseData struct {
	TaskId      string    `json:"task_id"`
	RequestId   string    `json:"request_id"`
	TaskStatus  string    `json:"task_status"`
	Choices     []Message `json:"choices"`
	model.Usage `json:"usage"`
}

type Response struct {
	Code    int          `json:"code"`
	Msg     string       `json:"msg"`
	Success bool         `json:"success"`
	Data    ResponseData `json:"data"`
}

type StreamMetaResponse struct {
	RequestId   string `json:"request_id"`
	TaskId      string `json:"task_id"`
	TaskStatus  string `json:"task_status"`
	model.Usage `json:"usage"`
}

type tokenData struct {
	Token      string
	ExpiryTime time.Time
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbeddingResponse struct {
	Model       string          `json:"model"`
	Object      string          `json:"object"`
	Embeddings  []EmbeddingData `json:"data"`
	model.Usage `json:"usage"`
}

type EmbeddingData struct {
	Index     int       `json:"index"`
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
}

type ImageRequest struct {
	Model   string `json:"model,omitempty"`
	Prompt  string `json:"prompt,omitempty"`
	Quality string `json:"quality,omitempty" validate:"oneof=hd standard low"`
	Size    string `json:"size,omitempty" validate:"oneof=1024x1024 768x1344 864x1152 1344x768 1152x864 1440x720 720x1440"`
	UserId  string `json:"user_id,omitempty"`
}
