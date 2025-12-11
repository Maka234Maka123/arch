package ioc

import (
	"arch3/internal/config"
	"arch3/pkg/logger"
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// App 应用程序主结构体，管理所有核心组件的生命周期
//
// 生命周期:
//   - 初始化: InitApp() 按顺序初始化所有依赖
//   - 运行: Run() 启动 HTTP 服务器
//   - 关闭: Shutdown() 按逆序优雅关闭所有组件
//
// 优雅关闭机制:
//   - shuttingDown: 原子标志，防止重复关闭
//   - shutdownCh: 关闭信号通道，健康检查通过 IsShuttingDown() 感知关闭状态
type App struct {
	Config    *config.Config // 应用配置
	Container *Container     // 依赖注入容器

	// 优雅关闭状态管理
	shuttingDown atomic.Bool   // 原子标志: 是否正在关闭
	shutdownCh   chan struct{} // 关闭信号: close 后健康检查返回 503
}

// InitApp 初始化应用程序
//
// 初始化顺序:
//  1. Config - 所有组件的基础依赖
//  2. Logger - 后续初始化需要日志记录
//  3. Container - 通过 ManualInitialize 构建依赖注入容器
//
// 错误处理策略:
//   - 任何组件初始化失败都返回错误，服务无法启动
func InitApp(configPath string) (*App, error) {
	// 提前创建关闭信号通道，用于传递给路由
	shutdownCh := make(chan struct{})

	// 创建关闭检查器函数
	isShuttingDown := func() bool {
		select {
		case <-shutdownCh:
			return true
		default:
			return false
		}
	}

	// 1. 加载配置
	cfg, err := InitConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("init config failed: %w", err)
	}

	// 2. 初始化日志系统
	if err := InitLogger(cfg); err != nil {
		return nil, fmt.Errorf("init logger failed: %w", err)
	}
	logger.RedirectStdLog()

	logger.Info("Config loaded successfully",
		zap.String("env", cfg.Server.Mode),
		zap.String("log_level", cfg.Log.Level),
	)

	// 3. 构建依赖注入容器
	container, err := ManualInitialize(cfg, isShuttingDown)
	if err != nil {
		return nil, fmt.Errorf("init container failed: %w", err)
	}

	logger.Info("All components initialized successfully")

	return &App{
		Config:     cfg,
		Container:  container,
		shutdownCh: shutdownCh,
	}, nil
}

// Run 启动应用程序（阻塞）
//
// 调用 Hertz 的 Spin() 方法启动 HTTP 服务器，此方法会阻塞直到服务器关闭。
// 优雅关闭由 Shutdown() 方法触发。
//
// 启动信号:
//   - "Server started successfully" 日志表示服务已开始监听，可接收请求
//   - K8s readinessProbe 应在此日志后开始检查 /ready 端点
func (app *App) Run() {
	addr := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)

	logger.Info("Starting server...",
		zap.String("address", addr),
		zap.String("mode", app.Config.Server.Mode),
	)

	// 输出启动完成信号（用于容器编排和监控）
	// 注意: Hertz Spin() 会阻塞，但在阻塞前会完成监听端口的绑定
	logger.Info("Server started successfully",
		zap.String("address", addr),
		zap.String("health_endpoint", "/health"),
		zap.String("ready_endpoint", "/ready"),
	)

	app.Container.Server.Spin()
}

// Shutdown 优雅关闭应用程序
//
// 关闭顺序 (与初始化相反):
//  1. 标记关闭状态  - 健康检查返回 503，LB 停止发送新请求
//  2. 等待 LB 感知  - 默认 2 秒，可配合 K8s preStop hook
//  3. 关闭 HTTP     - 等待当前请求完成（50% 超时时间）
//  4. 关闭 Tracing  - 确保 trace/metrics 数据发送完成（25% 超时时间）
//  5. 关闭 Redis    - 释放连接池资源（25% 超时时间）
//  6. 同步日志      - 刷新日志缓冲区
//
// 超时分配策略:
//   - HTTP 关闭: 50% (需要等待请求完成)
//   - Tracing 关闭: 25% (需要发送缓冲数据)
//   - Redis 关闭: 25% (通常很快)
//
// 注意事项:
//   - 使用 atomic.Bool 防止重复关闭
//   - 每个步骤有独立超时，避免一个步骤耗尽所有时间
func (app *App) Shutdown() error {
	// 防止重复关闭
	if app.shuttingDown.Swap(true) {
		logger.Warn("Shutdown already in progress")
		return nil
	}

	logger.Info("Starting graceful shutdown...")

	// 计算各步骤的超时时间
	totalTimeout := time.Duration(app.Config.Server.ShutdownTimeout) * time.Second
	httpTimeout := totalTimeout / 2    // 50% 给 HTTP
	tracingTimeout := totalTimeout / 4 // 25% 给 Tracing
	infraTimeout := totalTimeout / 4   // 25% 给基础设施

	// 1. 标记为正在关闭 - 健康检查 /ready 将返回 503
	close(app.shutdownCh)

	// 2. 等待负载均衡器感知并停止转发新请求
	logger.Info("Waiting for load balancer to drain connections...")
	time.Sleep(2 * time.Second)

	// 3. 关闭 Hertz 服务器 - 等待当前正在处理的请求完成
	logger.Info("Shutting down HTTP server...", zap.Duration("timeout", httpTimeout))
	httpCtx, httpCancel := context.WithTimeout(context.Background(), httpTimeout)
	if err := app.Container.Server.Shutdown(httpCtx); err != nil {
		if err.Error() != "engine is not running" {
			logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}
	httpCancel()

	// 4. 关闭 OpenTelemetry - 确保 trace/metrics 数据完整发送到后端
	if app.Container.Tracing != nil {
		logger.Info("Shutting down tracing...", zap.Duration("timeout", tracingTimeout))
		tracingCtx, tracingCancel := context.WithTimeout(context.Background(), tracingTimeout)
		if err := app.Container.Tracing.Shutdown(tracingCtx); err != nil {
			logger.Error("Tracing shutdown error", zap.Error(err))
		}
		tracingCancel()
	}

	// 5. 关闭基础设施资源（DB + Redis）
	if app.Container.Infra != nil {
		logger.Info("Closing infrastructure resources...", zap.Duration("timeout", infraTimeout))
		if err := app.Container.Infra.Close(); err != nil {
			logger.Error("Infrastructure close error", zap.Error(err))
		} else {
			logger.Info("Infrastructure resources closed")
		}
	}

	// 6. 同步日志缓冲区 - 确保所有日志都已写入
	if err := logger.Sync(); err != nil {
		logger.Debug("Logger sync warning", zap.Error(err))
	}

	logger.Info("Graceful shutdown completed")
	return nil
}
