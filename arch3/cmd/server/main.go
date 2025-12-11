package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"arch3/internal/ioc"
	"arch3/pkg/logger"

	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config/config.dev.yaml", "配置文件路径 (默认: config/config.dev.yaml)")
}

func main() {
	if err := run(); err != nil {
		// 尝试用 logger 输出（如果已初始化）
		if logger.Initialized() {
			logger.Fatal("Application failed to start", zap.Error(err))
		}
		// 兜底：输出到 stderr
		_, _ = fmt.Fprintf(os.Stderr, "Fatal: %v\n", err)
		os.Exit(1)
	}
}

// run 是程序的真正入口，返回 error 便于测试和优雅处理
func run() error {
	flag.Parse()

	// 创建可取消的 context，用于协调关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化应用
	app, err := ioc.InitApp(configPath)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	// 设置信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动信号监听 goroutine
	go func() {
		select {
		case sig := <-quit:
			logger.Info("Received shutdown signal",
				zap.String("signal", sig.String()),
			)
			if err := app.Shutdown(); err != nil {
				logger.Error("Server shutdown error", zap.Error(err))
			}
			cancel() // 通知主 goroutine 退出
		case <-ctx.Done():
			// context 被取消，正常退出
		}
	}()

	// 启动服务（阻塞直到服务器关闭）
	app.Run()

	return nil
}
