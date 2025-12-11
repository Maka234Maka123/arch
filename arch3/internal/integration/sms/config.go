package sms

import userservice "arch3/internal/service/user"

// Config 短信服务配置
type Config struct {
	Provider   string                         // 服务商: volcengine, aliyun
	AccessKey  string                         // AccessKey
	SecretKey  string                         // SecretKey
	SmsAccount string                         // 短信账户
	SignName   string                         // 签名
	Templates  map[userservice.SMSType]string // 模板ID映射
}
