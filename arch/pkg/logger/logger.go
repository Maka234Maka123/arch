package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertzzap "github.com/hertz-contrib/logger/zap"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	zapLogger *zap.Logger
	once      sync.Once
)

// Config 日志配置
type Config struct {
	Level        string // debug, info, warn, error
	Format       string // json, console
	Output       string // stdout, file, both
	Dir          string // 日志目录
	Filename     string // 日志文件名
	MaxSize      int    // 单个日志文件最大大小（MB）
	MaxBackups   int    // 最多保留的日志文件数
	MaxAge       int    // 日志文件最大保留天数
	Compress     bool   // 是否压缩
	RotationTime string // 时间轮转: daily(每天), hourly(每小时), 为空则按大小轮转
}

// Init 初始化日志系统
// 使用 hlog + zap 的形式，设置 hertz 的全局日志
func Init(cfg *Config) error {
	var err error
	once.Do(func() {
		err = initLogger(cfg)
	})
	return err
}

func initLogger(cfg *Config) error {
	// 解析日志级别
	level := parseLevel(cfg.Level)

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.DateTime),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据格式选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "file":
		writeSyncer = getFileWriter(cfg)
	case "both":
		writeSyncer = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			getFileWriter(cfg),
		)
	default: // stdout
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建 zap core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建 zap logger
	zapLogger = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// 创建 AtomicLevel 用于 hertz zap logger
	atomicLevel := zap.NewAtomicLevelAt(level)

	// 创建 hertz zap logger 并设置为全局日志
	hertzLogger := hertzzap.NewLogger(
		hertzzap.WithCoreEnc(encoder),
		hertzzap.WithCoreWs(writeSyncer),
		hertzzap.WithCoreLevel(atomicLevel),
		hertzzap.WithZapOptions(
			zap.AddCaller(),
			zap.AddCallerSkip(3),
		),
	)

	// 设置 hertz 全局日志
	hlog.SetLogger(hertzLogger)
	hlog.SetLevel(convertToHlogLevel(cfg.Level))

	return nil
}

// getFileWriter 获取文件写入器，支持日志轮转
func getFileWriter(cfg *Config) zapcore.WriteSyncer {
	// 确保日志目录存在
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		// 如果创建目录失败，回退到标准输出
		return zapcore.AddSync(os.Stdout)
	}

	// 日志文件完整路径
	logFile := filepath.Join(cfg.Dir, cfg.Filename)

	// 根据配置选择轮转策略
	switch cfg.RotationTime {
	case "daily", "hourly":
		return getTimeRotationWriter(cfg, logFile)
	default:
		// 默认按大小轮转，使用 lumberjack
		return getSizeRotationWriter(cfg, logFile)
	}
}

// getSizeRotationWriter 获取基于大小的日志轮转写入器
func getSizeRotationWriter(cfg *Config, logFile string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    cfg.MaxSize,    // MB
		MaxBackups: cfg.MaxBackups, // 保留的旧文件数
		MaxAge:     cfg.MaxAge,     // 天数
		Compress:   cfg.Compress,   // 是否压缩
		LocalTime:  true,           // 使用本地时间
	}
	return zapcore.AddSync(lumberJackLogger)
}

// getTimeRotationWriter 获取基于时间的日志轮转写入器
func getTimeRotationWriter(cfg *Config, logFile string) zapcore.WriteSyncer {
	// 根据轮转类型设置时间格式和轮转周期
	var (
		rotationTime time.Duration
		timeFormat   string
	)

	switch cfg.RotationTime {
	case "hourly":
		rotationTime = time.Hour
		timeFormat = ".%Y%m%d%H" // 每小时: app.2024010112.log
	case "daily":
		rotationTime = 24 * time.Hour
		timeFormat = ".%Y%m%d" // 每天: app.20240101.log
	default:
		rotationTime = 24 * time.Hour
		timeFormat = ".%Y%m%d"
	}

	// 构建轮转文件名模式
	// 例如: /var/log/app/app.log -> /var/log/app/app.%Y%m%d.log
	ext := filepath.Ext(logFile)
	baseName := logFile[:len(logFile)-len(ext)]
	pattern := baseName + timeFormat + ext

	// 配置 rotatelogs 选项
	options := []rotatelogs.Option{
		rotatelogs.WithRotationTime(rotationTime), // 轮转周期
		rotatelogs.WithLinkName(logFile),          // 创建软链接指向当前日志
		rotatelogs.WithClock(rotatelogs.Local),    // 使用本地时间
	}

	// 设置保留策略：优先使用 max_backups（文件数），否则使用 max_age（天数）
	// 注意：rotatelogs 不支持同时设置两者，会冲突
	if cfg.MaxBackups > 0 {
		options = append(options, rotatelogs.WithRotationCount(uint(cfg.MaxBackups)))
	} else if cfg.MaxAge > 0 {
		options = append(options, rotatelogs.WithMaxAge(time.Duration(cfg.MaxAge)*24*time.Hour))
	}

	// 创建 rotatelogs writer
	writer, err := rotatelogs.New(pattern, options...)
	if err != nil {
		// 如果创建失败，回退到标准输出
		return zapcore.AddSync(os.Stdout)
	}

	return zapcore.AddSync(writer)
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// convertToHlogLevel 转换为 hlog 日志级别
func convertToHlogLevel(level string) hlog.Level {
	switch level {
	case "debug":
		return hlog.LevelDebug
	case "info":
		return hlog.LevelInfo
	case "warn":
		return hlog.LevelWarn
	case "error":
		return hlog.LevelError
	default:
		return hlog.LevelInfo
	}
}

// L 获取 zap Logger 实例
func L() *zap.Logger {
	if zapLogger == nil {
		// 默认初始化
		Init(&Config{
			Level:  "info",
			Format: "console",
			Output: "stdout",
		})
	}
	return zapLogger
}

// S 获取 SugaredLogger 实例
func S() *zap.SugaredLogger {
	return L().Sugar()
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

// With 创建带字段的 Logger
func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}

// Sync 同步日志缓冲区
func Sync() error {
	if zapLogger != nil {
		return zapLogger.Sync()
	}
	return nil
}

// GetWriter 获取日志写入器，用于重定向标准输出
func GetWriter() io.Writer {
	return &logWriter{logger: L()}
}

// logWriter 实现 io.Writer 接口
type logWriter struct {
	logger *zap.Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(string(p))
	return len(p), nil
}
