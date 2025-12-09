package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用主配置
// 聚合所有子配置模块
type Config struct {
	// Server HTTP服务器配置
	Server ServerConfig `mapstructure:"server"`

	// Log 日志配置
	Log LogConfig `mapstructure:"log"`

	// DB 数据库配置
	DB DBConfig `mapstructure:"db"`

	// Redis Redis缓存配置
	Redis RedisConfig `mapstructure:"redis"`

	// JWT JWT认证配置
	JWT JWTConfig `mapstructure:"jwt"`

	// SMS 短信服务配置
	SMS SMSConfig `mapstructure:"sms"`

	// Middleware 中间件配置
	Middleware MiddlewareConfig `mapstructure:"middleware"`
}

// Load 加载配置文件
// path: 配置文件路径, 为空时自动搜索默认路径
// 支持环境变量覆盖: ECHO_<SECTION>_<KEY>
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
		v.AddConfigPath("/etc/archv2/")

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
	// 格式: ECHO_SERVER_PORT=8080
	v.SetEnvPrefix("ECHO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 解析配置到结构体
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 处理日志目录: 相对路径转换为绝对路径
	if cfg.Log.Dir != "" && !filepath.IsAbs(cfg.Log.Dir) {
		wd, err := os.Getwd()
		if err == nil {
			cfg.Log.Dir = filepath.Join(wd, cfg.Log.Dir)
		}
	}

	// 生产环境安全验证
	if err := validateProductionConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validateProductionConfig 验证生产环境必须的配置项
// 确保敏感配置不使用默认值，避免安全隐患
func validateProductionConfig(cfg *Config) error {
	if !cfg.Server.IsProd() {
		return nil // 非生产环境跳过验证
	}

	// JWT Secret 不能使用默认值
	if cfg.JWT.Secret == "mock-jwt-secret-for-development-only" {
		return fmt.Errorf("production config error: jwt.secret must be configured (use ECHO_JWT_SECRET env var)")
	}

	// JWT Secret 长度检查（至少 32 字符）
	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("production config error: jwt.secret must be at least 32 characters")
	}

	return nil
}

// setDefaults 设置所有配置项的默认值
func setDefaults(v *viper.Viper) {
	// Server 默认值
	setServerDefaults(v)

	// Log 默认值
	setLogDefaults(v)

	// DB 默认值
	setDBDefaults(v)

	// Redis 默认值
	setRedisDefaults(v)

	// JWT 默认值
	setJWTDefaults(v)

	// Middleware 默认值
	setMiddlewareDefaults(v)

	// SMS 默认值
	setSMSDefaults(v)
}

// setServerDefaults 设置服务器配置默认值
func setServerDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.max_request_body", 67108864) // 64MB
	v.SetDefault("server.shutdown_timeout", 10)
}

// setLogDefaults 设置日志配置默认值
func setLogDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.output", "both")
	v.SetDefault("log.dir", "./logs")
	v.SetDefault("log.filename", "app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)
}

// setDBDefaults 设置数据库配置默认值
func setDBDefaults(v *viper.Viper) {
	v.SetDefault("db.driver", "mysql")
	v.SetDefault("db.max_idle_conns", 10)
	v.SetDefault("db.max_open_conns", 100)
	v.SetDefault("db.conn_max_lifetime", 3600)
}

// setRedisDefaults 设置Redis配置默认值
func setRedisDefaults(v *viper.Viper) {
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)
}

// setJWTDefaults 设置JWT配置默认值
func setJWTDefaults(v *viper.Viper) {
	v.SetDefault("jwt.secret", "mock-jwt-secret-for-development-only")
	v.SetDefault("jwt.expire", 24)
	v.SetDefault("jwt.refresh_expire", 168)
}

// setMiddlewareDefaults 设置中间件配置默认值
func setMiddlewareDefaults(v *viper.Viper) {
	// CORS
	v.SetDefault("middleware.cors.enabled", true)
	v.SetDefault("middleware.cors.allow_origins", []string{"*"})
	v.SetDefault("middleware.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	v.SetDefault("middleware.cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"})
	v.SetDefault("middleware.cors.expose_headers", []string{"Content-Length", "X-Request-ID"})
	v.SetDefault("middleware.cors.allow_credentials", true)
	v.SetDefault("middleware.cors.max_age", 43200) // 12小时

	// Gzip
	v.SetDefault("middleware.gzip.enabled", true)
	v.SetDefault("middleware.gzip.level", 6)
	v.SetDefault("middleware.gzip.min_length", 1024)
	v.SetDefault("middleware.gzip.excluded_exts", []string{".png", ".gif", ".jpeg", ".jpg", ".webp"})

	// Limiter
	v.SetDefault("middleware.limiter.enabled", true)
	v.SetDefault("middleware.limiter.rate", 100)
	v.SetDefault("middleware.limiter.burst", 200)
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
	v.SetDefault("middleware.tracing.service_name", "archv2-service")
	v.SetDefault("middleware.tracing.endpoint", "localhost:4317")
	v.SetDefault("middleware.tracing.sample_rate", 1.0)
	v.SetDefault("middleware.tracing.insecure", true)
}

// setSMSDefaults 设置短信服务配置默认值
func setSMSDefaults(v *viper.Viper) {
	v.SetDefault("sms.provider", "volcengine")
	v.SetDefault("sms.access_key", "")
	v.SetDefault("sms.secret_key", "")
	v.SetDefault("sms.sms_account", "")
	v.SetDefault("sms.sign_name", "")
	v.SetDefault("sms.templates.login", "")
	v.SetDefault("sms.templates.register", "")
	v.SetDefault("sms.templates.forget", "")
}
