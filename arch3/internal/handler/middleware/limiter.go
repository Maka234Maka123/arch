package middleware

import (
	"arch3/internal/config"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/limiter"
)

// Limiter 返回限流中间件
//
// 使用自适应限流模式，根据 CPU 使用率动态调整限流阈值，适合云原生环境。
// 当 CPU 超过配置的阈值时开始拒绝请求，返回 503。
//
// 配置说明:
//   - CPUThreshold: CPU 使用率阈值，范围 0-1000，800 表示 80%
func Limiter(cfg *config.LimiterConfig) app.HandlerFunc {
	return limiter.AdaptiveLimit(
		limiter.WithCPUThreshold(cfg.CPUThreshold),
	)
}
