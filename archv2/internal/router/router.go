package router

import (
	"archv2/internal/config"
	userhandler "archv2/internal/handler/user"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// Router 路由管理器
//
// 职责:
//   - 统一管理所有路由注册
//   - 维护 Handler 依赖
//   - 组织路由分组（运维路由、业务路由）
//
// 扩展指南:
//  1. 添加新 Handler 字段
//  2. 在 NewRouter 中接收并赋值
//  3. 在 registerBusinessRoutes 中调用对应的 Register*Routes
type Router struct {
	cfg            *config.Config
	userHandler    *userhandler.UserHandler
	isShuttingDown ShutdownChecker // 检查服务是否正在关闭
	// 扩展点: 添加新的 handler
	// orderHandler   *orderhandler.OrderHandler
	// productHandler *producthandler.ProductHandler
}

// NewRouter 创建路由管理器
//
// 参数:
//   - isShuttingDown: 检查服务是否正在关闭的函数，用于就绪探针
func NewRouter(cfg *config.Config, userHandler *userhandler.UserHandler, isShuttingDown ShutdownChecker) *Router {
	return &Router{
		cfg:            cfg,
		userHandler:    userHandler,
		isShuttingDown: isShuttingDown,
	}
}

// Register 注册所有路由到 Hertz 服务器
//
// 这是唯一的路由注册入口，集中管理所有路由。
//
// 路由分类:
//   - 运维路由: /health, /ready, /swagger - 不需要认证
//   - 业务路由: /api/v1/* - 按业务模块组织
func (r *Router) Register(h *server.Hertz) {
	// 1. 注册运维路由 (健康检查、就绪检查、文档)
	RegisterOpsRoutes(h, r.cfg, r.isShuttingDown)

	// 2. 注册业务路由
	r.registerBusinessRoutes(h)
}

// registerBusinessRoutes 注册业务路由
//
// 所有业务 API 统一使用 /api/v1 前缀，便于版本管理。
func (r *Router) registerBusinessRoutes(h *server.Hertz) {
	// API v1 路由组
	apiV1 := h.Group("/api/v1")
	{
		// 用户模块路由
		RegisterUserRoutes(apiV1, r.userHandler)

		// 扩展点: 添加其他业务模块路由
		// RegisterOrderRoutes(apiV1, r.orderHandler)
		// RegisterProductRoutes(apiV1, r.productHandler)
	}
}
