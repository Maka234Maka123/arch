package user

import (
	"context"

	"arch3/pkg/jwt"
	"arch3/pkg/response"
	"arch3/pkg/tracer"

	"github.com/cloudwego/hertz/pkg/app"
)

// SMSLogin 短信验证码登录
// @Summary 短信验证码登录/注册
// @Description 使用短信验证码登录，如果用户不存在则自动注册。返回用户信息并设置双 token（access + refresh）到 cookie
// @Tags users
// @Accept json
// @Produce json
// @Param request body SMSLoginRequest true "登录请求"
// @Success 200 {object} response.Result{data=SMSLoginResponse}
// @Router /api/v1/user/sms-login [post]
func (h *Handler) SMSLogin(ctx context.Context, c *app.RequestContext) error {
	ctx, span := tracer.Start(ctx, "handler.SMSLogin")
	defer span.End()

	var req SMSLoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		tracer.RecordError(span, err)
		return response.Validation(err.Error())
	}

	// 记录关键属性（手机号已脱敏）
	span.SetAttributes(
		tracer.String(tracer.AttrPhoneMasked, tracer.MaskPhone(req.PhoneNumber)),
	)

	result, err := h.userService.SMSLogin(ctx, req.PhoneNumber, req.SMSCode)
	if err != nil {
		tracer.RecordError(span, err)
		return err
	}

	// 设置 token 到 cookie
	h.jwtManager.SetTokensInCookie(c, result.TokenPair)

	// 返回用户信息（平铺结构，与旧框架保持一致）
	return response.Success(c, NewSMSLoginResponse(result.User, result.IsNew))
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用 refresh token 刷新获取新的 token 对（token 轮转）
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} response.Result
// @Router /api/v1/user/refresh [post]
func (h *Handler) RefreshToken(ctx context.Context, c *app.RequestContext) error {
	ctx, span := tracer.Start(ctx, "handler.RefreshToken")
	defer span.End()

	// 从 cookie 获取 refresh token
	refreshTokenString := string(c.Cookie(jwt.RefreshTokenCookieKey))
	if refreshTokenString == "" {
		return response.Err(response.CodeUnauthorized, "未找到刷新令牌")
	}

	newTokenPair, err := h.userService.RefreshToken(ctx, refreshTokenString)
	if err != nil {
		tracer.RecordError(span, err)
		return err
	}

	// 设置新 token 到 cookie
	h.jwtManager.SetTokensInCookie(c, newTokenPair)

	return response.Success(c, nil)
}

// Logout 登出
// @Summary 登出
// @Description 登出并撤销所有 token
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} response.Result
// @Router /api/v1/user/logout [post]
func (h *Handler) Logout(ctx context.Context, c *app.RequestContext) error {
	ctx, span := tracer.Start(ctx, "handler.Logout")
	defer span.End()

	var accessJTI, refreshJTI string

	// 获取并解析 access token
	accessTokenString := string(c.Cookie(jwt.AccessTokenCookieKey))
	if accessTokenString != "" {
		if claims, err := h.jwtManager.ParseToken(accessTokenString, jwt.TokenTypeAccess); err == nil {
			accessJTI = claims.ID
		}
	}

	// 获取并解析 refresh token
	refreshTokenString := string(c.Cookie(jwt.RefreshTokenCookieKey))
	if refreshTokenString != "" {
		if claims, err := h.jwtManager.ParseToken(refreshTokenString, jwt.TokenTypeRefresh); err == nil {
			refreshJTI = claims.ID
		}
	}

	// 登出（将 token 加入黑名单）
	if err := h.userService.Logout(ctx, accessJTI, refreshJTI); err != nil {
		tracer.RecordError(span, err)
		// 记录错误但继续执行
	}

	// 清除 cookie 中的 token
	h.jwtManager.ClearTokensFromCookie(c)

	return response.Success(c, nil)
}
