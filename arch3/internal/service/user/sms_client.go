package user

import (
	"context"
	"errors"

	"arch3/pkg/response"
)

// SMSType 短信类型
type SMSType string

const (
	SMSTypeRegister SMSType = "register" // 注册
	SMSTypeLogin    SMSType = "login"    // 登录
	SMSTypeForget   SMSType = "forget"   // 忘记密码
)

// SMSClient 短信客户端接口
// 由 Service 层定义，Integration 层实现
type SMSClient interface {
	// Send 发送短信验证码
	Send(ctx context.Context, smsType SMSType, phone string) error
	// Verify 验证短信验证码
	Verify(ctx context.Context, smsType SMSType, phone, code string) error
}

// SMS 服务错误定义
var (
	// 发送限制
	ErrSMSTooFrequent = errors.New("一分钟内最多发送1次")
	ErrSMSDailyLimit  = errors.New("一天最多发送6次")

	// 发送失败
	ErrSMSSendFailed = errors.New("发送短信失败")

	// 配置错误
	ErrSMSTemplateNotFound = errors.New("短信模板未配置")

	// 验证码相关
	// 统一错误信息，避免攻击者通过不同错误信息探测验证码状态
	ErrSMSCodeInvalid   = errors.New("验证码错误或已过期")
	ErrSMSVerifyTooMany = errors.New("验证失败次数过多，请重新获取验证码")
)

// SMSToResponse 将 SMS 错误转换为业务响应
func SMSToResponse(err error) *response.Result {
	switch {
	// 短信发送限制
	case errors.Is(err, ErrSMSTooFrequent):
		return response.Err(response.CodeSMSTooFrequent, err.Error())
	case errors.Is(err, ErrSMSDailyLimit):
		return response.Err(response.CodeSMSDailyLimit, err.Error())

	// 短信验证码（统一使用 CodeSMSCodeInvalid，不区分过期和错误）
	case errors.Is(err, ErrSMSCodeInvalid):
		return response.Err(response.CodeSMSCodeInvalid, err.Error())
	case errors.Is(err, ErrSMSVerifyTooMany):
		return response.Err(response.CodeSMSVerifyTooMany, err.Error())

	// 短信服务配置/发送
	case errors.Is(err, ErrSMSTemplateNotFound), errors.Is(err, ErrSMSSendFailed):
		return response.Err(response.CodeSMSSendFailed, err.Error())

	default:
		return response.Err(response.CodeSMSSendFailed, err.Error())
	}
}
