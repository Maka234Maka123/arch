package middleware

import (
	"context"

	"archv2/pkg/logger"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	hertztracing "github.com/hertz-contrib/obs-opentelemetry/tracing"
)

const (
	// TraceIDHeaderKey 响应头中的 trace ID 字段名
	TraceIDHeaderKey = "X-Trace-ID"
)

// TracerConfig 类型别名，方便外部使用
type TracerConfig = hertztracing.Config

// NewServerTracer 创建服务端 tracer
// 返回的 tracer 需要在创建 Hertz 服务器时传入
// 用法: tracer, cfg := middleware.NewServerTracer()
//
//	h := server.New(tracer, ...)
//	h.Use(middleware.ServerMiddleware(cfg))
func NewServerTracer() (config.Option, *TracerConfig) {
	tracer, cfg := hertztracing.NewServerTracer()
	return tracer, cfg
}

// ServerMiddleware 返回 OpenTelemetry 链路追踪中间件
// 使用 hertz-contrib/obs-opentelemetry 官方实现
// 功能包括：
// 1. 自动创建 span 并记录 HTTP 请求信息
// 2. 支持 W3C Trace Context 传播
// 3. 自动记录请求/响应状态
// 4. 支持错误状态标记
// 5. 设置 trace ID 到响应头，便于调试和用户反馈
func ServerMiddleware(cfg *hertztracing.Config) app.HandlerFunc {
	tracing := hertztracing.ServerMiddleware(cfg)
	return func(ctx context.Context, c *app.RequestContext) {
		tracing(ctx, c)
		// 设置 trace ID 到响应头
		if traceID := logger.TraceIDFromContext(ctx); traceID != "" {
			c.Response.Header.Set(TraceIDHeaderKey, traceID)
		}
	}
}
