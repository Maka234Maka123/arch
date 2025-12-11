package ioc

import (
	"arch3/internal/config"
	"arch3/pkg/logger"
	"context"
	"fmt"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitRedis 初始化 Redis 客户端
//
// 功能:
//   - 创建 Redis 客户端连接
//   - 启用 OpenTelemetry tracing hook，自动为 Redis 操作创建 span
//   - 启用 OpenTelemetry metrics hook，收集 Redis 性能指标
//
// Trace 效果:
//
//	每个 Redis 命令会自动创建 span，包含:
//	- db.system: "redis"
//	- db.statement: 命令内容 (如 "GET key")
//	- net.peer.name: Redis 地址
//	- net.peer.port: Redis 端口
func InitRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// 启用 OpenTelemetry tracing hook
	// 这会自动为每个 Redis 命令创建 span，关联到当前的 trace context
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		logger.Warn("Failed to instrument Redis tracing", zap.Error(err))
	}

	// 启用 OpenTelemetry metrics hook
	// 收集 Redis 命令的延迟、错误率等指标
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		logger.Warn("Failed to instrument Redis metrics", zap.Error(err))
	}

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("addr", cfg.Redis.Addr),
		zap.Int("db", cfg.Redis.DB),
		zap.Bool("tracing_enabled", true),
		zap.Bool("metrics_enabled", true),
	)

	return rdb, nil
}
