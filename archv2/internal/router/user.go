package router

import (
	userhandler "archv2/internal/handler/user"

	"github.com/cloudwego/hertz/pkg/route"
)

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(r *route.RouterGroup, handler *userhandler.UserHandler) {
	userGroup := r.Group("/user")
	{
		// 短信验证码
		userGroup.POST("/sms", handler.SendSMS)

		// 后续可以添加其他路由
		// userGroup.POST("/login", handler.Login)
		// userGroup.POST("/register", handler.Register)
	}
}
