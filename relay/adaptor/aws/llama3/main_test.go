package aws_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	aws "github.com/songquanpeng/one-api/relay/adaptor/aws/llama3"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestRenderPrompt(t *testing.T) {
	messages := []relaymodel.Message{
		{
			Role:    "user",
			Content: "What's your name?",
		},
	}
	prompt := aws.RenderPrompt(messages)
	expected := `<|begin_of_text|><|start_header_id|>user<|end_header_id|>What's your name?<|eot_id|><|start_header_id|>assistant<|end_header_id|>
`
	require.Equal(t, expected, prompt)

	messages = []relaymodel.Message{
		{
			Role:    "system",
			Content: "Your name is Kat. You are a detective.",
		},
		{
			Role:    "user",
			Content: "What's your name?",
		},
		{
			Role:    "assistant",
			Content: "Kat",
		},
		{
			Role:    "user",
			Content: "What's your job?",
		},
	}
	prompt = aws.RenderPrompt(messages)
	expected = `<|begin_of_text|><|start_header_id|>system<|end_header_id|>Your name is Kat. You are a detective.<|eot_id|><|start_header_id|>user<|end_header_id|>What's your name?<|eot_id|><|start_header_id|>assistant<|end_header_id|>Kat<|eot_id|><|start_header_id|>user<|end_header_id|>What's your job?<|eot_id|><|start_header_id|>assistant<|end_header_id|>
`
	require.Equal(t, expected, prompt)
}
