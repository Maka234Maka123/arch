package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Log        LogConfig        `mapstructure:"log"`
	DB         DBConfig         `mapstructure:"db"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Middleware MiddlewareConfig `mapstructure:"middleware"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Mode            string `mapstructure:"mode"` // debug, release, test
	ReadTimeout     int    `mapstructure:"read_timeout"`
	WriteTimeout    int    `mapstructure:"write_timeout"`
	MaxRequestBody  int    `mapstructure:"max_request_body"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level        string `mapstructure:"level"`         // debug, info, warn, error
	Format       string `mapstructure:"format"`        // json, console
	Output       string `mapstructure:"output"`        // stdout, file, both
	Dir          string `mapstructure:"dir"`           // 日志目录
	Filename     string `mapstructure:"filename"`      // 日志文件名
	MaxSize      int    `mapstructure:"max_size"`      // 单个日志文件最大大小（MB）
	MaxBackups   int    `mapstructure:"max_backups"`   // 最多保留的日志文件数
	MaxAge       int    `mapstructure:"max_age"`       // 日志文件最大保留天数
	Compress     bool   `mapstructure:"compress"`      // 是否压缩
	RotationTime string `mapstructure:"rotation_time"` // 时间轮转: daily(每天), hourly(每小时), 为空则按大小轮转
}

// DBConfig 数据库配置
type DBConfig struct {
	Driver          string `mapstructure:"driver"`
	DSN             string `mapstructure:"dsn"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	Expire        int    `mapstructure:"expire"`         // 过期时间（小时）
	RefreshExpire int    `mapstructure:"refresh_expire"` // 刷新过期时间（小时）
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	CORS      CORSConfig      `mapstructure:"cors"`
	Gzip      GzipConfig      `mapstructure:"gzip"`
	Limiter   LimiterConfig   `mapstructure:"limiter"`
	Swagger   SwaggerConfig   `mapstructure:"swagger"`
	AccessLog AccessLogConfig `mapstructure:"access_log"`
	Tracing   TracingConfig   `mapstructure:"tracing"`
}

// CORSConfig CORS 跨域配置
type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // 预检请求缓存时间（秒）
}

// GzipConfig Gzip 压缩配置
type GzipConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	Level        int      `mapstructure:"level"`         // 压缩级别: 1-9
	MinLength    int      `mapstructure:"min_length"`    // 最小压缩长度（字节）
	ExcludedExts []string `mapstructure:"excluded_exts"` // 排除的扩展名
}

// LimiterConfig 限流配置
type LimiterConfig struct {
	Enabled   bool    `mapstructure:"enabled"`
	Rate      float64 `mapstructure:"rate"`       // 每秒请求数
	Burst     int     `mapstructure:"burst"`      // 突发请求数
	KeyFunc   string  `mapstructure:"key_func"`   // 限流 key: ip, path, user
}

// SwaggerConfig Swagger 文档配置
type SwaggerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	BasePath string `mapstructure:"base_path"` // Swagger UI 访问路径
	Title    string `mapstructure:"title"`
	Version  string `mapstructure:"version"`
}

// AccessLogConfig 访问日志配置
type AccessLogConfig struct {
	Enabled    bool     `mapstructure:"enabled"`
	Format     string   `mapstructure:"format"`      // json, text
	SkipPaths  []string `mapstructure:"skip_paths"`  // 跳过记录的路径
	TimeFormat string   `mapstructure:"time_format"` // 时间格式
}

