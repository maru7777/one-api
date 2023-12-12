package controller

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneralOpenAIRequest_TextMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		want     []Message
		wantErr  error
	}{
		{
			name:     "Test with []any messages",
			messages: []Message{Message{}, Message{}},
			want:     []Message{{}, {}},
			wantErr:  nil,
		},
		{
			name:     "Test with []Message messages",
			messages: []Message{{}, {}},
			want:     []Message{{}, {}},
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &GeneralOpenAIRequest{
				Messages: tt.messages,
			}
			got := new(GeneralOpenAIRequest)

			blob, err := json.Marshal(r)
			require.NoError(t, err)
			err = json.Unmarshal(blob, got)
			require.NoError(t, err)

			require.Equal(t, tt.want, got.Messages)
		})
	}
}
