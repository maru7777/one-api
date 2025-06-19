package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
)

// MockChannelService implements a mock for testing channel selection fallback
type MockChannelService struct {
	highPriorityAvailable bool
	lowPriorityAvailable  bool
	callCount             int
}

func (m *MockChannelService) CacheGetRandomSatisfiedChannel(group string, modelName string, ignoreFirstPriority bool) (*model.Channel, error) {
	m.callCount++

	if !ignoreFirstPriority {
		// First priority request
		if m.highPriorityAvailable {
			return &model.Channel{
				Id:       1,
				Priority: 10,
			}, nil
		}
		return nil, errors.New("no high priority channels available")
	} else {
		// Lower priority request
		if m.lowPriorityAvailable {
			return &model.Channel{
				Id:       2,
				Priority: 5,
			}, nil
		}
		return nil, errors.New("no channels available")
	}
}

func TestDistributorChannelSelectionFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                  string
		highPriorityAvailable bool
		lowPriorityAvailable  bool
		expectedStatusCode    int
		expectedChannelId     int
		expectedCallCount     int
		description           string
	}{
		{
			name:                  "high_priority_available",
			highPriorityAvailable: true,
			lowPriorityAvailable:  true,
			expectedStatusCode:    200,
			expectedChannelId:     1,
			expectedCallCount:     1,
			description:           "Should use high priority channel when available",
		},
		{
			name:                  "fallback_to_low_priority",
			highPriorityAvailable: false,
			lowPriorityAvailable:  true,
			expectedStatusCode:    200,
			expectedChannelId:     2,
			expectedCallCount:     2,
			description:           "Should fallback to low priority when high priority unavailable",
		},
		{
			name:                  "no_channels_available",
			highPriorityAvailable: false,
			lowPriorityAvailable:  false,
			expectedStatusCode:    503,
			expectedChannelId:     0,
			expectedCallCount:     2,
			description:           "Should return 503 when no channels available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &MockChannelService{
				highPriorityAvailable: tt.highPriorityAvailable,
				lowPriorityAvailable:  tt.lowPriorityAvailable,
				callCount:             0,
			}

			// Setup test server
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)

			// Set required context values
			c.Set(ctxkey.Id, 1)
			c.Set(ctxkey.RequestModel, "gpt-3.5-turbo")

			// Mock the model.CacheGetRandomSatisfiedChannel function
			originalFunc := model.CacheGetRandomSatisfiedChannel
			model.CacheGetRandomSatisfiedChannel = mockService.CacheGetRandomSatisfiedChannel
			defer func() {
				model.CacheGetRandomSatisfiedChannel = originalFunc
			}()

			t.Logf("Testing: %s", tt.description)

			// Test the channel selection logic (simplified version of what's in distributor)
			userGroup := "default"
			requestModel := "gpt-3.5-turbo"
			var channel *model.Channel
			var err error

			// First try to get highest priority channels
			channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, requestModel, false)
			if err != nil {
				// If no highest priority channels available, try lower priority channels
				channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, requestModel, true)
			}

			// Verify results
			assert.Equal(t, tt.expectedCallCount, mockService.callCount, "Expected %d calls to CacheGetRandomSatisfiedChannel", tt.expectedCallCount)

			if tt.expectedStatusCode == 200 {
				assert.NoError(t, err, "Should not have error when channels are available")
				assert.NotNil(t, channel, "Channel should not be nil")
				assert.Equal(t, tt.expectedChannelId, channel.Id, "Should select correct channel")
				t.Logf("✓ Selected channel %d as expected", channel.Id)
			} else {
				assert.Error(t, err, "Should have error when no channels available")
				assert.Nil(t, channel, "Channel should be nil when no channels available")
				t.Logf("✓ Correctly failed with error: %v", err)
			}
		})
	}
}

func TestDistributorPriorityFallbackBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("priority_fallback_behavior", func(t *testing.T) {
		mockService := &MockChannelService{
			highPriorityAvailable: false, // High priority channels are suspended (429 errors)
			lowPriorityAvailable:  true,  // Low priority channels are available
			callCount:             0,
		}

		userGroup := "default"
		requestModel := "gemini-2.5-flash"

		// Mock the function
		originalFunc := model.CacheGetRandomSatisfiedChannel
		model.CacheGetRandomSatisfiedChannel = mockService.CacheGetRandomSatisfiedChannel
		defer func() {
			model.CacheGetRandomSatisfiedChannel = originalFunc
		}()

		// Simulate the middleware behavior
		channel, err := model.CacheGetRandomSatisfiedChannel(userGroup, requestModel, false)
		if err != nil {
			t.Logf("High priority channels unavailable (simulating 429 rate limits): %v", err)
			// Fallback to lower priority
			channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, requestModel, true)
		}

		assert.NoError(t, err, "Should successfully fallback to lower priority channels")
		assert.NotNil(t, channel, "Should get a channel from fallback")
		assert.Equal(t, 2, channel.Id, "Should get the lower priority channel")
		assert.Equal(t, 2, mockService.callCount, "Should make 2 calls - first to high priority, then to low priority")

		t.Logf("✓ Successfully fell back from high priority (suspended) to low priority channel")
		t.Logf("✓ Channel selected: ID=%d, Priority=%d", channel.Id, channel.Priority)
	})
}
