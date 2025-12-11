package config

// SMSConfig 短信服务配置
// 支持多种短信服务提供商
type SMSConfig struct {
	// Provider 短信服务提供商
	// 可选值: volcengine(火山引擎), aliyun(阿里云)
	// 默认值: "volcengine"
	Provider string `mapstructure:"provider"`

	// AccessKey 访问密钥ID
	// 从服务提供商控制台获取
	AccessKey string `mapstructure:"access_key"`

	// SecretKey 访问密钥Secret
	// 从服务提供商控制台获取
	SecretKey string `mapstructure:"secret_key"`

	// SmsAccount 短信账号
	// 部分服务提供商需要此参数
	SmsAccount string `mapstructure:"sms_account"`

	// SignName 短信签名
	// 需在服务提供商控制台申请审核
	SignName string `mapstructure:"sign_name"`

	// Templates 短信模板配置
	Templates SMSTemplatesConfig `mapstructure:"templates"`
}

// SMSTemplatesConfig 短信模板配置
// 不同场景使用不同的短信模板
type SMSTemplatesConfig struct {
	// Login 登录验证码模板ID
	Login string `mapstructure:"login"`

	// Register 注册验证码模板ID
	Register string `mapstructure:"register"`

	// Forget 忘记密码验证码模板ID
	Forget string `mapstructure:"forget"`
}
