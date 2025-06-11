package openai

import (
	"strings"
	"testing"
)

// TestCompleteResponseAPIStream tests the complete Response API streaming workflow
// This test simulates the exact SSE format from the Response API specification
func TestCompleteResponseAPIStream(t *testing.T) {
	// This simulates exactly what would come from a Response API stream
	sseStreamExample := `event: response.created
data: {"type":"response.created","response":{"id":"resp_67c9fdcecf488190bdd9a0409de3a1ec07b8b0ad4e5eb654","object":"response","created_at":1741290958,"status":"in_progress","error":null,"incomplete_details":null,"instructions":"You are a helpful assistant.","max_output_tokens":null,"model":"gpt-4.1-2025-04-14","output":[],"parallel_tool_calls":true,"previous_response_id":null,"reasoning":{"effort":null,"summary":null},"store":true,"temperature":1.0,"text":{"format":{"type":"text"}},"tool_choice":"auto","tools":[],"top_p":1.0,"truncation":"disabled","usage":null,"user":null,"metadata":{}}}

event: response.output_item.added
data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","type":"message","status":"in_progress","role":"assistant","content":[]}}

event: response.content_part.added
data: {"type":"response.content_part.added","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"part":{"type":"output_text","text":"","annotations":[]}}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"delta":"Hi"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"delta":" there!"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"delta":" How can I assist you today?"}

event: response.output_text.done
data: {"type":"response.output_text.done","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"text":"Hi there! How can I assist you today?"}

event: response.content_part.done
data: {"type":"response.content_part.done","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"part":{"type":"output_text","text":"Hi there! How can I assist you today?","annotations":[]}}

event: response.output_item.done
data: {"type":"response.output_item.done","output_index":0,"item":{"id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Hi there! How can I assist you today?","annotations":[]}]}}

event: response.completed
data: {"type":"response.completed","response":{"id":"resp_67c9fdcecf488190bdd9a0409de3a1ec07b8b0ad4e5eb654","object":"response","created_at":1741290958,"status":"completed","error":null,"incomplete_details":null,"instructions":"You are a helpful assistant.","max_output_tokens":null,"model":"gpt-4.1-2025-04-14","output":[{"id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Hi there! How can I assist you today?","annotations":[]}]}],"parallel_tool_calls":true,"previous_response_id":null,"reasoning":{"effort":null,"summary":null},"store":true,"temperature":1.0,"text":{"format":{"type":"text"}},"tool_choice":"auto","tools":[],"top_p":1.0,"truncation":"disabled","usage":{"input_tokens":37,"output_tokens":11,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":48},"user":null,"metadata":{}}}

data: [DONE]`

	// Split into lines and process exactly like the ResponseAPIStreamHandler would
	lines := strings.Split(sseStreamExample, "\n")

	const dataPrefix = "data: "
	const dataPrefixLength = len(dataPrefix)

	responseText := ""
	eventCount := 0
	deltaCount := 0

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and event lines (exactly like NormalizeDataLine + ResponseAPIStreamHandler)
		if line == "" {
			continue
		}

		data := NormalizeDataLine(line)

		if !strings.HasPrefix(data, dataPrefix) {
			continue
		}

		// Extract JSON data
		jsonData := data[dataPrefixLength:]

		if jsonData == "[DONE]" {
			break
		}

		eventCount++

		// Parse using the improved parsing logic
		fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(jsonData))
		if err != nil {
			t.Errorf("Line %d: Parse error: %v", i+1, err)
			continue
		}

		// Convert to ResponseAPIResponse (same as ResponseAPIStreamHandler)
		var responseAPIChunk ResponseAPIResponse
		if fullResponse != nil {
			responseAPIChunk = *fullResponse
		} else if streamEvent != nil {
			responseAPIChunk = ConvertStreamEventToResponse(streamEvent)

			// Track delta events specifically
			if strings.Contains(streamEvent.Type, "delta") && streamEvent.Delta != "" {
				deltaCount++
				responseText += streamEvent.Delta
			}
		}

		// Convert to ChatCompletion format (same as ResponseAPIStreamHandler)
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)

		// Verify conversion worked
		if len(chatCompletionChunk.Choices) == 0 {
			t.Errorf("Line %d: ChatCompletion conversion failed - no choices", i+1)
		}
	}

	// Verify we got the expected content
	expectedText := "Hi there! How can I assist you today?"
	if responseText != expectedText {
		t.Errorf("Response text mismatch: expected '%s', got '%s'", expectedText, responseText)
	}

	// Verify we processed the expected number of events
	if eventCount == 0 {
		t.Error("No events were processed")
	}

	if deltaCount != 3 {
		t.Errorf("Expected 3 delta events, got %d", deltaCount)
	}
}

