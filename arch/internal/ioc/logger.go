package ioc

import (
	"echo/internal/config"
	"echo/pkg/logger"
)

// InitLogger 初始化日志系统
func InitLogger(cfg *config.Config) error {
	return logger.Init(&logger.Config{
		Level:        cfg.Log.Level,
		Format:       cfg.Log.Format,
		Output:       cfg.Log.Output,
		Dir:          cfg.Log.Dir,
		Filename:     cfg.Log.Filename,
		MaxSize:      cfg.Log.MaxSize,
		MaxBackups:   cfg.Log.MaxBackups,
		MaxAge:       cfg.Log.MaxAge,
		Compress:     cfg.Log.Compress,
		RotationTime: cfg.Log.RotationTime,
	})
}
