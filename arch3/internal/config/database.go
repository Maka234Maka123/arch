package config

import "strconv"

// DBConfig 数据库配置
// 支持MySQL等关系型数据库的连接池配置
type DBConfig struct {
	// Driver 数据库驱动类型
	// 可选值: mysql, postgres, sqlite
	// 默认值: "mysql"
	Driver string `mapstructure:"driver"`

	// Host 数据库主机地址
	// 默认值: "127.0.0.1"
	Host string `mapstructure:"host"`

	// Port 数据库端口
	// 默认值: 3306
	Port int `mapstructure:"port"`

	// Username 数据库用户名
	// 默认值: "root"
	Username string `mapstructure:"username"`

	// Password 数据库密码
	Password string `mapstructure:"password"`

	// Database 数据库名称
	Database string `mapstructure:"database"`

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

// DSN 返回数据库连接字符串
// MySQL格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
func (c *DBConfig) DSN() string {
	return c.Username + ":" + c.Password + "@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" + c.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}
