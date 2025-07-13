package model

import (
	"encoding/json"
	"testing"
)

// TestToolIndexField tests that the Index field is properly serialized in streaming tool calls
func TestToolIndexField(t *testing.T) {
	// Test streaming tool call with Index field set
	index := 0
	streamingTool := Tool{
		Id:   "call_123",
		Type: "function",
		Function: Function{
			Name:      "get_weather",
			Arguments: `{"location": "Paris"}`,
		},
		Index: &index,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(streamingTool)
	if err != nil {
		t.Fatalf("Failed to marshal streaming tool: %v", err)
	}

	// Verify that the index field is present in JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that index field exists and has correct value
	if indexValue, exists := result["index"]; !exists {
		t.Error("Index field is missing from JSON output")
	} else if indexValue != float64(0) { // JSON numbers are float64
		t.Errorf("Expected index to be 0, got %v", indexValue)
	}

	// Test non-streaming tool call without Index field
	nonStreamingTool := Tool{
		Id:   "call_456",
		Type: "function",
		Function: Function{
			Name:      "send_email",
			Arguments: `{"to": "test@example.com"}`,
		},
		// Index is nil for non-streaming responses
	}

	// Serialize to JSON
	jsonData2, err := json.Marshal(nonStreamingTool)
	if err != nil {
		t.Fatalf("Failed to marshal non-streaming tool: %v", err)
	}

	// Verify that the index field is omitted in JSON (due to omitempty)
	var result2 map[string]interface{}
	err = json.Unmarshal(jsonData2, &result2)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that index field does not exist
	if _, exists := result2["index"]; exists {
		t.Error("Index field should be omitted for non-streaming tool calls")
	}
}

// TestStreamingToolCallAccumulation tests the complete streaming tool call accumulation workflow
func TestStreamingToolCallAccumulation(t *testing.T) {
	// Simulate streaming tool call deltas as they would come from the API
	streamingDeltas := []Tool{
		{
			Id:    "call_123",
			Type:  "function",
			Index: intPtr(0),
			Function: Function{
				Name:      "get_weather",
				Arguments: "",
			},
		},
		{
			Index: intPtr(0),
			Function: Function{
				Arguments: `{"location":`,
			},
		},
		{
			Index: intPtr(0),
			Function: Function{
				Arguments: ` "Paris"}`,
			},
		},
	}

	// Accumulate the deltas (simulating client-side accumulation)
	finalToolCalls := make(map[int]Tool)

	for _, delta := range streamingDeltas {
		if delta.Index == nil {
			t.Error("Index field should be present in streaming tool call deltas")
			continue
		}

		index := *delta.Index

		if _, exists := finalToolCalls[index]; !exists {
			// First delta for this tool call
			finalToolCalls[index] = delta
		} else {
			// Subsequent delta - accumulate arguments
			existing := finalToolCalls[index]
			existingArgs, _ := existing.Function.Arguments.(string)
			deltaArgs, _ := delta.Function.Arguments.(string)
			existing.Function.Arguments = existingArgs + deltaArgs
			finalToolCalls[index] = existing
		}
	}

	// Verify the final accumulated tool call
	if len(finalToolCalls) != 1 {
		t.Fatalf("Expected 1 final tool call, got %d", len(finalToolCalls))
	}

	finalTool := finalToolCalls[0]
	expectedArgs := `{"location": "Paris"}`
	actualArgs, _ := finalTool.Function.Arguments.(string)
	if actualArgs != expectedArgs {
		t.Errorf("Expected accumulated arguments '%s', got '%s'", expectedArgs, actualArgs)
	}

	if finalTool.Id != "call_123" {
		t.Errorf("Expected tool call id 'call_123', got '%s'", finalTool.Id)
	}

	if finalTool.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got '%s'", finalTool.Function.Name)
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// TestToolIndexFieldDeserialization tests that the Index field can be properly deserialized
func TestToolIndexFieldDeserialization(t *testing.T) {
	// JSON with index field (streaming response)
	streamingJSON := `{
		"id": "call_789",
		"type": "function",
		"function": {
			"name": "calculate",
			"arguments": "{\"x\": 5, \"y\": 3}"
		},
		"index": 1
	}`

	var streamingTool Tool
	err := json.Unmarshal([]byte(streamingJSON), &streamingTool)
	if err != nil {
		t.Fatalf("Failed to unmarshal streaming tool JSON: %v", err)
	}

	// Verify index field is properly set
	if streamingTool.Index == nil {
		t.Error("Index field should not be nil for streaming tool")
	} else if *streamingTool.Index != 1 {
		t.Errorf("Expected index to be 1, got %d", *streamingTool.Index)
	}

	// JSON without index field (non-streaming response)
	nonStreamingJSON := `{
		"id": "call_101",
		"type": "function",
		"function": {
			"name": "search",
			"arguments": "{\"query\": \"test\"}"
		}
	}`

	var nonStreamingTool Tool
	err = json.Unmarshal([]byte(nonStreamingJSON), &nonStreamingTool)
	if err != nil {
		t.Fatalf("Failed to unmarshal non-streaming tool JSON: %v", err)
	}

	// Verify index field is nil
	if nonStreamingTool.Index != nil {
		t.Error("Index field should be nil for non-streaming tool")
	}
}
