// Package response 提供统一的 HTTP 响应处理
// 所有接口统一返回 HTTP 200，通过业务码区分错误类型
// 前端根据 code 字段判断请求结果
package response

import (
	"context"
	"errors"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

// HandlerFunc 返回 error 的 Handler 函数签名
type HandlerFunc func(ctx context.Context, c *app.RequestContext) error

// Result 统一响应结构，同时实现 error 接口
type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error 实现 error 接口，使 Result 可以作为错误返回
func (r *Result) Error() string {
	return r.Message
}

// Wrap 包装 HandlerFunc 为 Hertz 标准 Handler
func Wrap(fn HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if err := fn(ctx, c); err != nil {
			_ = c.Error(err)
			Error(c, err)
		}
	}
}

// Success 发送成功响应
func Success(c *app.RequestContext, data any) error {
	c.JSON(http.StatusOK, Result{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
	return nil
}

// Error 发送错误响应
func Error(c *app.RequestContext, err error) {
	var r *Result
	if errors.As(err, &r) {
		c.JSON(http.StatusOK, r)
		return
	}
	// 非 Result 错误统一返回系统错误，避免泄露内部错误信息
	c.JSON(http.StatusOK, &Result{
		Code:    CodeInternal,
		Message: "系统错误",
	})
}

