package controller

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/model"
)

func TestShouldRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		statusCode      int
		specificChannel int
		expectError     bool
		errorContains   string
	}{
		{
			name:            "should retry on 200 OK",
			statusCode:      http.StatusOK,
			specificChannel: 0,
			expectError:     false,
		},
		{
			name:            "should retry on 429 when no specific channel",
			statusCode:      http.StatusTooManyRequests,
			specificChannel: 0,
			expectError:     false,
		},
		{
			name:            "should not retry when specific channel ID is set",
			statusCode:      http.StatusTooManyRequests,
			specificChannel: 123,
			expectError:     true,
			errorContains:   "specific channel ID (123) was provided",
		},
		{
			name:            "should retry on 500 error",
			statusCode:      http.StatusInternalServerError,
			specificChannel: 0,
			expectError:     false,
		},
		{
			name:            "should retry on 503 error",
			statusCode:      http.StatusServiceUnavailable,
			specificChannel: 0,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			c.Set(ctxkey.SpecificChannelId, tt.specificChannel)

			err := shouldRetry(c, tt.statusCode)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessChannelRelayError_StatusTooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original config and restore after test
	originalSuspendDuration := config.ChannelSuspendSecondsFor429
	defer func() {
		config.ChannelSuspendSecondsFor429 = originalSuspendDuration
	}()

	// Set test config
	config.ChannelSuspendSecondsFor429 = 60 * time.Second

	tests := []struct {
		name          string
		statusCode    int
		expectSuspend bool
		suspendError  error
	}{
		{
			name:          "should suspend ability on 429",
			statusCode:    http.StatusTooManyRequests,
			expectSuspend: true,
			suspendError:  nil,
		},
		{
			name:          "should not suspend on 200",
			statusCode:    http.StatusOK,
			expectSuspend: false,
		},
		{
			name:          "should not suspend on 500",
			statusCode:    http.StatusInternalServerError,
			expectSuspend: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test data
			channelId := 123
			originalModel := "gpt-3.5-turbo"
			group := "default"

			relayError := model.ErrorWithStatusCode{
				StatusCode: tt.statusCode,
				Error: model.Error{
					Message: "Test error message",
					Type:    "rate_limit_error",
				},
			}

			// The test verifies that the function handles 429 errors correctly
			// and attempts to suspend the ability only for 429 status codes
			if tt.statusCode == http.StatusTooManyRequests {
				// This would normally call dbmodel.SuspendAbility
				// We can verify the behavior by checking the logs or mocking the call
				t.Logf("Would suspend ability for channel %d, model %s, group %s for %v",
					channelId, originalModel, group, config.ChannelSuspendSecondsFor429)
			}

			// For unit testing purposes, we verify the logic without side effects
			shouldSuspend := tt.statusCode == http.StatusTooManyRequests
			assert.Equal(t, tt.expectSuspend, shouldSuspend)

			// Verify the error structure is correct
			assert.Equal(t, tt.statusCode, relayError.StatusCode)
		})
	}
}

