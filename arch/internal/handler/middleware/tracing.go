package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// 确保 headerCarrier 实现 propagation.TextMapCarrier 接口
var _ propagation.TextMapCarrier = (*headerCarrier)(nil)

const tracerName = "echo-http-server"

// Tracing 返回 OpenTelemetry 链路追踪中间件
func Tracing() app.HandlerFunc {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(ctx context.Context, c *app.RequestContext) {
		// 从请求头中提取 trace context
		carrier := &headerCarrier{c: c}
		ctx = propagator.Extract(ctx, carrier)

		// 创建 span
		path := string(c.Path())
		method := string(c.Method())
		spanName := method + " " + path

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.url", string(c.URI().FullURI())),
				attribute.String("http.path", path),
				attribute.String("http.host", string(c.Host())),
				attribute.String("http.user_agent", string(c.UserAgent())),
				attribute.String("http.request_id", string(c.GetHeader("X-Request-ID"))),
				attribute.String("net.peer.ip", c.ClientIP()),
			),
		)
		defer span.End()

		// 将 span context 注入到响应头
		propagator.Inject(ctx, carrier)

		// 设置 trace ID 到响应头（便于调试）
		if span.SpanContext().HasTraceID() {
			c.Response.Header.Set("X-Trace-ID", span.SpanContext().TraceID().String())
		}

		// 继续处理请求
		c.Next(ctx)

		// 记录响应状态
		statusCode := c.Response.StatusCode()
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int("http.response_size", c.Response.Header.ContentLength()),
		)

		// 标记错误状态
		if statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}

// headerCarrier 实现 propagation.TextMapCarrier 接口
type headerCarrier struct {
	c *app.RequestContext
}

func (h *headerCarrier) Get(key string) string {
	return string(h.c.GetHeader(key))
}

func (h *headerCarrier) Set(key string, value string) {
	h.c.Response.Header.Set(key, value)
}

func (h *headerCarrier) Keys() []string {
	var keys []string
	h.c.Request.Header.VisitAll(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}
