package response

import (
	"errors"
	"fmt"
)

// Err 创建错误响应
func Err(code int, message string) *Result {
	return &Result{
		Code:    code,
		Message: message,
	}
}

// Errf 创建错误响应（格式化消息）
func Errf(code int, format string, args ...any) *Result {
	return Err(code, fmt.Sprintf(format, args...))
}

// CodeFromError 从 error 中提取业务码，默认返回 CodeInternal
func CodeFromError(err error) int {
	var r *Result
	if errors.As(err, &r) {
		return r.Code
	}
	return CodeInternal
}
