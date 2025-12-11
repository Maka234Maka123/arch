package router

import (
	userhandler "arch3/internal/handler/user"
	"arch3/pkg/response"

	"github.com/cloudwego/hertz/pkg/route"
)

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(r *route.RouterGroup, handler *userhandler.Handler) {
	userGroup := r.Group("/user")
	{
		// 短信验证码
		userGroup.POST("/sms", response.Wrap(handler.SendSMS))

		// 认证路由（无需登录）
		userGroup.POST("/sms-login", response.Wrap(handler.SMSLogin))   // 验证码登录/注册
		userGroup.POST("/refresh", response.Wrap(handler.RefreshToken)) // 刷新 token
		userGroup.POST("/logout", response.Wrap(handler.Logout))        // 登出
	}
}
