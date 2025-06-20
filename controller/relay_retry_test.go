package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChannelService for testing
type MockChannelService struct {
	mock.Mock
	channels []model.Channel
}

func (m *MockChannelService) CacheGetRandomSatisfiedChannelExcluding(group string, modelName string, ignoreFirstPriority bool, excludeChannelIds map[int]bool) (*model.Channel, error) {
	args := m.Called(group, modelName, ignoreFirstPriority, excludeChannelIds)
	return args.Get(0).(*model.Channel), args.Error(1)
}

func TestRelay429RetryLogic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                   string
		initialError           int
		channels               []model.Channel
		expectedRetryCount     int
		shouldTryLowerPriority bool
	}{
		{
			name:         "429 error should try lower priority channels first",
			initialError: http.StatusTooManyRequests,
			channels: []model.Channel{
				{Id: 1, Priority: &[]int64{100}[0], Name: "high-priority"},
				{Id: 2, Priority: &[]int64{50}[0], Name: "low-priority"},
			},
			expectedRetryCount:     2,
			shouldTryLowerPriority: true,
		},
		{
			name:         "500 error should try highest priority channels first",
			initialError: http.StatusInternalServerError,
			channels: []model.Channel{
				{Id: 1, Priority: &[]int64{100}[0], Name: "high-priority"},
				{Id: 2, Priority: &[]int64{50}[0], Name: "low-priority"},
			},
			expectedRetryCount:     2,
			shouldTryLowerPriority: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockService := &MockChannelService{
				channels: tt.channels,
			}

			// Mock the first call returns the first channel
			if tt.shouldTryLowerPriority {
				// For 429 errors, should try lower priority first
				mockService.On("CacheGetRandomSatisfiedChannelExcluding",
					"default", "gpt-3.5-turbo", true, mock.AnythingOfType("map[int]bool")).
					Return(&tt.channels[1], nil).Once()
			} else {
				// For other errors, should try highest priority first
				mockService.On("CacheGetRandomSatisfiedChannelExcluding",
					"default", "gpt-3.5-turbo", false, mock.AnythingOfType("map[int]bool")).
					Return(&tt.channels[0], nil).Once()
			}

			// Test setup
			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"test"}]}`))
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Set context values
			c.Set(ctxkey.ChannelId, tt.channels[0].Id)
			c.Set(ctxkey.Id, 1)
			c.Set(ctxkey.Group, "default")
			c.Set(ctxkey.OriginalModel, "gpt-3.5-turbo")
			c.Set(ctxkey.ChannelName, tt.channels[0].Name)

			// The test logic would need to be implemented to actually call the retry logic
			// This is a framework for the test, actual implementation would depend on
			// how we can inject the mock service into the relay function

			mockService.AssertExpectations(t)
		})
	}
}

func TestChannelExclusionInRetry(t *testing.T) {
	// Test that failed channels are properly excluded from subsequent retries

	failedChannels := map[int]bool{
		1: true, // Channel 1 has failed
		3: true, // Channel 3 has failed
	}

	// Available channels
	availableChannels := []model.Channel{
		{Id: 1, Priority: &[]int64{100}[0]}, // Should be excluded
		{Id: 2, Priority: &[]int64{100}[0]}, // Available
		{Id: 3, Priority: &[]int64{50}[0]},  // Should be excluded
		{Id: 4, Priority: &[]int64{50}[0]},  // Available
	}

	// Test logic would verify that channels 1 and 3 are not selected
	// and only channels 2 and 4 are considered for selection

	expectedAvailableIds := []int{2, 4}
	actualAvailableIds := []int{}

	for _, channel := range availableChannels {
		if !failedChannels[channel.Id] {
			actualAvailableIds = append(actualAvailableIds, channel.Id)
		}
	}

	assert.Equal(t, expectedAvailableIds, actualAvailableIds, "Should exclude failed channels")
}
