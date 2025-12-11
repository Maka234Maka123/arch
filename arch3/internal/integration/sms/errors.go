package sms

import userservice "arch3/internal/service/user"

// 短信服务错误别名，指向 service 层定义
var (
	// 发送限制
	ErrSendTooFrequent    = userservice.ErrSMSTooFrequent
	ErrDailyLimitExceeded = userservice.ErrSMSDailyLimit

	// 发送失败
	ErrSendFailed = userservice.ErrSMSSendFailed

	// 配置错误
	ErrTemplateNotFound = userservice.ErrSMSTemplateNotFound

	// 验证码相关（统一使用 ErrCodeInvalid，不区分过期和错误）
	ErrCodeInvalid   = userservice.ErrSMSCodeInvalid
	ErrVerifyTooMany = userservice.ErrSMSVerifyTooMany
)
