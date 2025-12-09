package middleware

import (
	"archv2/internal/config"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// Register 注册所有中间件到 Hertz 服务器
//
// 中间件执行顺序（洋葱模型，请求从外到内，响应从内到外）:
//
//	Request → Recovery → Metrics → Tracing → AccessLog → CORS → Gzip → Limiter → Handler
//	         ↑                                                                     ↓
//	         └─────────────────────── Response ────────────────────────────────────┘
//
// 顺序设计原则:
//  1. Recovery 最外层 - 捕获所有 panic，确保服务稳定
//  2. Metrics 紧随其后 - 记录所有请求（包括 panic 的）
//  3. Tracing 提供 trace_id - 后续中间件和 handler 都可使用
//  4. AccessLog 记录访问 - 需要 trace_id 关联日志
//  5. CORS/Gzip/Limiter 业务相关 - 按需启用
//
// 使用说明:
//   - logger.Ctx(ctx) 记录日志会自动包含 trace_id
//   - 响应头 X-Trace-ID 可用于问题排查
//
// 参数:
//   - tracerCfg: 如果启用了 tracing，需要传入 NewServerTracer() 返回的配置
//
// 注意: Metrics 使用 OTEL，依赖 TracingManager 初始化的 MeterProvider。
// 如果 Tracing 未启用，Metrics 将使用 noop provider（不记录数据但不报错）。
func Register(h *server.Hertz, cfg *config.Config, tracerCfg *TracerConfig) {
	// 1. Recovery - 捕获所有 panic，防止服务崩溃
	h.Use(Recovery())

	// 2. Metrics - 收集请求指标（总数、延迟、状态码分布等）
	h.Use(Metrics())

	// 3. Tracing - 链路追踪，提供 trace_id 作为请求唯一标识，并设置到响应头
	if cfg.Middleware.Tracing.Enabled && tracerCfg != nil {
		h.Use(ServerMiddleware(tracerCfg))
	}

	// 4. AccessLog - 访问日志，自动关联 trace_id
	if cfg.Middleware.AccessLog.Enabled {
		h.Use(AccessLog(&cfg.Middleware.AccessLog))
	}

	// 5. CORS - 跨域资源共享
	if cfg.Middleware.CORS.Enabled {
		h.Use(CORS(&cfg.Middleware.CORS))
	}

	// 6. Gzip - 响应压缩
	if cfg.Middleware.Gzip.Enabled {
		h.Use(Gzip(&cfg.Middleware.Gzip))
	}

	// 7. Limiter - 自适应限流（基于 CPU 使用率）
	if cfg.Middleware.Limiter.Enabled {
		h.Use(Limiter())
	}

	// 8. Pprof - 性能分析端点（仅 debug 模式）
	if cfg.Server.IsDebug() {
		RegisterPprof(h)
	}
}
