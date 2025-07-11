package common

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/metrics"
)

// PrometheusRedisHook implements redis.Hook for monitoring Redis operations
type PrometheusRedisHook struct{}

func (p *PrometheusRedisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, "redis_start_time", time.Now())
	return ctx, nil
}

func (p *PrometheusRedisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	p.recordMetrics(ctx, cmd)
	return nil
}

func (p *PrometheusRedisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, "redis_start_time", time.Now())
	return ctx, nil
}

func (p *PrometheusRedisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	// For pipeline, we'll record metrics for each command
	for _, cmd := range cmds {
		p.recordMetrics(ctx, cmd)
	}
	return nil
}

func (p *PrometheusRedisHook) recordMetrics(ctx context.Context, cmd redis.Cmder) {
	startTimeInterface := ctx.Value("redis_start_time")
	if startTimeInterface == nil {
		return
	}

	startTime, ok := startTimeInterface.(time.Time)
	if !ok {
		return
	}

	// Get command name
	cmdName := strings.ToUpper(cmd.Name())

	// Check if command was successful
	success := cmd.Err() == nil

	// Record metrics
	metrics.GlobalRecorder.RecordRedisCommand(startTime, cmdName, success)
}

// InitPrometheusRedisMonitoring adds Prometheus monitoring to Redis client
func InitPrometheusRedisMonitoring() {
	if RDB != nil {
		// Try to cast to concrete Redis client type to add hooks
		if client, ok := RDB.(*redis.Client); ok {
			hook := &PrometheusRedisHook{}
			client.AddHook(hook)

			// Update connection metrics periodically
			go func() {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()

				for range ticker.C {
					if client != nil {
						stats := client.PoolStats()
						metrics.GlobalRecorder.UpdateRedisConnectionMetrics(int(stats.TotalConns))
					}
				}
			}()
		} else {
			// For other Redis client types, we can't easily add hooks
			// But we can still provide basic monitoring by wrapping commands
			logger.SysLog("Redis monitoring: Using basic monitoring (hooks not available for this client type)")
		}
	}
}
