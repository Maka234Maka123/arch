package config

// LogConfig 日志配置
// 支持多种输出方式、日志轮转和压缩
type LogConfig struct {
	// Level 日志级别
	// 可选值: debug, info, warn, error
	// 默认值: "info"
	Level string `mapstructure:"level"`

	// Format 日志格式
	// 可选值: json(JSON格式), console(控制台友好格式)
	// 默认值: "console"
	Format string `mapstructure:"format"`

	// Output 输出目标
	// 可选值: stdout(标准输出), file(文件), both(同时输出)
	// 默认值: "both"
	Output string `mapstructure:"output"`

	// Dir 日志文件目录
	// 默认值: "./logs"
	Dir string `mapstructure:"dir"`

	// Filename 日志文件名
	// 默认值: "app.log"
	Filename string `mapstructure:"filename"`

	// MaxSize 单个日志文件最大大小(MB)
	// 超过此大小将触发轮转
	// 默认值: 100
	MaxSize int `mapstructure:"max_size"`

	// MaxBackups 最多保留的日志文件数
	// 默认值: 10
	MaxBackups int `mapstructure:"max_backups"`

	// MaxAge 日志文件最大保留天数
	// 默认值: 30
	MaxAge int `mapstructure:"max_age"`

	// Compress 是否压缩归档日志文件
	// 默认值: true
	Compress bool `mapstructure:"compress"`

	// RotationTime 时间轮转策略
	// 可选值: daily(每天), hourly(每小时), 空字符串(按大小轮转)
	// 默认值: ""
	RotationTime string `mapstructure:"rotation_time"`
}
