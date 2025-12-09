package ioc

import (
	"time"

	"archv2/internal/config"
	"archv2/internal/handler/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
	hertzconfig "github.com/cloudwego/hertz/pkg/common/config"
)

// InitHertz 初始化 Hertz 服务器
//
// 职责范围:
//   - 配置服务器参数（端口、超时、请求体限制等）
//   - 配置链路追踪的 Server Tracer
//   - 注册全局中间件
//
// 路由注册不在此处理，统一由 router.Router.Register() 管理。
//
// Tracing 配置说明:
//
//	Hertz 的 tracing 需要两步配置:
//	1. 创建时传入 ServerTracer (作为 server option)
//	2. 中间件注册时传入 TracerConfig (用于创建中间件)
//
//	两者必须配对使用，所以需要将 tracerCfg 传递给 middleware.Register()。
func InitHertz(cfg *config.Config) *server.Hertz {
	// 构建服务器选项
	opts := []hertzconfig.Option{
		server.WithHostPorts(cfg.Server.Address()),                                  // 监听地址
		server.WithReadTimeout(time.Duration(cfg.Server.ReadTimeout) * time.Second), // 读取超时
		server.WithWriteTimeout(time.Duration(cfg.Server.WriteTimeout) * time.Second),
		server.WithMaxRequestBodySize(cfg.Server.MaxRequestBody),                         // 最大请求体
		server.WithExitWaitTime(time.Duration(cfg.Server.ShutdownTimeout) * time.Second), // 优雅关闭等待时间
	}

	// 如果启用了 tracing，添加 server tracer
	// tracer 需要在创建 Hertz 时传入，tracerCfg 用于创建中间件
	var tracerCfg *middleware.TracerConfig
	if cfg.Middleware.Tracing.Enabled {
		tracer, tc := middleware.NewServerTracer()
		opts = append(opts, tracer)
		tracerCfg = tc
	}

	// 创建 Hertz 实例
	h := server.New(opts...)

	// 注册全局中间件（传入 tracer 配置以创建 tracing 中间件）
	middleware.Register(h, cfg, tracerCfg)

	return h
}