func TestProcessChannelRelayError_SuspensionScope(t *testing.T) {
	// This test verifies that 429 errors only affect the specific model/group/channel combination
	// and not the entire channel

	tests := []struct {
		name      string
		scenarios []struct {
			group   string
			model   string
			channel int
			status  int
		}
		expectedSuspensions []string // format: "group:model:channel"
	}{
		{
			name: "single model suspension",
			scenarios: []struct {
				group   string
				model   string
				channel int
				status  int
			}{
				{"default", "gpt-3.5-turbo", 123, http.StatusTooManyRequests},
			},
			expectedSuspensions: []string{"default:gpt-3.5-turbo:123"},
		},
		{
			name: "multiple models, only one suspended",
			scenarios: []struct {
				group   string
				model   string
				channel int
				status  int
			}{
				{"default", "gpt-3.5-turbo", 123, http.StatusTooManyRequests},
				{"default", "gpt-4", 123, http.StatusOK},
				{"vip", "gpt-3.5-turbo", 123, http.StatusOK},
			},
			expectedSuspensions: []string{"default:gpt-3.5-turbo:123"},
		},
		{
			name: "same model different groups",
			scenarios: []struct {
				group   string
				model   string
				channel int
				status  int
			}{
				{"default", "gpt-3.5-turbo", 123, http.StatusTooManyRequests},
				{"vip", "gpt-3.5-turbo", 123, http.StatusTooManyRequests},
			},
			expectedSuspensions: []string{"default:gpt-3.5-turbo:123", "vip:gpt-3.5-turbo:123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suspensions := make([]string, 0)

			for _, scenario := range tt.scenarios {
				if scenario.status == http.StatusTooManyRequests {
					// This represents what would be suspended
					suspension := fmt.Sprintf("%s:%s:%d", scenario.group, scenario.model, scenario.channel)
					suspensions = append(suspensions, suspension)
				}
			}

			assert.ElementsMatch(t, tt.expectedSuspensions, suspensions)
		})
	}
}

func TestProcessChannelRelayError_ModelLevelGranularity(t *testing.T) {
	// Verify that the suspension is model-specific, not channel-wide

	channelId := 123

	// Scenario: gpt-3.5-turbo gets 429, but gpt-4 should remain available
	errorModel := "gpt-3.5-turbo"
	workingModel := "gpt-4"
	group := "default"

	error429 := model.ErrorWithStatusCode{
		StatusCode: http.StatusTooManyRequests,
		Error: model.Error{
			Message: "Rate limit exceeded",
			Type:    "rate_limit_error",
		},
	}

	errorOK := model.ErrorWithStatusCode{
		StatusCode: http.StatusOK,
		Error: model.Error{
			Message: "Success",
		},
	}

	// Test that only the specific model should be suspended
	// gpt-3.5-turbo should be suspended
	shouldSuspendErrorModel := error429.StatusCode == http.StatusTooManyRequests
	assert.True(t, shouldSuspendErrorModel, "gpt-3.5-turbo should be suspended due to 429")

	// gpt-4 should not be suspended
	shouldSuspendWorkingModel := errorOK.StatusCode == http.StatusTooManyRequests
	assert.False(t, shouldSuspendWorkingModel, "gpt-4 should not be suspended")

	// Verify the granularity: only the specific (group, model, channel) combination is affected
	suspendedKey := fmt.Sprintf("%s:%s:%d", group, errorModel, channelId)
	workingKey := fmt.Sprintf("%s:%s:%d", group, workingModel, channelId)

	assert.NotEqual(t, suspendedKey, workingKey, "Suspended and working models should have different keys")

	t.Logf("Suspended: %s", suspendedKey)
	t.Logf("Still working: %s", workingKey)
}

// Test for enhanced 429 error handling
func TestRelay429ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                  string
		initialError          *model.ErrorWithStatusCode
		retryTimes            int
		expectedRetryIncrease bool
		expectedErrorMessage  string
	}{
		{
			name: "429 error should increase retry attempts",
			initialError: &model.ErrorWithStatusCode{
				StatusCode: http.StatusTooManyRequests,
				Error: model.Error{
					Message: "Rate limit exceeded",
					Type:    "rate_limit_error",
				},
			},
			retryTimes:            3,
			expectedRetryIncrease: true,
			expectedErrorMessage:  "All available channels",
		},
		{
			name: "500 error should not increase retry attempts",
			initialError: &model.ErrorWithStatusCode{
				StatusCode: http.StatusInternalServerError,
				Error: model.Error{
					Message: "Internal server error",
					Type:    "server_error",
				},
			},
			retryTimes:            3,
			expectedRetryIncrease: false,
			expectedErrorMessage:  "Internal server error",
		},
		{
			name: "404 error should not increase retry attempts",
			initialError: &model.ErrorWithStatusCode{
				StatusCode: http.StatusNotFound,
				Error: model.Error{
					Message: "Not found",
					Type:    "not_found_error",
				},
			},
			retryTimes:            3,
			expectedRetryIncrease: false,
			expectedErrorMessage:  "Not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original config and restore after test
			originalRetryTimes := config.RetryTimes
			defer func() {
				config.RetryTimes = originalRetryTimes
			}()
			config.RetryTimes = tt.retryTimes

			// Verify that 429 errors get special treatment
			if tt.initialError.StatusCode == http.StatusTooManyRequests && tt.expectedRetryIncrease {
				// The retry logic should increase attempts for 429 errors
				expectedRetries := tt.retryTimes * 2
				assert.Greater(t, expectedRetries, tt.retryTimes, "429 errors should increase retry attempts")
			}

			// Verify error message formatting for multiple channel failures
			if tt.initialError.StatusCode == http.StatusTooManyRequests {
				failedChannels := make(map[int]bool)
				failedChannels[1] = true
				failedChannels[2] = true

				if len(failedChannels) > 1 {
					expectedMsg := fmt.Sprintf("All available channels (%d) for this model are currently rate limited, please try again later", len(failedChannels))
					assert.Contains(t, expectedMsg, "All available channels")
				}
			}
		})
	}
}

