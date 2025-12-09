package user

import (
	"context"

	"archv2/internal/dto/request"
	"archv2/internal/dto/response"
	"archv2/internal/integration/sms"
)

// UserService 用户服务主接口
type UserService interface {
	UserSMSService
	// 后续可以添加其他子接口
	// UserAuthService
	// UserProfileService
}

// UserSMSService 短信服务接口
type UserSMSService interface {
	// SendSMS 发送短信验证码
	SendSMS(ctx context.Context, req *request.SendSMSRequest) (*response.SendSMSResponse, error)
}

// userService 用户服务实现
type userService struct {
	smsClient sms.Client
}

// NewUserService 创建用户服务实例
func NewUserService(smsClient sms.Client) UserService {
	return &userService{
		smsClient: smsClient,
	}
}
