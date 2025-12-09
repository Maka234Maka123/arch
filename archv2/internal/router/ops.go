package router

import (
	"context"
	"net/http"
	"time"

	"archv2/internal/config"
	"archv2/internal/handler/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// ShutdownChecker 用于检查服务是否正在关闭
type ShutdownChecker func() bool

// RegisterOpsRoutes 注册运维相关路由
//
// 路由列表:
//   - GET /health    - 健康检查（存活探针）
//   - GET /ready     - 就绪检查（就绪探针，关闭时返回 503）
//   - GET /swagger/* - API 文档（可选）
//
// K8s 集成:
//   - livenessProbe: /health（检查进程是否存活）
//   - readinessProbe: /ready（检查是否准备好接收流量）
//
// 参数:
//   - isShuttingDown: 检查服务是否正在关闭的函数，可为 nil（始终返回 ready）
func RegisterOpsRoutes(h *server.Hertz, cfg *config.Config, isShuttingDown ShutdownChecker) {
	// 存活探针 - 只要进程活着就返回 200
	// 用于 K8s 判断是否需要重启容器
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// 就绪探针 - 检查服务是否准备好接收流量
	// 关闭时返回 503，让负载均衡器停止转发新请求
	h.GET("/ready", func(ctx context.Context, c *app.RequestContext) {
		if isShuttingDown != nil && isShuttingDown() {
			c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status":    "shutting_down",
				"timestamp": time.Now().Unix(),
			})
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now().Unix(),
		})
	})

	// Swagger API 文档（可选）
	if cfg.Middleware.Swagger.Enabled {
		middleware.RegisterSwagger(h, &cfg.Middleware.Swagger)
	}
}
