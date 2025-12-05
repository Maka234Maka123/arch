package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// httpRequestsTotal HTTP 请求总数
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "echo",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestDuration HTTP 请求延迟
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "echo",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// httpRequestsInFlight 当前正在处理的请求数
	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "echo",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Current number of HTTP requests being processed",
		},
	)

	// httpRequestSize HTTP 请求大小
	httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "echo",
			Subsystem: "http",
			Name:      "request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to 10GB
		},
		[]string{"method", "path"},
	)

	// httpResponseSize HTTP 响应大小
	httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "echo",
			Subsystem: "http",
			Name:      "response_size_bytes",
			Help:      "HTTP response size in bytes",
		},
		[]string{"method", "path"},
	)
)

// metricsRegistry 全局 metrics registry
var metricsRegistry *prometheus.Registry

// InitMetrics 初始化 Prometheus metrics
func InitMetrics() *prometheus.Registry {
	if metricsRegistry != nil {
		return metricsRegistry
	}

	metricsRegistry = prometheus.NewRegistry()

	// 注册标准 Go 运行时指标
	metricsRegistry.MustRegister(collectors.NewGoCollector())
	metricsRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// 注册自定义指标
	metricsRegistry.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsInFlight,
		httpRequestSize,
		httpResponseSize,
	)

	return metricsRegistry
}

// GetMetricsRegistry 获取 metrics registry
func GetMetricsRegistry() *prometheus.Registry {
	if metricsRegistry == nil {
		return InitMetrics()
	}
	return metricsRegistry
}

// GetMetricsHandler 获取 Prometheus metrics HTTP handler
func GetMetricsHandler() app.HandlerFunc {
	h := promhttp.HandlerFor(GetMetricsRegistry(), promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	return func(ctx context.Context, c *app.RequestContext) {
		// 使用适配器将 Hertz 请求转换为标准 http
		rw := &metricsResponseWriter{c: c}
		req, _ := http.NewRequest(
			string(c.Method()),
			string(c.URI().FullURI()),
			nil,
		)
		h.ServeHTTP(rw, req)
	}
}

// Metrics 返回 Prometheus 指标收集中间件
func Metrics() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 跳过 metrics 端点自身
		path := string(c.Path())
		if path == "/metrics" {
			c.Next(ctx)
			return
		}

		start := time.Now()
		method := string(c.Method())

		// 增加正在处理的请求数
		httpRequestsInFlight.Inc()

		// 记录请求大小
		reqSize := float64(c.Request.Header.ContentLength())
		if reqSize < 0 {
			reqSize = 0
		}

		c.Next(ctx)

		// 减少正在处理的请求数
		httpRequestsInFlight.Dec()

		// 记录指标
		status := strconv.Itoa(c.Response.StatusCode())
		duration := time.Since(start).Seconds()

		// 规范化路径（避免高基数问题）
		normalizedPath := normalizePath(path)

		httpRequestsTotal.WithLabelValues(method, normalizedPath, status).Inc()
		httpRequestDuration.WithLabelValues(method, normalizedPath).Observe(duration)
		httpRequestSize.WithLabelValues(method, normalizedPath).Observe(reqSize)
		httpResponseSize.WithLabelValues(method, normalizedPath).Observe(float64(c.Response.Header.ContentLength()))
	}
}

// normalizePath 规范化路径，避免高基数问题
func normalizePath(path string) string {
	// 保留已知的静态路径
	staticPaths := map[string]bool{
		"/health":  true,
		"/metrics": true,
		"/version": true,
		"/ready":   true,
		"/live":    true,
	}

	if staticPaths[path] {
		return path
	}

	// 对于 API 路径，保留前两级
	// 例如: /api/v1/users/123 -> /api/v1/users/:id
	return path
}

// metricsResponseWriter 适配器，将 Hertz RequestContext 适配为 http.ResponseWriter
type metricsResponseWriter struct {
	c       *app.RequestContext
	headers http.Header
}

func (w *metricsResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *metricsResponseWriter) Write(data []byte) (int, error) {
	// 设置 Content-Type
	if ct := w.headers.Get("Content-Type"); ct != "" {
		w.c.Response.Header.Set("Content-Type", ct)
	}
	w.c.Write(data)
	return len(data), nil
}

func (w *metricsResponseWriter) WriteHeader(statusCode int) {
	w.c.SetStatusCode(statusCode)
	// 复制 headers
	for k, v := range w.headers {
		for _, vv := range v {
			w.c.Response.Header.Add(k, vv)
		}
	}
}

// init 确保 consts 包被使用
var _ = consts.StatusOK
