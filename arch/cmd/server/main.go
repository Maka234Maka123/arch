package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"echo/internal/ioc"
	"echo/pkg/logger"

	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config/config.yaml", "配置文件路径 (默认: config/config.yaml)")
}

func main() {
	flag.Parse()

	// 初始化应用
	app, err := ioc.InitApp(configPath)
	if err != nil {
		panic(err)
	}

	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		logger.Info("Received shutdown signal",
			zap.String("signal", sig.String()),
		)

		if err := app.Shutdown(); err != nil {
			logger.Error("Server shutdown error", zap.Error(err))
		}
		os.Exit(0)
	}()

	// 启动服务
	app.Run()
}