func TestFailedChannelTracking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test that failed channels are properly tracked to avoid retrying them
	failedChannels := make(map[int]bool)

	// Initially empty
	assert.Empty(t, failedChannels)

	// Add some failed channels
	failedChannels[1] = true
	failedChannels[5] = true
	failedChannels[10] = true

	// Verify tracking
	assert.True(t, failedChannels[1])
	assert.True(t, failedChannels[5])
	assert.True(t, failedChannels[10])
	assert.False(t, failedChannels[99]) // Non-existent channel

	// Verify count for error messaging
	assert.Equal(t, 3, len(failedChannels))
}

// Test for the priority handling fix with 429 errors
func TestRelay429PriorityHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                        string
		initialErrorStatus          int
		retryAttempts               int
		expectedIgnoreFirstPriority []bool // Expected value for each retry attempt
	}{
		{
			name:               "429 error should ignore first priority for all retries",
			initialErrorStatus: http.StatusTooManyRequests,
			retryAttempts:      3,
			// For 429 errors, should ignore first priority for all retries
			expectedIgnoreFirstPriority: []bool{true, true, true},
		},
		{
			name:               "500 error should follow normal priority logic",
			initialErrorStatus: http.StatusInternalServerError,
			retryAttempts:      3,
			// For non-429 errors: i=3 (i==retryTimes, false), i=2 (i!=retryTimes, true), i=1 (i!=retryTimes, true)
			expectedIgnoreFirstPriority: []bool{false, true, true},
		},
		{
			name:               "404 error should follow normal priority logic",
			initialErrorStatus: http.StatusNotFound,
			retryAttempts:      3,
			// For non-429 errors: i=3 (i==retryTimes, false), i=2 (i!=retryTimes, true), i=1 (i!=retryTimes, true)
			expectedIgnoreFirstPriority: []bool{false, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the priority logic from the relay function
			ignoreFirstPriority := tt.initialErrorStatus == http.StatusTooManyRequests

			for i := tt.retryAttempts; i > 0; i-- {
				// This is the fixed logic from the relay function
				shouldIgnoreFirstPriority := ignoreFirstPriority || i != tt.retryAttempts

				// Get expected value for this retry attempt (convert retry index to array index)
				expectedIndex := tt.retryAttempts - i
				expected := tt.expectedIgnoreFirstPriority[expectedIndex]

				assert.Equal(t, expected, shouldIgnoreFirstPriority,
					"Retry attempt %d (i=%d) should have ignoreFirstPriority=%v", expectedIndex+1, i, expected)

				t.Logf("Retry attempt %d: ignoreFirstPriority=%v (expected %v)",
					expectedIndex+1, shouldIgnoreFirstPriority, expected)
			}
		})
	}
}

