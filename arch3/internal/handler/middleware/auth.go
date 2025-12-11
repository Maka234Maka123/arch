package middleware

import (
	"context"

	"arch3/internal/config"
	"arch3/pkg/jwt"
	"arch3/pkg/logger"
	"arch3/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

// 公开路径白名单 - 这些路径不需要认证
// 使用 config.APIPrefix 常量确保与路由定义一致
var publicPaths = map[string]bool{
	// 认证相关接口
	"POST:" + config.APIPrefix + "/user/sms":       true, // 发送短信验证码
	"POST:" + config.APIPrefix + "/user/sms-login": true, // 验证码登录
	"POST:" + config.APIPrefix + "/user/refresh":   true, // 刷新 token
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	jwtManager *jwt.Manager
	enabled    bool
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtManager *jwt.Manager, enabled bool) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		enabled:    enabled,
	}
}

// Handle 处理认证
func (m *AuthMiddleware) Handle() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 未启用认证
		if !m.enabled {
			c.Next(ctx)
			return
		}

		// 检查是否是公开路径
		pathKey := string(c.Method()) + ":" + string(c.Path())
		if publicPaths[pathKey] {
			c.Next(ctx)
			return
		}

		// 从 cookie 获取 access token
		accessTokenString := string(c.Cookie(jwt.AccessTokenCookieKey))
		if accessTokenString == "" {
			response.Error(c, response.Err(response.CodeUnauthorized, "未登录"))
			c.Abort()
			return
		}

		// 解析 access token
		claims, err := m.jwtManager.ParseToken(accessTokenString, jwt.TokenTypeAccess)
		if err != nil {
			logger.Ctx(ctx).Warn("parse access token failed", zap.Error(err))
			response.Error(c, response.Err(response.CodeTokenInvalid, "登录已失效"))
			c.Abort()
			return
		}

		// 检查 token 是否在黑名单中
		blacklisted, err := m.jwtManager.IsTokenBlacklisted(ctx, claims.ID)
		if err != nil {
			logger.Ctx(ctx).Error("check token blacklist failed", zap.Error(err))
			response.Error(c, response.Err(response.CodeCacheError, "系统错误"))
			c.Abort()
			return
		}
		if blacklisted {
			response.Error(c, response.Err(response.CodeTokenInvalid, "登录已失效"))
			c.Abort()
			return
		}

		// 设置用户 ID 到上下文
		c.Set("userID", claims.UserID)
		c.Next(ctx)
	}
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *app.RequestContext) string {
	if v, exists := c.Get("userID"); exists {
		if userID, ok := v.(string); ok {
			return userID
		}
	}
	return ""
}
