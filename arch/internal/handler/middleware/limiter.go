package middleware

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/limiter"
)

// Limiter 返回自适应限流中间件
func Limiter() app.HandlerFunc {
	return limiter.AdaptiveLimit(
		limiter.WithCPUThreshold(800), // CPU 使用率阈值 80%
	)
}
