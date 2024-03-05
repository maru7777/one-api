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

type Response struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
	Error      Error  `json:"error"`
}
