package user

import (
	"context"

	domain "arch3/internal/domain/user"
	"arch3/pkg/jwt"
)

// Service 用户服务主接口
type Service interface {
	SMSService
	AuthService
}

// SMSService 短信服务接口
type SMSService interface {
	// SendSMS 发送短信验证码
	// smsType: login/register/forget
	SendSMS(ctx context.Context, phoneNumber, smsType string) error
}

// AuthService 认证服务接口
type AuthService interface {
	// SMSLogin 短信验证码登录（用户不存在则自动注册）
	SMSLogin(ctx context.Context, phoneNumber, smsCode string) (*domain.LoginResult, error)
	// RefreshToken 刷新 token
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
	// Logout 登出
	Logout(ctx context.Context, accessJTI, refreshJTI string) error
	// GetUserByID 根据 ID 获取用户
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
}
