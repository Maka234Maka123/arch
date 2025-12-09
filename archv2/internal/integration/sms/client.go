package sms

import "context"

// Type 短信类型
type Type string

const (
	TypeRegister Type = "register" // 注册
	TypeLogin    Type = "login"    // 登录
	TypeForget   Type = "forget"   // 忘记密码
)

// Client 短信客户端接口
type Client interface {
	// Send 发送短信验证码
	Send(ctx context.Context, smsType Type, phone string) error
	// Verify 验证短信验证码
	Verify(ctx context.Context, smsType Type, phone, code string) error
}

// Config 短信服务配置
type Config struct {
	Provider   string            // 服务商: volcengine, aliyun
	AccessKey  string            // AccessKey
	SecretKey  string            // SecretKey
	SmsAccount string            // 短信账户
	SignName   string            // 签名
	Templates  map[Type]string   // 模板ID映射
}

// NewClient 根据配置创建对应的客户端
func NewClient(cfg *Config) (Client, error) {
	switch cfg.Provider {
	case "volcengine":
		return NewVolcengineClient(cfg)
	default:
		return NewVolcengineClient(cfg)
	}
}
