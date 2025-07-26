package groq

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
)

func TestGetRequestURL(t *testing.T) {
	adaptor := &Adaptor{}

	testCases := []struct {
		name           string
		requestURLPath string
		expectedURL    string
		baseURL        string
		channelType    int
	}{
		{
			name:           "Claude Messages API conversion",
			requestURLPath: "/v1/messages",
			expectedURL:    "https://api.groq.com/v1/chat/completions",
			baseURL:        "https://api.groq.com",
			channelType:    channeltype.Groq,
		},
		{
			name:           "OpenAI Chat Completions passthrough",
			requestURLPath: "/v1/chat/completions",
			expectedURL:    "https://api.groq.com/v1/chat/completions",
			baseURL:        "https://api.groq.com",
			channelType:    channeltype.Groq,
		},
		{
			name:           "Other endpoints passthrough",
			requestURLPath: "/v1/models",
			expectedURL:    "https://api.groq.com/v1/models",
			baseURL:        "https://api.groq.com",
			channelType:    channeltype.Groq,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			meta := &meta.Meta{
				RequestURLPath: tc.requestURLPath,
				BaseURL:        tc.baseURL,
				ChannelType:    tc.channelType,
			}

			url, err := adaptor.GetRequestURL(meta)
			if err != nil {
				t.Fatalf("GetRequestURL failed: %v", err)
			}

			if url != tc.expectedURL {
				t.Errorf("Expected URL %s, got %s", tc.expectedURL, url)
			}
		})
	}
}
