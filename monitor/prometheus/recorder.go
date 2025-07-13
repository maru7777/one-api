package prometheus

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/songquanpeng/one-api/common/metrics"
)

// PrometheusRecorder implements the MetricsRecorder interface using Prometheus
type PrometheusRecorder struct{}

// Prometheus metrics definitions
var (
	// HTTP request metrics
	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "one_api_http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"path", "method", "status_code"})

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"path", "method", "status_code"})

	httpActiveRequests = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_http_active_requests",
		Help: "Number of active HTTP requests",
	}, []string{"path", "method"})

	// API relay metrics
	relayRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "one_api_relay_request_duration_seconds",
		Help:    "Duration of API relay requests in seconds",
		Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120},
	}, []string{"channel_id", "channel_type", "model", "user_id", "success"})

	relayRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_relay_requests_total",
		Help: "Total number of API relay requests",
	}, []string{"channel_id", "channel_type", "model", "user_id", "success"})

	relayTokensUsed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_relay_tokens_total",
		Help: "Total number of tokens used in relay requests",
	}, []string{"channel_id", "channel_type", "model", "user_id", "token_type"})

	relayQuotaUsed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_relay_quota_used_total",
		Help: "Total quota used in relay requests",
	}, []string{"channel_id", "channel_type", "model", "user_id"})

	// Channel metrics
	channelStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_channel_status",
		Help: "Channel status (1=enabled, 0=disabled, -1=auto_disabled)",
	}, []string{"channel_id", "channel_name", "channel_type"})

	channelBalance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_channel_balance_usd",
		Help: "Channel balance in USD",
	}, []string{"channel_id", "channel_name", "channel_type"})

	channelResponseTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_channel_response_time_ms",
		Help: "Channel response time in milliseconds",
	}, []string{"channel_id", "channel_name", "channel_type"})

	channelSuccessRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_channel_success_rate",
		Help: "Channel success rate (0-1)",
	}, []string{"channel_id", "channel_name", "channel_type"})

	channelRequestsInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_channel_requests_in_flight",
		Help: "Number of requests currently being processed by channel",
	}, []string{"channel_id", "channel_name", "channel_type"})

	// User metrics
	userRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_user_requests_total",
		Help: "Total number of requests by user",
	}, []string{"user_id", "username", "group"})

	userQuotaUsed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_user_quota_used_total",
		Help: "Total quota used by user",
	}, []string{"user_id", "username", "group"})

	userTokensUsed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_user_tokens_total",
		Help: "Total tokens used by user",
	}, []string{"user_id", "username", "group", "token_type"})

	userBalance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_user_balance",
		Help: "User balance/quota remaining",
	}, []string{"user_id", "username", "group"})

	// Database metrics
	dbConnectionsInUse = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "one_api_db_connections_in_use",
		Help: "Number of database connections currently in use",
	})

	dbConnectionsIdle = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "one_api_db_connections_idle",
		Help: "Number of idle database connections",
	})

	dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "one_api_db_query_duration_seconds",
		Help:    "Duration of database queries in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	}, []string{"operation", "table"})

	dbQueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_db_queries_total",
		Help: "Total number of database queries",
	}, []string{"operation", "table", "success"})

	// Redis metrics
	redisConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "one_api_redis_connections_active",
		Help: "Number of active Redis connections",
	})

	redisCommandDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "one_api_redis_command_duration_seconds",
		Help:    "Duration of Redis commands in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	}, []string{"command"})

	redisCommandsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_redis_commands_total",
		Help: "Total number of Redis commands",
	}, []string{"command", "success"})

	// System metrics
	systemInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_system_info",
		Help: "System information",
	}, []string{"version", "build_time", "go_version"})

	systemStartTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "one_api_system_start_time_seconds",
		Help: "Unix timestamp when the system started",
	})

	// Rate limiting metrics
	rateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_rate_limit_hits_total",
		Help: "Total number of rate limit hits",
	}, []string{"type", "identifier"})

	rateLimitRemaining = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_rate_limit_remaining",
		Help: "Remaining rate limit tokens",
	}, []string{"type", "identifier"})

	// Token authentication metrics
	tokenAuthAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_token_auth_attempts_total",
		Help: "Total number of token authentication attempts",
	}, []string{"success"})

	activeTokens = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "one_api_active_tokens",
		Help: "Number of active API tokens",
	}, []string{"user_id", "token_name"})

	// Error metrics
	errorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_errors_total",
		Help: "Total number of errors by type",
	}, []string{"error_type", "component"})

	// Model usage metrics
	modelUsage = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "one_api_model_usage_total",
		Help: "Total usage count per model",
	}, []string{"model_name", "channel_type"})

	modelLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "one_api_model_latency_seconds",
		Help:    "Model response latency in seconds",
		Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120},
	}, []string{"model_name", "channel_type"})
)

