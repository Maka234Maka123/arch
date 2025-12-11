package config

// RedisConfig Redis缓存配置
// 用于缓存、会话存储、限流等场景
type RedisConfig struct {
	// Addr Redis服务器地址
	// 格式: host:port
	// 默认值: "localhost:6379"
	Addr string `mapstructure:"addr"`

	// Password Redis密码
	// 无密码时留空
	// 默认值: ""
	Password string `mapstructure:"password"`

	// DB Redis数据库编号
	// 范围: 0-15
	// 默认值: 0
	DB int `mapstructure:"db"`

	// PoolSize 连接池大小
	// 默认值: 100
	PoolSize int `mapstructure:"pool_size"`
}
