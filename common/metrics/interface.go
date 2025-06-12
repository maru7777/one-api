package metrics

import (
	"time"
)

// MetricsRecorder defines the interface for recording metrics
type MetricsRecorder interface {
	// HTTP metrics
	RecordHTTPRequest(startTime time.Time, path, method, statusCode string)
	RecordHTTPActiveRequest(path, method string, delta float64)

	// Relay metrics
	RecordRelayRequest(startTime time.Time, channelId int, channelType, model, userId string, success bool, promptTokens, completionTokens int, quotaUsed float64)

	// Channel metrics
	UpdateChannelMetrics(channelId int, channelName, channelType string, status int, balance float64, responseTimeMs int, successRate float64)
	UpdateChannelRequestsInFlight(channelId int, channelName, channelType string, delta float64)

	// User metrics
	RecordUserMetrics(userId, username, group string, quotaUsed float64, promptTokens, completionTokens int, balance float64)

	// Database metrics
	RecordDBQuery(startTime time.Time, operation, table string, success bool)
	UpdateDBConnectionMetrics(inUse, idle int)

	// Redis metrics
	RecordRedisCommand(startTime time.Time, command string, success bool)
	UpdateRedisConnectionMetrics(active int)

	// Rate limit metrics
	RecordRateLimitHit(limitType, identifier string)
	UpdateRateLimitRemaining(limitType, identifier string, remaining int)

	// Authentication metrics
	RecordTokenAuth(success bool)
	UpdateActiveTokens(userId, tokenName string, count int)

	// Error metrics
	RecordError(errorType, component string)

	// Model metrics
	RecordModelUsage(modelName, channelType string, latency time.Duration)

	// System metrics
	InitSystemMetrics(version, buildTime, goVersion string, startTime time.Time)
}

// Global metrics recorder instance
var GlobalRecorder MetricsRecorder

// NoOpRecorder is a no-operation implementation for when metrics are disabled
type NoOpRecorder struct{}

func (n *NoOpRecorder) RecordHTTPRequest(startTime time.Time, path, method, statusCode string) {}
func (n *NoOpRecorder) RecordHTTPActiveRequest(path, method string, delta float64)             {}
func (n *NoOpRecorder) RecordRelayRequest(startTime time.Time, channelId int, channelType, model, userId string, success bool, promptTokens, completionTokens int, quotaUsed float64) {
}
func (n *NoOpRecorder) UpdateChannelMetrics(channelId int, channelName, channelType string, status int, balance float64, responseTimeMs int, successRate float64) {
}
func (n *NoOpRecorder) UpdateChannelRequestsInFlight(channelId int, channelName, channelType string, delta float64) {
}
func (n *NoOpRecorder) RecordUserMetrics(userId, username, group string, quotaUsed float64, promptTokens, completionTokens int, balance float64) {
}
func (n *NoOpRecorder) RecordDBQuery(startTime time.Time, operation, table string, success bool)    {}
func (n *NoOpRecorder) UpdateDBConnectionMetrics(inUse, idle int)                                   {}
func (n *NoOpRecorder) RecordRedisCommand(startTime time.Time, command string, success bool)        {}
func (n *NoOpRecorder) UpdateRedisConnectionMetrics(active int)                                     {}
func (n *NoOpRecorder) RecordRateLimitHit(limitType, identifier string)                             {}
func (n *NoOpRecorder) UpdateRateLimitRemaining(limitType, identifier string, remaining int)        {}
func (n *NoOpRecorder) RecordTokenAuth(success bool)                                                {}
func (n *NoOpRecorder) UpdateActiveTokens(userId, tokenName string, count int)                      {}
func (n *NoOpRecorder) RecordError(errorType, component string)                                     {}
func (n *NoOpRecorder) RecordModelUsage(modelName, channelType string, latency time.Duration)       {}
func (n *NoOpRecorder) InitSystemMetrics(version, buildTime, goVersion string, startTime time.Time) {}

// Initialize with no-op recorder by default
func init() {
	GlobalRecorder = &NoOpRecorder{}
}
