package sms

import userservice "arch3/internal/service/user"

// Type 短信类型别名，指向 service 层定义
type Type = userservice.SMSType

const (
	TypeRegister = userservice.SMSTypeRegister // 注册
	TypeLogin    = userservice.SMSTypeLogin    // 登录
	TypeForget   = userservice.SMSTypeForget   // 忘记密码
)
