package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// ============================================================================
// OTEL Metrics 指标定义（RED 方法）
// ============================================================================
//
// 所有指标使用 "archv2." 前缀
//
// 核心指标:
//   - archv2.http.request.duration (histogram) - Rate & Duration
//   - archv2.http.error.total (counter) - Errors
//   - archv2.http.panic.total (counter) - Panic

// MetricsManager 管理 OTEL Metrics 相关资源
//
// 职责:
//   - 初始化 OTEL Meter 和核心指标
//   - 提供指标记录的公共方法供 middleware 使用
//
// 核心指标（遵循 RED 方法）:
//   - Rate: 请求速率（通过 duration histogram 的 count 计算）
//   - Errors: 错误计数
//   - Duration: 请求延迟分布
type MetricsManager struct {
	meter metric.Meter

	// ========== 核心指标（RED 方法） ==========
	HttpRequestDuration metric.Float64Histogram // 包含 count，可计算 Rate
	HttpErrorTotal      metric.Int64Counter     // Errors
	HttpPanicTotal      metric.Int64Counter     // Panic 单独计数
}

var (
	metricsManager *MetricsManager
	metricsOnce    sync.Once
)

// InitMetrics 初始化 OTEL Metrics
//
// 注意: 必须在 TracingManager 初始化之后调用，
// 因为 OTEL Meter 依赖全局的 MeterProvider。
//
// 如果 Tracing 未启用，metrics 将使用 noop provider，
// 不会产生任何数据，但也不会报错。
func InitMetrics() error {
	var initErr error

	metricsOnce.Do(func() {
		mm := &MetricsManager{}
		initErr = mm.init()
		if initErr == nil {
			metricsManager = mm
		}
	})

	return initErr
}

func (mm *MetricsManager) init() error {
	var err error

	// 获取 Meter，名称用于标识指标来源
	mm.meter = otel.Meter("archv2",
		metric.WithInstrumentationVersion("1.0.0"),
	)

	// ========== 核心指标（RED 方法） ==========

	// Duration histogram 同时提供延迟分布和请求计数（Rate）
	mm.HttpRequestDuration, err = mm.meter.Float64Histogram(
		"archv2.http.request.duration",
		metric.WithDescription("HTTP request latency in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5),
	)
	if err != nil {
		return err
	}

	// 错误计数（4xx + 5xx）
	mm.HttpErrorTotal, err = mm.meter.Int64Counter(
		"archv2.http.error.total",
		metric.WithDescription("Total number of HTTP errors (4xx + 5xx)"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	// Panic 单独计数（严重性高于普通错误）
	mm.HttpPanicTotal, err = mm.meter.Int64Counter(
		"archv2.http.panic.total",
		metric.WithDescription("Total number of panic recoveries"),
		metric.WithUnit("{panic}"),
	)
	if err != nil {
		return err
	}

	return nil
}

// Metrics 返回 OTEL 指标收集中间件
//
// 基于 RED 方法收集核心指标:
//   - archv2.http.request.duration: 请求延迟分布（histogram 的 count 可计算 Rate）
//   - archv2.http.error.total: HTTP 错误计数（4xx + 5xx）
//
// 标签维度:
//   - method: HTTP 方法
//   - status: 响应状态码
func Metrics() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		mm := metricsManager
		if mm == nil {
			c.Next(ctx)
			return
		}

		// 跳过健康检查等运维端点
		path := string(c.Path())
		if isOpsPath(path) {
			c.Next(ctx)
			return
		}

		start := time.Now()

		c.Next(ctx)

		// 记录核心指标
		method := string(c.Method())
		statusCode := c.Response.StatusCode()

		// Duration（histogram count 自动提供 Rate）
		mm.HttpRequestDuration.Record(ctx, time.Since(start).Seconds(),
			metric.WithAttributes(
				attribute.String("method", method),
				attribute.Int("status", statusCode),
			),
		)

		// Errors
		if statusCode >= 400 {
			mm.HttpErrorTotal.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("method", method),
					attribute.Int("status", statusCode),
				),
			)
		}
	}
}

// RecordPanic 记录 panic 恢复（供 Recovery 中间件调用）
func RecordPanic(ctx context.Context, method string) {
	if metricsManager == nil {
		return
	}
	metricsManager.HttpPanicTotal.Add(ctx, 1,
		metric.WithAttributes(attribute.String("method", method)),
	)
}

// isOpsPath 判断是否为运维路径
func isOpsPath(path string) bool {
	switch path {
	case "/health", "/ready", "/live", "/version":
		return true
	default:
		return false
	}
}
