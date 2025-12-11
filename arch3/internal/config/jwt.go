package config

// JWTConfig JWT认证配置
// 用于用户认证和授权的Token管理
type JWTConfig struct {
	// Secret JWT签名密钥
	// 生产环境必须使用强密钥
	// 默认值: "change-me-in-production"
	Secret string `mapstructure:"secret"`

	// AccessExpire Access Token过期时间(分钟)
	// 默认值: 15
	AccessExpire int `mapstructure:"access_expire"`

	// RefreshExpire Refresh Token过期时间(分钟)
	// 默认值: 10080 (7天)
	RefreshExpire int `mapstructure:"refresh_expire"`

	// CookieSecure Cookie 是否仅通过 HTTPS 传输
	// 生产环境应设为 true
	// 默认值: false
	CookieSecure bool `mapstructure:"cookie_secure"`
}
