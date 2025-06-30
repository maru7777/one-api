package monitor

import (
	"runtime"
	"time"

	"github.com/songquanpeng/one-api/common/metrics"
	"github.com/songquanpeng/one-api/monitor/prometheus"
)

// InitPrometheusMonitoring initializes all Prometheus monitoring components
func InitPrometheusMonitoring(version, buildTime, goVersion string, startTime time.Time) error {
	// Set up the Prometheus recorder as the global metrics recorder
	metrics.GlobalRecorder = &prometheus.PrometheusRecorder{}

	// Initialize system metrics
	metrics.GlobalRecorder.InitSystemMetrics(version, buildTime, goVersion, startTime)

	// Start background metric collection
	go collectSystemMetrics()
	go collectChannelMetrics()
	go collectUserMetrics()

	return nil
}

// collectSystemMetrics collects system-wide metrics periodically
func collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Update memory and runtime metrics (can be extended)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		// You can add custom metrics for memory usage if needed
	}
}

// collectChannelMetrics collects channel-related metrics periodically
func collectChannelMetrics() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Channel metrics are now updated through the metrics interface
		// when actual requests are made
	}
}

// collectUserMetrics collects user-related metrics periodically
func collectUserMetrics() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// User metrics are recorded per-request through the metrics interface
	}
}