// TestResponseAPIStreamingEvents tests individual streaming events
func TestResponseAPIStreamingEvents(t *testing.T) {
	t.Run("Problematic streaming event", func(t *testing.T) {
		// Test the problematic streaming event that was causing the parsing error
		problematicEvent := `{"type":"response.output_text.done","sequence_number":22,"item_id":"msg_6849865110908191a4809c86e082ff710008bd3c6060334b","output_index":1,"content_index":0,"text":"Why don't skeletons fight each other?\n\nThey don't have the guts."}`

		// Test the new flexible parsing approach
		fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(problematicEvent))
		if err != nil {
			t.Fatalf("Failed to parse streaming event: %v", err)
		}

		if fullResponse != nil {
			t.Error("Expected stream event, got full response")
		}

		if streamEvent == nil {
			t.Fatal("Expected streamEvent to be non-nil")
		}

		if streamEvent.Type != "response.output_text.done" {
			t.Errorf("Expected type 'response.output_text.done', got '%s'", streamEvent.Type)
		}

		// Test conversion to ResponseAPIResponse
		responseAPIChunk := ConvertStreamEventToResponse(streamEvent)
		if len(responseAPIChunk.Output) == 0 {
			t.Error("Expected output items in converted response")
		}

		// Test conversion to ChatCompletion format
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)
		if len(chatCompletionChunk.Choices) == 0 {
			t.Error("Expected choices in ChatCompletion chunk")
		}
	})

	t.Run("Delta streaming event", func(t *testing.T) {
		deltaEvent := `{"type":"response.output_text.delta","sequence_number":6,"item_id":"msg_6849865110908191a4809c86e082ff710008bd3c6060334b","output_index":1,"content_index":0,"delta":"Why"}`

		_, streamEvent, err := ParseResponseAPIStreamEvent([]byte(deltaEvent))
		if err != nil {
			t.Fatalf("Failed to parse delta event: %v", err)
		}

		if streamEvent == nil {
			t.Fatal("Expected streamEvent to be non-nil")
		}

		if streamEvent.Type != "response.output_text.delta" {
			t.Errorf("Expected type 'response.output_text.delta', got '%s'", streamEvent.Type)
		}

		if streamEvent.Delta != "Why" {
			t.Errorf("Expected delta 'Why', got '%s'", streamEvent.Delta)
		}

		// Test conversion
		responseAPIChunk := ConvertStreamEventToResponse(streamEvent)
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)

		if len(chatCompletionChunk.Choices) == 0 {
			t.Error("Expected choices in ChatCompletion chunk")
		}

		if content, ok := chatCompletionChunk.Choices[0].Delta.Content.(string); !ok || content != "Why" {
			t.Errorf("Expected ChatCompletion delta content 'Why', got '%v'", chatCompletionChunk.Choices[0].Delta.Content)
		}
	})

	t.Run("Full response event", func(t *testing.T) {
		fullResponseEvent := `{"id":"resp_123","object":"response","created_at":1749648976,"status":"completed","model":"o3-2025-04-16","output":[{"type":"message","id":"msg_123","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Hello world"}]}],"usage":{"input_tokens":9,"output_tokens":22,"total_tokens":31}}`

		fullResponse, _, err := ParseResponseAPIStreamEvent([]byte(fullResponseEvent))
		if err != nil {
			t.Fatalf("Failed to parse full response event: %v", err)
		}

		if fullResponse == nil {
			t.Fatal("Expected fullResponse to be non-nil")
		}

		if fullResponse.Id != "resp_123" {
			t.Errorf("Expected ID 'resp_123', got '%s'", fullResponse.Id)
		}

		if fullResponse.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", fullResponse.Status)
		}

		if fullResponse.Usage == nil || fullResponse.Usage.TotalTokens != 31 {
			t.Errorf("Expected usage with total tokens 31, got %v", fullResponse.Usage)
		}
	})
}

