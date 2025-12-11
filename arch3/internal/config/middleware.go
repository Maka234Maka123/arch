package config

// MiddlewareConfig 中间件配置
// 统一管理所有HTTP中间件的配置
type MiddlewareConfig struct {
	// Auth JWT 认证配置
	Auth AuthConfig `mapstructure:"auth"`

	// CORS 跨域资源共享配置
	CORS CORSConfig `mapstructure:"cors"`

	// Gzip 响应压缩配置
	Gzip GzipConfig `mapstructure:"gzip"`

	// Limiter 请求限流配置
	Limiter LimiterConfig `mapstructure:"limiter"`

	// Swagger API文档配置
	Swagger SwaggerConfig `mapstructure:"swagger"`

	// Tracing 链路追踪配置
	Tracing TracingConfig `mapstructure:"tracing"`

	// AccessLog 访问日志配置
	AccessLog AccessLogConfig `mapstructure:"access_log"`
}

// AuthConfig JWT 认证中间件配置
type AuthConfig struct {
	// Enabled 是否启用认证
	// 默认值: false
	Enabled bool `mapstructure:"enabled"`
}

// CORSConfig CORS跨域资源共享配置
// 控制浏览器跨域请求的访问策略
type CORSConfig struct {
	// Enabled 是否启用CORS
	// 默认值: true
	Enabled bool `mapstructure:"enabled"`

	// AllowOrigins 允许的源列表
	// 使用 "*" 允许所有源(不建议生产环境使用)
	// 默认值: ["*"]
	AllowOrigins []string `mapstructure:"allow_origins"`

	// AllowMethods 允许的HTTP方法
	// 默认值: ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"]
	AllowMethods []string `mapstructure:"allow_methods"`

	// AllowHeaders 允许的请求头
	// 默认值: ["Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"]
	AllowHeaders []string `mapstructure:"allow_headers"`

	// ExposeHeaders 暴露给客户端的响应头
	// 默认值: ["Content-Length", "X-Request-ID"]
	ExposeHeaders []string `mapstructure:"expose_headers"`

	// AllowCredentials 是否允许携带凭证(Cookie)
	// 默认值: true
	AllowCredentials bool `mapstructure:"allow_credentials"`

	// MaxAge 预检请求结果缓存时间(秒)
	// 默认值: 43200 (12小时)
	MaxAge int `mapstructure:"max_age"`
}

// GzipConfig Gzip响应压缩配置
// 减少网络传输数据量
type GzipConfig struct {
	// Enabled 是否启用Gzip压缩
	// 默认值: true
	Enabled bool `mapstructure:"enabled"`

	// Level 压缩级别
	// 范围: 1-9, 级别越高压缩率越高但CPU消耗越大
	// 默认值: 6
	Level int `mapstructure:"level"`

	// MinLength 触发压缩的最小响应长度(字节)
	// 小于此长度的响应不压缩
	// 默认值: 1024
	MinLength int `mapstructure:"min_length"`

	// ExcludedExts 排除压缩的文件扩展名
	// 已压缩的资源(如图片)无需再压缩
	// 默认值: [".png", ".gif", ".jpeg", ".jpg", ".webp"]
	ExcludedExts []string `mapstructure:"excluded_exts"`
}

// LimiterConfig 请求限流配置
// 保护服务器免受过量请求冲击
type LimiterConfig struct {
	// Enabled 是否启用限流
	// 默认值: true
	Enabled bool `mapstructure:"enabled"`

	// CPUThreshold CPU 使用率阈值（自适应限流模式）
	// 范围: 0-1000，800 表示 80%
	// 当 CPU 超过阈值时开始拒绝请求，返回 503
	// 默认值: 800
	CPUThreshold int64 `mapstructure:"cpu_threshold"`

	// Rate 每秒允许的请求数（预留：固定速率模式）
	// 默认值: 100
	Rate float64 `mapstructure:"rate"`

	// Burst 突发请求数上限（预留：固定速率模式）
	// 允许短时间内超过Rate的请求数
	// 默认值: 200
	Burst int `mapstructure:"burst"`

	// KeyFunc 限流维度（预留：固定速率模式）
	// 可选值: ip(按IP限流), path(按路径限流), user(按用户限流)
	// 默认值: "ip"
	KeyFunc string `mapstructure:"key_func"`
}

// SwaggerConfig Swagger API文档配置
// 自动生成API文档和测试界面
type SwaggerConfig struct {
	// Enabled 是否启用Swagger
	// 默认值: true
	Enabled bool `mapstructure:"enabled"`

	// BasePath Swagger UI访问路径
	// 默认值: "/swagger"
	BasePath string `mapstructure:"base_path"`

	// Title API文档标题
	// 默认值: "Echo API"
	Title string `mapstructure:"title"`

	// Version API版本号
	// 默认值: "1.0.0"
	Version string `mapstructure:"version"`
}

// AccessLogConfig 访问日志配置
// 记录HTTP请求的详细信息
type AccessLogConfig struct {
	// Enabled 是否启用访问日志
	// 默认值: true
	Enabled bool `mapstructure:"enabled"`

	// ErrorOnly 是否只记录错误响应 (状态码 >= 400)
	// 启用后，正常请求不打日志，依赖 tracing span 记录
	// 默认值: false
	ErrorOnly bool `mapstructure:"error_only"`

	// SkipPaths 不记录日志的路径列表
	// 通常排除健康检查等高频路径
	// 默认值: ["/health", "/metrics", "/swagger"]
	SkipPaths []string `mapstructure:"skip_paths"`
}

// TracingConfig 链路追踪配置(OpenTelemetry)
// 分布式系统调用链追踪，支持火山引擎 APMPlus
type TracingConfig struct {
	// Enabled 是否启用链路追踪
	// 默认值: false
	Enabled bool `mapstructure:"enabled"`

	// ServiceName 服务名称
	// 在追踪系统中标识本服务
	// 默认值: "arch3"
	ServiceName string `mapstructure:"service_name"`

	// Endpoint OTLP采集端点地址
	// 火山引擎示例: "apmplus-cn-beijing.volces.com:4317"
	// 本地开发: "localhost:4317"
	Endpoint string `mapstructure:"endpoint"`

	// AppKey 火山引擎 APMPlus 认证密钥
	// 从火山引擎控制台获取
	AppKey string `mapstructure:"app_key"`

	// SampleRate 采样率
	// 范围: 0.0-1.0, 1.0表示100%采样
	// 默认值: 1.0
	SampleRate float64 `mapstructure:"sample_rate"`

	// Insecure 是否禁用TLS
	// 火山引擎生产环境应设为 false
	// 默认值: true
	Insecure bool `mapstructure:"insecure"`
}
