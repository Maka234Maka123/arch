package config

import "fmt"

// ServerConfig 服务器配置
// 包含HTTP服务器的基本配置参数
type ServerConfig struct {
	// Host 服务器监听地址
	// 默认值: "0.0.0.0" (监听所有网络接口)
	Host string `mapstructure:"host"`

	// Port 服务器监听端口
	// 默认值: 8080
	Port int `mapstructure:"port"`

	// Mode 运行模式
	// 可选值: debug(调试模式), release(生产模式), test(测试模式)
	// 默认值: "debug"
	Mode string `mapstructure:"mode"`

	// ReadTimeout 读取请求超时时间(秒)
	// 默认值: 30
	ReadTimeout int `mapstructure:"read_timeout"`

	// WriteTimeout 写入响应超时时间(秒)
	// 默认值: 30
	WriteTimeout int `mapstructure:"write_timeout"`

	// MaxRequestBody 最大请求体大小(字节)
	// 默认值: 67108864 (64MB)
	MaxRequestBody int `mapstructure:"max_request_body"`

	// ShutdownTimeout 优雅关闭超时时间(秒)
	// 默认值: 10
	ShutdownTimeout int `mapstructure:"shutdown_timeout"`
}

// Address 获取服务器完整监听地址
// 返回格式: "host:port"
func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// IsProd 判断是否为生产环境
func (s *ServerConfig) IsProd() bool {
	return s.Mode == "release"
}

// IsDebug 判断是否为调试模式
func (s *ServerConfig) IsDebug() bool {
	return s.Mode == "debug"
}