// TestSSEProcessing tests SSE line processing logic
func TestSSEProcessing(t *testing.T) {
	testCases := []struct {
		name          string
		line          string
		shouldProcess bool
		expectError   bool
	}{
		{
			name:          "Valid data line",
			line:          `data: {"type":"response.output_text.delta","delta":"Hi"}`,
			shouldProcess: true,
			expectError:   false,
		},
		{
			name:          "Event line (should skip)",
			line:          "event: response.created",
			shouldProcess: false,
			expectError:   false,
		},
		{
			name:          "Empty line (should skip)",
			line:          "",
			shouldProcess: false,
			expectError:   false,
		},
		{
			name:          "DONE signal",
			line:          "data: [DONE]",
			shouldProcess: false, // DONE is handled specially
			expectError:   false,
		},
		{
			name:          "Malformed JSON",
			line:          `data: {"invalid": json}`,
			shouldProcess: true,
			expectError:   true,
		},
	}

	const dataPrefix = "data: "
	const dataPrefixLength = len(dataPrefix)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := NormalizeDataLine(tc.line)

			// Check if line should be processed
			if !strings.HasPrefix(data, dataPrefix) {
				if tc.shouldProcess {
					t.Errorf("Expected line to be processed, but it was skipped")
				}
				return
			}

			jsonData := data[dataPrefixLength:]

			if jsonData == "[DONE]" {
				if tc.shouldProcess {
					t.Errorf("DONE signal should not be processed as JSON")
				}
				return
			}

			if !tc.shouldProcess {
				t.Errorf("Expected line to be skipped, but it was processed")
				return
			}

			// Test parsing
			_, _, err := ParseResponseAPIStreamEvent([]byte(jsonData))

			if tc.expectError && err == nil {
				t.Errorf("Expected parsing error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected parsing error: %v", err)
			}
		})
	}
}

// TestStreamEventTypes tests various Response API stream event types
func TestStreamEventTypes(t *testing.T) {
	testCases := []struct {
		name         string
		eventData    string
		expectedType string
	}{
		{
			name:         "response.created event",
			eventData:    `{"type":"response.created","response":{"id":"resp_123","status":"in_progress"}}`,
			expectedType: "response.created",
		},
		{
			name:         "response.output_text.delta event",
			eventData:    `{"type":"response.output_text.delta","delta":"Hi"}`,
			expectedType: "response.output_text.delta",
		},
		{
			name:         "response.output_text.done event",
			eventData:    `{"type":"response.output_text.done","text":"Complete text"}`,
			expectedType: "response.output_text.done",
		},
		{
			name:         "response.completed event",
			eventData:    `{"type":"response.completed","response":{"id":"resp_123","status":"completed"}}`,
			expectedType: "response.completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(tc.eventData))
			if err != nil {
				t.Fatalf("Parsing failed: %v", err)
			}

			if fullResponse != nil && streamEvent != nil {
				t.Error("Both fullResponse and streamEvent should not be non-nil")
			}

			if fullResponse == nil && streamEvent == nil {
				t.Error("Both fullResponse and streamEvent should not be nil")
			}

			var eventType string
			if fullResponse != nil {
				// For full response events, we need to extract the type from the original data
				// This is a limitation of the current parsing approach
				if strings.Contains(tc.eventData, `"type":"response.created"`) {
					eventType = "response.created"
				} else if strings.Contains(tc.eventData, `"type":"response.completed"`) {
					eventType = "response.completed"
				}
			} else if streamEvent != nil {
				eventType = streamEvent.Type
			}

			if eventType != tc.expectedType {
				t.Errorf("Expected event type '%s', got '%s'", tc.expectedType, eventType)
			}

			// Test conversion for stream events
			if streamEvent != nil {
				responseAPIChunk := ConvertStreamEventToResponse(streamEvent)
				chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)

				if chatCompletionChunk.Object != "chat.completion.chunk" {
					t.Errorf("Expected object 'chat.completion.chunk', got '%s'", chatCompletionChunk.Object)
				}
			}
		})
	}
}

// TestNormalizeDataLine tests the data line normalization function
func TestNormalizeDataLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Already normalized",
			input:    "data: {\"test\": true}",
			expected: "data: {\"test\": true}",
		},
		{
			name:     "No space after colon",
			input:    "data:{\"test\": true}",
			expected: "data: {\"test\": true}",
		},
		{
			name:     "Multiple spaces after colon",
			input:    "data:   {\"test\": true}",
			expected: "data: {\"test\": true}",
		},
		{
			name:     "Tab after colon",
			input:    "data:\t{\"test\": true}",
			expected: "data: \t{\"test\": true}", // TrimLeft only removes spaces, not tabs
		},
		{
			name:     "Non-data line",
			input:    "event: test",
			expected: "event: test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeDataLine(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
