package middleware

import (
	"echo/internal/config"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// Register 注册所有中间件到 Hertz 服务器
// 中间件执行顺序：Recovery → Metrics → Tracing → RequestID → AccessLog → CORS → Gzip → Limiter
func Register(h *server.Hertz, cfg *config.Config) {
	// 0. 初始化 Prometheus Metrics
	InitMetrics()

	// 1. Recovery 中间件（最先注册，捕获所有 panic）
	h.Use(Recovery())

	// 2. Metrics 指标收集中间件
	h.Use(Metrics())

	// 3. Tracing 链路追踪中间件
	if cfg.Middleware.Tracing.Enabled {
		h.Use(Tracing())
	}

	// 4. RequestID 中间件（为每个请求生成唯一 ID）
	h.Use(RequestID())

	// 5. AccessLog 访问日志中间件
	if cfg.Middleware.AccessLog.Enabled {
		h.Use(AccessLog(&cfg.Middleware.AccessLog))
	}

	// 6. CORS 跨域中间件
	if cfg.Middleware.CORS.Enabled {
		h.Use(CORS(&cfg.Middleware.CORS))
	}

	// 7. Gzip 压缩中间件
	if cfg.Middleware.Gzip.Enabled {
		h.Use(Gzip(&cfg.Middleware.Gzip))
	}

	// 8. Limiter 限流中间件
	if cfg.Middleware.Limiter.Enabled {
		h.Use(Limiter())
	}

	// 9. Pprof 性能分析（仅在 debug 模式下启用）
	if cfg.Server.IsDebug() {
		RegisterPprof(h)
	}
}
