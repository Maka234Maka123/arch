package ioc

import (
	"arch3/internal/config"
	"arch3/pkg/logger"
	"time"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

// InitDB 初始化数据库连接
//
// 功能说明:
//   - 使用 GORM 连接 MySQL 数据库
//   - 配置连接池参数
//   - 启用 OpenTelemetry tracing，自动为 SQL 操作创建 span
//   - 验证数据库连接
//
// Trace 效果:
//
//	每个 SQL 查询会自动创建 span，包含:
//	- db.system: "mysql"
//	- db.statement: SQL 语句
//	- db.operation: 操作类型 (SELECT/INSERT/UPDATE/DELETE)
//	- db.sql.table: 表名
//
// 参数:
//   - cfg: 应用配置
//
// 返回值:
//   - *gorm.DB: GORM 数据库连接实例
//   - error: 连接失败时返回错误
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// 构建 DSN
	dsn := cfg.DB.DSN()

	// GORM 日志配置
	var logLevel gormlogger.LogLevel
	switch cfg.Log.Level {
	case "debug":
		logLevel = gormlogger.Info
	case "info":
		logLevel = gormlogger.Warn
	default:
		logLevel = gormlogger.Error
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	// 启用 OpenTelemetry tracing 插件
	// 自动为每个 SQL 操作创建 span，关联到当前的 trace context
	// 注意: WithDBSystem 设置的是 db.system.name (semconv v1.30.0)
	// 火山引擎等平台可能使用旧版 db.system，需要通过 WithAttributes 额外设置
	if err := db.Use(tracing.NewPlugin(
		tracing.WithDBSystem(cfg.DB.Driver), // 设置 db.system.name
		tracing.WithAttributes(
			attribute.String("db.system", cfg.DB.Driver), // 兼容旧版 db.system 属性
			semconv.DBName(cfg.DB.Database),              // 数据库名称
		),
		tracing.WithoutQueryVariables(), // 不记录查询参数值，避免敏感数据泄露
	)); err != nil {
		logger.Warn("Failed to enable GORM tracing plugin", zap.Error(err))
	}

	// 获取底层 sql.DB 以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetime) * time.Second)

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	logger.L().Info("database connected",
		zap.String("host", cfg.DB.Host),
		zap.Int("port", cfg.DB.Port),
		zap.String("database", cfg.DB.Database),
		zap.Bool("tracing_enabled", true),
	)

	return db, nil
}
