package config

// JWTConfig JWT认证配置
// 用于用户认证和授权的Token管理
type JWTConfig struct {
	// Secret JWT签名密钥
	// 生产环境必须使用强密钥
	// 默认值: "mock-jwt-secret-for-development-only"
	Secret string `mapstructure:"secret"`

	// Expire Access Token过期时间(小时)
	// 默认值: 24
	Expire int `mapstructure:"expire"`

	// RefreshExpire Refresh Token过期时间(小时)
	// 默认值: 168 (7天)
	RefreshExpire int `mapstructure:"refresh_expire"`
}
