package middleware

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/limiter"
)

// Limiter 返回自适应限流中间件
//
// 使用 hertz-contrib/limiter 的自适应限流算法:
//   - 根据 CPU 使用率动态调整限流阈值
//   - 当 CPU 使用率超过阈值时自动限流
//   - 比固定速率限流更适合云原生环境
//
// 配置说明:
//
//	CPUThreshold 800 = 80% CPU 使用率
//	当 CPU 超过 80% 时开始拒绝请求，返回 503
//
// 注意: config.LimiterConfig 中的 rate/burst 配置用于传统限流模式，
// 当前使用自适应模式不需要这些配置。如需切换到固定速率限流，
// 请使用 limiter.TokenBucket 等其他限流器。
func Limiter() app.HandlerFunc {
	return limiter.AdaptiveLimit(
		limiter.WithCPUThreshold(800), // CPU 使用率阈值 80%
	)
}
