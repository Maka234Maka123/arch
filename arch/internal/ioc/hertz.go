package ioc

import (
	"context"
	"time"

	"echo/internal/config"
	"echo/internal/handler/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// InitHertz 初始化 Hertz 服务器
func InitHertz(cfg *config.Config) *server.Hertz {
	// 创建 Hertz 实例
	h := server.New(
		server.WithHostPorts(cfg.Server.Address()),
		server.WithReadTimeout(time.Duration(cfg.Server.ReadTimeout)*time.Second),
		server.WithWriteTimeout(time.Duration(cfg.Server.WriteTimeout)*time.Second),
		server.WithMaxRequestBodySize(cfg.Server.MaxRequestBody),
		server.WithExitWaitTime(time.Duration(cfg.Server.ShutdownTimeout)*time.Second),
	)

	// 注册中间件
	middleware.Register(h, cfg)

	// 注册运维路由
	registerOpsRoutes(h, cfg)

	return h
}

// registerOpsRoutes 注册运维相关路由
func registerOpsRoutes(h *server.Hertz, cfg *config.Config) {
	// 健康检查 - 存活探针
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(200, map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// Prometheus 指标
	h.GET("/metrics", middleware.GetMetricsHandler())

	// Swagger API 文档
	if cfg.Middleware.Swagger.Enabled {
		middleware.RegisterSwagger(h, &cfg.Middleware.Swagger)
	}
}
