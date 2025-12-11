package user

import (
	domain "arch3/internal/domain/user"
	"arch3/pkg/ptr"
)

// SMSLoginResponse 短信登录响应
// 注意：仅返回必要的用户信息，敏感信息（身份证号等）不在登录响应中返回
type SMSLoginResponse struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	UserName    string `json:"user_name"`
	AvatarURL   string `json:"avatar_url"`
	Email       string `json:"email"`
	DeviceID    string `json:"device_id"`
	Type        string `json:"type"` // "login" 或 "register"
}

// NewSMSLoginResponse 从 domain.User 创建登录响应
func NewSMSLoginResponse(u *domain.User, isNew bool) *SMSLoginResponse {
	loginType := "login"
	if isNew {
		loginType = "register"
	}
	return &SMSLoginResponse{
		ID:          u.UserID,
		PhoneNumber: u.PhoneNumber,
		UserName:    u.UserName,
		AvatarURL:   ptr.Value(u.AvatarURL),
		Email:       ptr.Value(u.Email),
		DeviceID:    ptr.Value(u.DeviceID),
		Type:        loginType,
	}
}
