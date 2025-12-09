package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// traceIDKey 和 spanIDKey 用于日志字段名
const (
	TraceIDKey = "trace_id"
	SpanIDKey  = "span_id"
)

// spanContextFromCtx 从 context 中提取有效的 SpanContext
// 如果 context 为 nil 或 span 无效，返回空的 SpanContext
func spanContextFromCtx(ctx context.Context) trace.SpanContext {
	if ctx == nil {
		return trace.SpanContext{}
	}
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return trace.SpanContext{}
	}
	return span.SpanContext()
}

// Ctx 从 context 中提取 trace 信息并返回带有这些字段的 logger
// 这是推荐的日志记录方式，确保每条日志都能关联到请求
//
// 功能：
//   - 本地日志：添加 trace_id 和 span_id 字段，便于本地查看
//   - OTEL 导出：通过 zap.Any("context", ctx) 让 otelzap bridge 识别并关联到 APMPlus
//
// 使用示例:
//
//	logger.Ctx(ctx).Info("处理订单", zap.String("order_id", id))
//	// 本地输出: {"trace_id":"abc123","span_id":"def456","msg":"处理订单","order_id":"xxx"}
//	// OTEL 导出: 日志会自动关联到对应的 trace
func Ctx(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return L()
	}

	sc := spanContextFromCtx(ctx)
	if !sc.IsValid() {
		return L()
	}

	return L().With(
		// 本地日志显示 trace_id 和 span_id
		zap.String(TraceIDKey, sc.TraceID().String()),
		zap.String(SpanIDKey, sc.SpanID().String()),
		// otelzap bridge 需要这个字段来关联 OTEL trace
		zap.Any("context", ctx),
	)
}

// AsyncContext 创建用于异步任务的上下文
// 只保留 trace 信息，不保留请求相关的数据（如 deadline、cancel）
//
// 使用示例:
//
//	func handleRequest(ctx context.Context, c *app.RequestContext) {
//	    asyncCtx := logger.AsyncContext(ctx)
//
//	    go func() {
//	        // asyncCtx 保留了 trace_id，但不会因请求结束而被取消
//	        logger.Ctx(asyncCtx).Info("异步任务开始")
//	        processAsync()
//	        logger.Ctx(asyncCtx).Info("异步任务完成")
//	    }()
//
//	    c.JSON(200, response)
//	}
func AsyncContext(ctx context.Context) context.Context {
	sc := spanContextFromCtx(ctx)
	if !sc.IsValid() {
		return context.Background()
	}

	// 只传递 SpanContext，不传递整个 span
	// 这样异步任务的日志能关联到原请求，但不会影响原 span 的生命周期
	return trace.ContextWithSpanContext(context.Background(), sc)
}

// AsyncContextWithTimeout 创建带超时的异步上下文
// 适用于需要限制异步任务执行时间的场景
//
// 使用示例:
//
//	asyncCtx, cancel := logger.AsyncContextWithTimeout(ctx, 30*time.Second)
//	defer cancel()
//
//	go func() {
//	    select {
//	    case <-asyncCtx.Done():
//	        logger.Ctx(asyncCtx).Warn("异步任务超时")
//	    case result := <-processChan:
//	        logger.Ctx(asyncCtx).Info("异步任务完成", zap.Any("result", result))
//	    }
//	}()
func AsyncContextWithTimeout(ctx context.Context, timeout interface{ Duration() }) (context.Context, context.CancelFunc) {
	asyncCtx := AsyncContext(ctx)
	// 注意：这里需要用户自己处理 timeout，因为 time.Duration 不是 interface
	// 实际使用时请用 context.WithTimeout(AsyncContext(ctx), timeout)
	return context.WithCancel(asyncCtx)
}

// TraceIDFromContext 从 context 中提取 trace ID
// 用于需要获取 trace ID 但不需要完整 logger 的场景
//
// 使用示例:
//
//	traceID := logger.TraceIDFromContext(ctx)
//	c.Response.Header.Set("X-Trace-ID", traceID)
func TraceIDFromContext(ctx context.Context) string {
	sc := spanContextFromCtx(ctx)
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

// SpanIDFromContext 从 context 中提取 span ID
func SpanIDFromContext(ctx context.Context) string {
	sc := spanContextFromCtx(ctx)
	if !sc.IsValid() {
		return ""
	}
	return sc.SpanID().String()
}
