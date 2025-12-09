package config

// DBConfig 数据库配置
// 支持MySQL等关系型数据库的连接池配置
type DBConfig struct {
	// Driver 数据库驱动类型
	// 可选值: mysql, postgres, sqlite
	// 默认值: "mysql"
	Driver string `mapstructure:"driver"`

	// DSN 数据库连接字符串
	// MySQL格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	DSN string `mapstructure:"dsn"`

	// MaxIdleConns 连接池最大空闲连接数
	// 默认值: 10
	MaxIdleConns int `mapstructure:"max_idle_conns"`

	// MaxOpenConns 连接池最大打开连接数
	// 默认值: 100
	MaxOpenConns int `mapstructure:"max_open_conns"`

	// ConnMaxLifetime 连接最大生命周期(秒)
	// 默认值: 3600 (1小时)
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
}

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
