package user

import (
	"arch3/pkg/jwt"
)

// service 用户服务实现
type service struct {
	smsClient  SMSClient
	userRepo   Repository
	jwtManager *jwt.Manager
}

// NewService 创建用户服务实例
func NewService(smsClient SMSClient, userRepo Repository, jwtManager *jwt.Manager) Service {
	return &service{
		smsClient:  smsClient,
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}