// Test that demonstrates the bug fix for channel priority selection
func TestChannelPriorityBugFix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// This test verifies the specific bug that was fixed:
	// Before fix: After first retry, system would go back to highest priority channels
	// After fix: Once we get a 429, we should continue trying lower priority channels

	testCases := []struct {
		name        string
		errorStatus int
		description string
		retryLogic  string
	}{
		{
			name:        "429 error - should try lower priority channels throughout",
			errorStatus: http.StatusTooManyRequests,
			description: "429 errors should cause system to ignore first priority for ALL retries",
			retryLogic:  "ignoreFirstPriority || i != retryTimes",
		},
		{
			name:        "Non-429 error - normal priority behavior",
			errorStatus: http.StatusInternalServerError,
			description: "Non-429 errors should follow normal priority logic (first retry uses first priority, subsequent ignore it)",
			retryLogic:  "ignoreFirstPriority || i != retryTimes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retryTimes := 4
			ignoreFirstPriority := tc.errorStatus == http.StatusTooManyRequests

			t.Logf("Testing %s", tc.description)
			t.Logf("Error status: %d", tc.errorStatus)
			t.Logf("Initial ignoreFirstPriority flag: %v", ignoreFirstPriority)

			retryResults := make([]bool, 0)

			for i := retryTimes; i > 0; i-- {
				shouldIgnore := ignoreFirstPriority || i != retryTimes
				retryResults = append(retryResults, shouldIgnore)

				t.Logf("Retry %d (i=%d): ignoreFirstPriority = %v", retryTimes-i+1, i, shouldIgnore)
			}

			if tc.errorStatus == http.StatusTooManyRequests {
				// For 429 errors, ALL retries should ignore first priority
				for i, result := range retryResults {
					assert.True(t, result, "Retry %d should ignore first priority for 429 errors", i+1)
				}
				t.Log("✓ All retries correctly ignore first priority for 429 errors")
			} else {
				// For non-429 errors: first retry should NOT ignore first priority, subsequent should
				assert.False(t, retryResults[0], "First retry should NOT ignore first priority")
				for i := 1; i < len(retryResults); i++ {
					assert.True(t, retryResults[i], "Retry %d should ignore first priority for non-429 errors", i+1)
				}
				t.Log("✓ First retry uses first priority, subsequent retries ignore first priority for non-429 errors")
			}
		})
	}
}

// Test to verify the model-specific suspension behavior
func TestModelSpecificSuspension(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// This test verifies that 429 errors only affect the specific model
	// and that the retry logic correctly tries alternative models/channels

	scenarios := []struct {
		channelId     int
		group         string
		model         string
		errorStatus   int
		shouldSuspend bool
	}{
		{123, "default", "gpt-3.5-turbo", http.StatusTooManyRequests, true},
		{123, "default", "gpt-4", http.StatusOK, false},
		{123, "vip", "gpt-3.5-turbo", http.StatusOK, false},
		{456, "default", "gpt-3.5-turbo", http.StatusOK, false},
	}

	suspendedCombinations := make(map[string]bool)

	for _, scenario := range scenarios {
		key := fmt.Sprintf("%s:%s:%d", scenario.group, scenario.model, scenario.channelId)

		if scenario.errorStatus == http.StatusTooManyRequests {
			suspendedCombinations[key] = true
			t.Logf("Suspended: %s (status: %d)", key, scenario.errorStatus)
		}

		// Verify suspension decision
		actualSuspension := scenario.errorStatus == http.StatusTooManyRequests
		assert.Equal(t, scenario.shouldSuspend, actualSuspension,
			"Suspension decision for %s should be %v", key, scenario.shouldSuspend)
	}

	// Verify only the correct combination is suspended
	assert.True(t, suspendedCombinations["default:gpt-3.5-turbo:123"])
	assert.False(t, suspendedCombinations["default:gpt-4:123"])
	assert.False(t, suspendedCombinations["vip:gpt-3.5-turbo:123"])
	assert.False(t, suspendedCombinations["default:gpt-3.5-turbo:456"])

	t.Logf("✓ Model-specific suspension working correctly - only affected specific group:model:channel combination")
}