// TracingConfig 链路追踪配置 (OpenTelemetry)
type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	ServiceName  string  `mapstructure:"service_name"`
	Endpoint     string  `mapstructure:"endpoint"`      // OTLP 端点
	SampleRate   float64 `mapstructure:"sample_rate"`   // 采样率 0.0-1.0
	Insecure     bool    `mapstructure:"insecure"`      // 是否禁用 TLS
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 如果指定了配置文件路径
	if path != "" {
		v.SetConfigFile(path)
	} else {
		// 设置配置文件搜索路径
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/echo/")

		// 根据环境变量选择配置文件
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "dev"
		}

		// 先尝试加载环境特定的配置
		v.SetConfigName(fmt.Sprintf("config.%s", env))
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			// 如果找不到环境特定的配置，尝试加载默认配置
			v.SetConfigName("config")
			if err := v.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return nil, fmt.Errorf("read config file failed: %w", err)
				}
				// 配置文件不存在，使用默认配置
			}
		}
	}

	// 读取配置文件
	if path != "" {
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file failed: %w", err)
		}
	}

	// 支持环境变量覆盖
	v.SetEnvPrefix("ECHO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 解析配置到结构体
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 处理日志目录
	if cfg.Log.Dir != "" && !filepath.IsAbs(cfg.Log.Dir) {
		// 相对路径转换为绝对路径
		wd, err := os.Getwd()
		if err == nil {
			cfg.Log.Dir = filepath.Join(wd, cfg.Log.Dir)
		}
	}

	return cfg, nil
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// Server 默认值
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.max_request_body", 67108864) // 64MB
	v.SetDefault("server.shutdown_timeout", 10)

	// Log 默认值
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.output", "both")
	v.SetDefault("log.dir", "./logs")
	v.SetDefault("log.filename", "app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)

	// DB 默认值
	v.SetDefault("db.driver", "mysql")
	v.SetDefault("db.max_idle_conns", 10)
	v.SetDefault("db.max_open_conns", 100)
	v.SetDefault("db.conn_max_lifetime", 3600)

	// Redis 默认值
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)

	// JWT 默认值
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.expire", 24)
	v.SetDefault("jwt.refresh_expire", 168)

	// Middleware 默认值
	// CORS
	v.SetDefault("middleware.cors.enabled", true)
	v.SetDefault("middleware.cors.allow_origins", []string{"*"})
	v.SetDefault("middleware.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	v.SetDefault("middleware.cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"})
	v.SetDefault("middleware.cors.expose_headers", []string{"Content-Length", "X-Request-ID"})
	v.SetDefault("middleware.cors.allow_credentials", true)
	v.SetDefault("middleware.cors.max_age", 43200) // 12 小时

	// Gzip
	v.SetDefault("middleware.gzip.enabled", true)
	v.SetDefault("middleware.gzip.level", 6)
	v.SetDefault("middleware.gzip.min_length", 1024)
	v.SetDefault("middleware.gzip.excluded_exts", []string{".png", ".gif", ".jpeg", ".jpg", ".webp"})

	// Limiter
	v.SetDefault("middleware.limiter.enabled", true)
	v.SetDefault("middleware.limiter.rate", 100)    // 每秒 100 个请求
	v.SetDefault("middleware.limiter.burst", 200)   // 突发 200 个请求
	v.SetDefault("middleware.limiter.key_func", "ip")

	// Swagger
	v.SetDefault("middleware.swagger.enabled", true)
	v.SetDefault("middleware.swagger.base_path", "/swagger")
	v.SetDefault("middleware.swagger.title", "Echo API")
	v.SetDefault("middleware.swagger.version", "1.0.0")

	// AccessLog
	v.SetDefault("middleware.access_log.enabled", true)
	v.SetDefault("middleware.access_log.format", "json")
	v.SetDefault("middleware.access_log.skip_paths", []string{"/health", "/metrics", "/swagger"})
	v.SetDefault("middleware.access_log.time_format", "2006-01-02 15:04:05")

	// Tracing (OpenTelemetry)
	v.SetDefault("middleware.tracing.enabled", false)
	v.SetDefault("middleware.tracing.service_name", "echo-service")
	v.SetDefault("middleware.tracing.endpoint", "localhost:4317")
	v.SetDefault("middleware.tracing.sample_rate", 1.0)
	v.SetDefault("middleware.tracing.insecure", true)
}

// Address 获取服务器地址
func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// IsProd 是否生产环境
func (s *ServerConfig) IsProd() bool {
	return s.Mode == "release"
}

// IsDebug 是否调试模式
func (s *ServerConfig) IsDebug() bool {
	return s.Mode == "debug"
}
