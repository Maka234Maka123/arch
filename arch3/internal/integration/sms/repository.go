package sms

import (
	"context"
	"time"
)

// CodeRepository 验证码存储接口
type CodeRepository interface {
	// StoreCode 存储验证码
	StoreCode(ctx context.Context, smsType Type, phone, code string, ttl time.Duration) error
	// GetCode 获取验证码
	GetCode(ctx context.Context, smsType Type, phone string) (string, error)
	// DeleteCode 删除验证码
	DeleteCode(ctx context.Context, smsType Type, phone string) error
	// GetSendCount 获取发送次数 (minute, day)
	GetSendCount(ctx context.Context, smsType Type, phone string) (minuteCount, dayCount int, err error)
	// IncrSendCount 增加发送次数
	IncrSendCount(ctx context.Context, smsType Type, phone string) error
	// GetVerifyFailCount 获取验证失败次数
	GetVerifyFailCount(ctx context.Context, smsType Type, phone string) (int, error)
	// IncrVerifyFailCount 增加验证失败次数
	IncrVerifyFailCount(ctx context.Context, smsType Type, phone string) error
	// ResetVerifyFailCount 重置验证失败次数
	ResetVerifyFailCount(ctx context.Context, smsType Type, phone string) error
}
