package response

import "fmt"

// 常用错误构造函数

// BadRequest 请求参数错误
func BadRequest(message string) *Result {
	return Err(CodeBadRequest, message)
}

// Forbidden 无权限
func Forbidden(message string) *Result {
	return Err(CodeForbidden, message)
}

// NotFound 资源不存在
func NotFound(message string) *Result {
	return Err(CodeNotFound, message)
}

// Conflict 资源冲突
func Conflict(message string) *Result {
	return Err(CodeConflict, message)
}

// Validation 参数校验失败
func Validation(message string) *Result {
	return Err(CodeValidation, message)
}

// TooManyRequests 请求过于频繁
func TooManyRequests(message string) *Result {
	return Err(CodeTooManyRequests, message)
}

// Internal 服务器内部错误
func Internal(message string) *Result {
	return Err(CodeInternal, message)
}

// Internalf 服务器内部错误（格式化消息）
func Internalf(format string, args ...any) *Result {
	return Internal(fmt.Sprintf(format, args...))
}

// ServiceUnavailable 服务不可用
func ServiceUnavailable(message string) *Result {
	return Err(CodeUnavailable, message)
}