// RecordHTTPRequest records HTTP request metrics
func (p *PrometheusRecorder) RecordHTTPRequest(startTime time.Time, path, method, statusCode string) {
	duration := time.Since(startTime).Seconds()
	httpRequestDuration.WithLabelValues(path, method, statusCode).Observe(duration)
	httpRequestsTotal.WithLabelValues(path, method, statusCode).Inc()
}

// RecordHTTPActiveRequest tracks active HTTP requests
func (p *PrometheusRecorder) RecordHTTPActiveRequest(path, method string, delta float64) {
	httpActiveRequests.WithLabelValues(path, method).Add(delta)
}

// RecordRelayRequest records API relay request metrics
func (p *PrometheusRecorder) RecordRelayRequest(startTime time.Time, channelId int, channelType, model, userId string, success bool, promptTokens, completionTokens int, quotaUsed float64) {
	duration := time.Since(startTime).Seconds()
	channelIdStr := strconv.Itoa(channelId)
	successStr := strconv.FormatBool(success)

	relayRequestDuration.WithLabelValues(channelIdStr, channelType, model, userId, successStr).Observe(duration)
	relayRequestsTotal.WithLabelValues(channelIdStr, channelType, model, userId, successStr).Inc()

	if promptTokens > 0 {
		relayTokensUsed.WithLabelValues(channelIdStr, channelType, model, userId, "prompt").Add(float64(promptTokens))
	}
	if completionTokens > 0 {
		relayTokensUsed.WithLabelValues(channelIdStr, channelType, model, userId, "completion").Add(float64(completionTokens))
	}
	if quotaUsed > 0 {
		relayQuotaUsed.WithLabelValues(channelIdStr, channelType, model, userId).Add(quotaUsed)
	}
}

// UpdateChannelMetrics updates channel-related metrics
func (p *PrometheusRecorder) UpdateChannelMetrics(channelId int, channelName, channelType string, status int, balance float64, responseTimeMs int, successRate float64) {
	channelIdStr := strconv.Itoa(channelId)
	var statusValue float64
	switch status {
	case 1: // enabled
		statusValue = 1
	case 2: // auto disabled
		statusValue = -1
	default: // disabled
		statusValue = 0
	}

	channelStatus.WithLabelValues(channelIdStr, channelName, channelType).Set(statusValue)
	channelBalance.WithLabelValues(channelIdStr, channelName, channelType).Set(balance)
	channelResponseTime.WithLabelValues(channelIdStr, channelName, channelType).Set(float64(responseTimeMs))
	channelSuccessRate.WithLabelValues(channelIdStr, channelName, channelType).Set(successRate)
}

// UpdateChannelRequestsInFlight updates the number of requests currently being processed
func (p *PrometheusRecorder) UpdateChannelRequestsInFlight(channelId int, channelName, channelType string, delta float64) {
	channelIdStr := strconv.Itoa(channelId)
	channelRequestsInFlight.WithLabelValues(channelIdStr, channelName, channelType).Add(delta)
}

