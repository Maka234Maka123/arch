package ioc

import (
	"context"
	"echo/internal/config"
	"echo/pkg/logger"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

// App 应用程序
type App struct {
	Config *config.Config
	Server *server.Hertz

	// 优雅关闭相关
	shuttingDown atomic.Bool
	shutdownCh   chan struct{}
}

// InitApp 初始化应用程序
func InitApp(configPath string) (*App, error) {
	// 1. 加载配置
	cfg, err := InitConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("init config failed: %w", err)
	}

	// 2. 初始化日志系统
	if err := InitLogger(cfg); err != nil {
		return nil, fmt.Errorf("init logger failed: %w", err)
	}

	// 3. 重定向标准库 log 到 zap
	logger.RedirectStdLog()

	logger.Info("Config loaded successfully",
		zap.String("env", cfg.Server.Mode),
		zap.String("log_level", cfg.Log.Level),
	)

	// 4. 初始化 OpenTelemetry 链路追踪
	if err := InitTracing(cfg); err != nil {
		return nil, fmt.Errorf("init tracing failed: %w", err)
	}

	// 5. 初始化 Hertz 服务器
	h := InitHertz(cfg)

	return &App{
		Config:     cfg,
		Server:     h,
		shutdownCh: make(chan struct{}),
	}, nil
}

// Run 启动应用程序
func (app *App) Run() {
	logger.Info("Starting server...",
		zap.String("host", app.Config.Server.Host),
		zap.Int("port", app.Config.Server.Port),
		zap.String("mode", app.Config.Server.Mode),
	)

	app.Server.Spin()
}

// Shutdown 优雅关闭应用程序
func (app *App) Shutdown() error {
	// 防止重复关闭
	if app.shuttingDown.Swap(true) {
		logger.Warn("Shutdown already in progress")
		return nil
	}

	logger.Info("Starting graceful shutdown...")

	// 创建带超时的 context
	timeout := time.Duration(app.Config.Server.ShutdownTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 1. 标记为正在关闭（可用于健康检查返回 503）
	close(app.shutdownCh)

	// 2. 等待一小段时间让负载均衡器感知（K8s preStop hook）
	logger.Info("Waiting for load balancer to drain connections...")
	time.Sleep(2 * time.Second)

	// 3. 关闭 Hertz 服务器（会等待当前请求完成）
	logger.Info("Shutting down HTTP server...")
	if err := app.Server.Shutdown(ctx); err != nil {
		// 忽略 "engine is not running" 错误，因为 Hertz Spin() 可能已经处理了关闭
		if err.Error() != "engine is not running" {
			logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}

	// 4. 关闭 OpenTelemetry（确保 trace 数据发送完成）
	if err := ShutdownTracing(ctx); err != nil {
		logger.Error("Tracing shutdown error", zap.Error(err))
	}

	// 5. 同步日志缓冲区
	if err := logger.Sync(); err != nil {
		// 忽略 sync 错误（在某些平台上可能会失败）
		logger.Debug("Logger sync warning", zap.Error(err))
	}

	logger.Info("Graceful shutdown completed")
	return nil
}
