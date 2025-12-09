package user

import (
	"archv2/internal/service/user"
)

// UserHandler 用户 HTTP 处理器
type UserHandler struct {
	userService user.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
