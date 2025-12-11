package user

import (
	"arch3/internal/service/user"
	"arch3/pkg/jwt"
)

// Handler 用户 HTTP 处理器
type Handler struct {
	userService user.Service
	jwtManager  *jwt.Manager
}

// NewHandler 创建用户处理器实例
func NewHandler(userService user.Service, jwtManager *jwt.Manager) *Handler {
	return &Handler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}
