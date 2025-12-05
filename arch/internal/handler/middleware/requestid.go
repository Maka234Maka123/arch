package middleware

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/requestid"
)

// RequestID 返回请求 ID 中间件
func RequestID() app.HandlerFunc {
	return requestid.New()
}

// GetRequestID 从请求上下文中获取请求 ID
func GetRequestID(c *app.RequestContext) string {
	return requestid.Get(c)
}
