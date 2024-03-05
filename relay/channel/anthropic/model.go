package anthropic

import (
	"github.com/songquanpeng/one-api/relay/model"
)

type Metadata struct {
	UserId string `json:"user_id"`
}

type Request struct {
	model.GeneralOpenAIRequest
	// System anthropic messages API use system to represent the system prompt
	System string `json:"system"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ResponseType string

const (
	TypeError        ResponseType = "error"
	TypeStart        ResponseType = "message_start"
	TypeContentStart ResponseType = "content_block_start"
	TypeContent      ResponseType = "content_block_delta"
	TypePing         ResponseType = "ping"
	TypeContentStop  ResponseType = "content_block_stop"
	TypeMessageDelta ResponseType = "message_delta"
	TypeMessageStop  ResponseType = "message_stop"
)

// https://docs.anthropic.com/claude/reference/messages-streaming
type Response struct {
	Type  ResponseType `json:"type"`
	Index int          `json:"index,omitempty"`
	Delta struct {
		Type       string `json:"type,omitempty"`
		Text       string `json:"text,omitempty"`
		StopReason string `json:"stop_reason,omitempty"`
	} `json:"delta,omitempty"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}
