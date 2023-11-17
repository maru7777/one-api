package controller

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneralOpenAIRequest_TextMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages interface{}
		want     []Message
		wantErr  error
	}{
		{
			name:     "Test with []any messages",
			messages: []any{Message{}, Message{}},
			want:     []Message{{}, {}},
			wantErr:  nil,
		},
		{
			name:     "Test with []Message messages",
			messages: []Message{{}, {}},
			want:     []Message{{}, {}},
			wantErr:  nil,
		},
		{
			name:     "Test with invalid message type",
			messages: "invalid",
			want:     nil,
			wantErr:  fmt.Errorf("invalid message type string"),
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

			gotMessages, err := got.TextMessages()
			if tt.wantErr != nil {
				require.ErrorContains(t, err, "cannot unmarshal string into Go value")
				return
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, gotMessages)
		})
	}
}
