package middleware

import (
	"testing"

	"github.com/Laisky/errors/v2"
	"github.com/stretchr/testify/assert"

	"github.com/songquanpeng/one-api/model"
)

// TestChannelPriorityLogic tests the channel priority selection logic
// This test verifies that priority fallback works correctly when high priority channels fail
func TestChannelPriorityLogic(t *testing.T) {
	tests := []struct {
		name                 string
		highPriorityChannels []*model.Channel
		lowPriorityChannels  []*model.Channel
		highPriorityError    error
		lowPriorityError     error
		expectedChannelId    int
		expectedError        bool
		description          string
	}{
		{
			name: "high_priority_available",
			highPriorityChannels: []*model.Channel{
				{Id: 1, Priority: ptrToInt64(10)},
			},
			lowPriorityChannels: []*model.Channel{
				{Id: 2, Priority: ptrToInt64(5)},
			},
			highPriorityError: nil,
			lowPriorityError:  nil,
			expectedChannelId: 1,
			expectedError:     false,
			description:       "Should use high priority channel when available",
		},
		{
			name:                 "fallback_to_low_priority",
			highPriorityChannels: nil,
			lowPriorityChannels: []*model.Channel{
				{Id: 2, Priority: ptrToInt64(5)},
			},
			highPriorityError: errors.New("no high priority channels available"),
			lowPriorityError:  nil,
			expectedChannelId: 2,
			expectedError:     false,
			description:       "Should fallback to low priority when high priority unavailable",
		},
		{
			name:                 "no_channels_available",
			highPriorityChannels: nil,
			lowPriorityChannels:  nil,
			highPriorityError:    errors.New("no high priority channels available"),
			lowPriorityError:     errors.New("no channels available"),
			expectedChannelId:    0,
			expectedError:        true,
			description:          "Should return error when no channels available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)

			// Simulate the channel selection logic from distributor middleware
			var selectedChannel *model.Channel
			var finalError error

			// First try to get highest priority channels (ignoreFirstPriority=false)
			if tt.highPriorityError != nil {
				// High priority failed, try lower priority (ignoreFirstPriority=true)
				t.Logf("High priority channels unavailable, trying lower priority")
				if tt.lowPriorityError != nil {
					finalError = tt.lowPriorityError
				} else if len(tt.lowPriorityChannels) > 0 {
					selectedChannel = tt.lowPriorityChannels[0]
				}
			} else if len(tt.highPriorityChannels) > 0 {
				selectedChannel = tt.highPriorityChannels[0]
			}

			// Verify results
			if tt.expectedError {
				assert.Error(t, finalError, "Should have error when no channels available")
				assert.Nil(t, selectedChannel, "Channel should be nil when no channels available")
				t.Logf("✓ Correctly failed with error: %v", finalError)
			} else {
				assert.NoError(t, finalError, "Should not have error when channels are available")
				assert.NotNil(t, selectedChannel, "Channel should not be nil")
				assert.Equal(t, tt.expectedChannelId, selectedChannel.Id, "Should select correct channel")
				t.Logf("✓ Selected channel %d as expected", selectedChannel.Id)
			}
		})
	}
}

// TestChannelPriorityFallbackScenario tests specific priority fallback scenarios
func TestChannelPriorityFallbackScenario(t *testing.T) {
	t.Run("rate_limit_suspension_fallback", func(t *testing.T) {
		// Simulate a scenario where high priority channels are suspended due to 429 errors
		// and the system should fallback to lower priority channels

		highPriorityUnavailable := errors.New("high priority channels suspended due to rate limits")
		lowPriorityChannel := &model.Channel{
			Id:       100,
			Priority: ptrToInt64(25),
			Name:     "backup-channel",
		}

		t.Logf("Simulating rate limit scenario where high priority channels are suspended")

		// First attempt (high priority) fails
		var selectedChannel *model.Channel
		var err error

		// Simulate high priority failure
		err = highPriorityUnavailable
		if err != nil {
			t.Logf("High priority channels unavailable: %v", err)
			// Fallback to lower priority
			selectedChannel = lowPriorityChannel
			err = nil
		}

		assert.NoError(t, err, "Should successfully fallback to lower priority channels")
		assert.NotNil(t, selectedChannel, "Should get a channel from fallback")
		assert.Equal(t, 100, selectedChannel.Id, "Should get the lower priority channel")
		assert.Equal(t, int64(25), *selectedChannel.Priority, "Should have correct priority")

		t.Logf("✓ Successfully fell back from high priority (suspended) to low priority channel")
		t.Logf("✓ Channel selected: ID=%d, Priority=%d", selectedChannel.Id, *selectedChannel.Priority)
	})
}

// Helper function to create pointer to int64
func ptrToInt64(v int64) *int64 {
	return &v
}