// RecordUserMetrics records user-related metrics
func (p *PrometheusRecorder) RecordUserMetrics(userId, username, group string, quotaUsed float64, promptTokens, completionTokens int, balance float64) {
	userRequestsTotal.WithLabelValues(userId, username, group).Inc()
	if quotaUsed > 0 {
		userQuotaUsed.WithLabelValues(userId, username, group).Add(quotaUsed)
	}
	if promptTokens > 0 {
		userTokensUsed.WithLabelValues(userId, username, group, "prompt").Add(float64(promptTokens))
	}
	if completionTokens > 0 {
		userTokensUsed.WithLabelValues(userId, username, group, "completion").Add(float64(completionTokens))
	}
	userBalance.WithLabelValues(userId, username, group).Set(balance)
}

// RecordDBQuery records database-related metrics
func (p *PrometheusRecorder) RecordDBQuery(startTime time.Time, operation, table string, success bool) {
	duration := time.Since(startTime).Seconds()
	successStr := strconv.FormatBool(success)

	dbQueryDuration.WithLabelValues(operation, table).Observe(duration)
	dbQueriesTotal.WithLabelValues(operation, table, successStr).Inc()
}

// UpdateDBConnectionMetrics updates database connection metrics
func (p *PrometheusRecorder) UpdateDBConnectionMetrics(inUse, idle int) {
	dbConnectionsInUse.Set(float64(inUse))
	dbConnectionsIdle.Set(float64(idle))
}

// RecordRedisCommand records Redis command metrics
func (p *PrometheusRecorder) RecordRedisCommand(startTime time.Time, command string, success bool) {
	duration := time.Since(startTime).Seconds()
	successStr := strconv.FormatBool(success)

	redisCommandDuration.WithLabelValues(command).Observe(duration)
	redisCommandsTotal.WithLabelValues(command, successStr).Inc()
}

// UpdateRedisConnectionMetrics updates Redis connection metrics
func (p *PrometheusRecorder) UpdateRedisConnectionMetrics(active int) {
	redisConnectionsActive.Set(float64(active))
}

// RecordRateLimitHit records rate limiting metrics
func (p *PrometheusRecorder) RecordRateLimitHit(limitType, identifier string) {
	rateLimitHits.WithLabelValues(limitType, identifier).Inc()
}

// UpdateRateLimitRemaining updates remaining rate limit tokens
func (p *PrometheusRecorder) UpdateRateLimitRemaining(limitType, identifier string, remaining int) {
	rateLimitRemaining.WithLabelValues(limitType, identifier).Set(float64(remaining))
}

// RecordTokenAuth records token authentication attempts
func (p *PrometheusRecorder) RecordTokenAuth(success bool) {
	successStr := strconv.FormatBool(success)
	tokenAuthAttempts.WithLabelValues(successStr).Inc()
}

// UpdateActiveTokens updates the count of active tokens
func (p *PrometheusRecorder) UpdateActiveTokens(userId, tokenName string, count int) {
	activeTokens.WithLabelValues(userId, tokenName).Set(float64(count))
}

// RecordError records errors by type and component
func (p *PrometheusRecorder) RecordError(errorType, component string) {
	errorsTotal.WithLabelValues(errorType, component).Inc()
}

// RecordModelUsage records model usage and latency
func (p *PrometheusRecorder) RecordModelUsage(modelName, channelType string, latency time.Duration) {
	modelUsage.WithLabelValues(modelName, channelType).Inc()
	modelLatency.WithLabelValues(modelName, channelType).Observe(latency.Seconds())
}

// InitSystemMetrics initializes system-wide metrics
func (p *PrometheusRecorder) InitSystemMetrics(version, buildTime, goVersion string, startTime time.Time) {
	systemInfo.WithLabelValues(version, buildTime, goVersion).Set(1)
	systemStartTime.Set(float64(startTime.Unix()))
}

// InitPrometheusRecorder initializes the Prometheus recorder and sets it as the global recorder
func InitPrometheusRecorder() {
	metrics.GlobalRecorder = &PrometheusRecorder{}
}
