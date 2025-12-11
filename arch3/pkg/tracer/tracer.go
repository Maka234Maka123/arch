// Package tracer 提供 OpenTelemetry trace 的便捷封装
//
// 使用方式:
//
//	ctx, span := tracer.Start(ctx, "operation.name")
//	defer span.End()
//
//	// 设置属性
//	span.SetAttributes(attribute.String("key", "value"))
//
//	// 记录错误
//	if err != nil {
//	    tracer.RecordError(span, err)
//	}
package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TracerName 是默认的 tracer 名称，用于在 APM 中标识来源
	TracerName = "arch3"
)

// tracer 全局 tracer 实例
var tracer = otel.Tracer(TracerName)

// Start 创建一个新的 span 并返回包含该 span 的 context
//
// 使用示例:
//
//	func DoSomething(ctx context.Context) error {
//	    ctx, span := tracer.Start(ctx, "DoSomething")
//	    defer span.End()
//
//	    // 业务逻辑...
//	    return nil
//	}
//
// 命名规范:
//   - Handler 层: "handler.{HandlerName}" 如 "handler.SendSMS"
//   - Service 层: "service.{ServiceName}.{Method}" 如 "service.user.SendSMS"
//   - 集成层: "{provider}.{operation}" 如 "sms.volcengine.Send"
//   - 数据库: "db.{operation}" 如 "db.query", "db.insert"
//   - 缓存: "cache.{operation}" 如 "cache.get", "cache.set"
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...)
}

// RecordError 记录错误到 span 并设置错误状态
//
// 使用示例:
//
//	if err != nil {
//	    tracer.RecordError(span, err)
//	    return err
//	}
func RecordError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetStatus 设置 span 状态
func SetStatus(span trace.Span, code codes.Code, description string) {
	span.SetStatus(code, description)
}

// AddEvent 添加事件到 span
//
// 事件用于记录 span 生命周期内的重要时刻，如:
//   - 开始处理请求
//   - 完成某个关键步骤
//   - 发生重试
//
// 使用示例:
//
//	tracer.AddEvent(span, "validation.passed")
//	tracer.AddEvent(span, "retry.attempt", attribute.Int("attempt", 2))
func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SpanFromContext 从 context 中获取当前 span
// 如果 context 中没有 span，返回一个 no-op span
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// Attribute 快捷方法，用于创建常用属性
var (
	// String 创建字符串属性
	String = attribute.String
	// Int 创建整数属性
	Int = attribute.Int
	// Int64 创建 int64 属性
	Int64 = attribute.Int64
	// Bool 创建布尔属性
	Bool = attribute.Bool
	// Float64 创建浮点数属性
	Float64 = attribute.Float64
	// StringSlice 创建字符串切片属性
	StringSlice = attribute.StringSlice
)

// 常用属性键
const (
	// HTTP 相关
	AttrHTTPMethod     = "http.method"
	AttrHTTPStatusCode = "http.status_code"
	AttrHTTPURL        = "http.url"

	// 数据库相关
	AttrDBSystem    = "db.system"
	AttrDBStatement = "db.statement"
	AttrDBOperation = "db.operation"

	// 业务相关
	AttrUserID      = "user.id"
	AttrPhoneNumber = "phone.number"
	AttrPhoneMasked = "phone.masked"
	AttrSMSType     = "sms.type"
	AttrSMSProvider = "sms.provider"

	// 错误相关
	AttrErrorType    = "error.type"
	AttrErrorMessage = "error.message"
)

// MaskPhone 手机号脱敏，保留前3位和后4位
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
