package ioc

import (
	"time"

	"arch3/internal/config"
	"arch3/internal/handler/middleware"
	"arch3/pkg/jwt"

	"github.com/cloudwego/hertz/pkg/app/server"
	hertzconfig "github.com/cloudwego/hertz/pkg/common/config"
)

// initServer 创建 Hertz 服务器实例
//
// 职责范围:
//   - 配置服务器参数（端口、超时、请求体限制等）
//   - 配置链路追踪的 Server Tracer（如果启用）
//
// 不包含中间件注册和路由注册，这些由调用方单独处理。
//
// 返回值:
//   - *server.Hertz: 服务器实例
//   - *middleware.TracerConfig: Tracing 中间件配置（如果启用），否则为 nil
func initServer(cfg *config.Config) (*server.Hertz, *middleware.TracerConfig) {
	opts := []hertzconfig.Option{
		server.WithHostPorts(cfg.Server.Address()),
		server.WithReadTimeout(time.Duration(cfg.Server.ReadTimeout) * time.Second),
		server.WithWriteTimeout(time.Duration(cfg.Server.WriteTimeout) * time.Second),
		server.WithMaxRequestBodySize(cfg.Server.MaxRequestBody),
		server.WithExitWaitTime(time.Duration(cfg.Server.ShutdownTimeout) * time.Second),
	}

	// 如果启用了 tracing，添加 server tracer
	var tracerCfg *middleware.TracerConfig
	if cfg.Middleware.Tracing.Enabled {
		tracer, tc := middleware.NewServerTracer()
		opts = append(opts, tracer)
		tracerCfg = tc
	}

	return server.New(opts...), tracerCfg
}

// registerMiddleware 注册全局中间件
//
// 参数:
//   - h: Hertz 服务器实例
//   - cfg: 应用配置
//   - tracerCfg: Tracing 中间件配置（可为 nil）
//   - jwtManager: JWT 管理器（可为 nil，此时跳过认证中间件）
func registerMiddleware(h *server.Hertz, cfg *config.Config, tracerCfg *middleware.TracerConfig, jwtManager *jwt.Manager) {
	middleware.Register(h, cfg, tracerCfg, jwtManager)
}
