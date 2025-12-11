package user

import (
	"context"
	"errors"
	"time"

	domain "arch3/internal/domain/user"
	"arch3/pkg/jwt"
	"arch3/pkg/response"
	"arch3/pkg/tracer"
	"arch3/pkg/ulid"
)

// SMSLogin 短信验证码登录（用户不存在则自动注册）
func (s *service) SMSLogin(ctx context.Context, phoneNumber, smsCode string) (*domain.LoginResult, error) {
	ctx, span := tracer.Start(ctx, "service.user.SMSLogin")
	defer span.End()

	// 验证短信验证码
	if err := s.smsClient.Verify(ctx, SMSTypeLogin, phoneNumber, smsCode); err != nil {
		tracer.RecordError(span, err)
		return nil, SMSToResponse(err)
	}

	// 查询或创建用户
	isNew := false
	u, err := s.userRepo.FindByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// 用户不存在，自动注册
			u, err = s.registerUserByPhone(ctx, phoneNumber)
			if err != nil {
				tracer.RecordError(span, err)
				return nil, response.Err(response.CodeDatabaseError, "注册失败")
			}
			isNew = true
		} else {
			tracer.RecordError(span, err)
			return nil, response.Err(response.CodeDatabaseError, "查询用户失败")
		}
	}

	// 生成 token 对
	tokenPair, err := s.jwtManager.GenerateTokenPair(u.UserID)
	if err != nil {
		tracer.RecordError(span, err)
		return nil, response.Err(response.CodeInternal, "生成令牌失败")
	}

	return &domain.LoginResult{
		User:      u,
		TokenPair: tokenPair,
		IsNew:     isNew,
	}, nil
}

// registerUserByPhone 通过手机号注册新用户
func (s *service) registerUserByPhone(ctx context.Context, phoneNumber string) (*domain.User, error) {
	now := time.Now().UTC()

	// 生成用户 ID
	userID, err := ulid.New()
	if err != nil {
		return nil, err
	}

	// 生成默认用户名：用户_手机号后四位
	defaultUserName := "用户_" + phoneNumber[len(phoneNumber)-4:]

	u := &domain.User{
		UserID:      userID,
		UserName:    defaultUserName,
		PhoneNumber: phoneNumber,
		Gender:      "other",
		Status:      "real_name_unverified",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

// RefreshToken 刷新 token
func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	ctx, span := tracer.Start(ctx, "service.user.RefreshToken")
	defer span.End()

	// 解析 refresh token
	claims, err := s.jwtManager.ParseToken(refreshToken, jwt.TokenTypeRefresh)
	if err != nil {
		tracer.RecordError(span, err)
		return nil, response.Err(response.CodeTokenInvalid, "刷新令牌无效")
	}

	// 检查 token 是否在黑名单中
	blacklisted, err := s.jwtManager.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		tracer.RecordError(span, err)
		return nil, response.Err(response.CodeCacheError, "检查令牌状态失败")
	}
	if blacklisted {
		return nil, response.Err(response.CodeTokenInvalid, "刷新令牌已失效")
	}

	// 验证用户是否存在
	_, err = s.userRepo.FindByUserID(ctx, claims.UserID)
	if err != nil {
		tracer.RecordError(span, err)
		return nil, response.Err(response.CodeUserNotFound, "用户不存在")
	}

	// 先将旧的 refresh token 加入黑名单（token 轮转）
	// 必须在生成新 token 之前完成，防止旧 token 继续使用
	if err := s.jwtManager.RevokeRefreshToken(ctx, claims.ID); err != nil {
		tracer.RecordError(span, err)
		return nil, response.Err(response.CodeCacheError, "令牌轮转失败")
	}

	// 生成新的 token 对
	newTokenPair, err := s.jwtManager.GenerateTokenPair(claims.UserID)
	if err != nil {
		tracer.RecordError(span, err)
		return nil, response.Err(response.CodeInternal, "生成令牌失败")
	}

	return newTokenPair, nil
}

// Logout 登出
func (s *service) Logout(ctx context.Context, accessJTI, refreshJTI string) error {
	ctx, span := tracer.Start(ctx, "service.user.Logout")
	defer span.End()

	// 将 access token 加入黑名单
	if accessJTI != "" {
		if err := s.jwtManager.AddTokenToBlacklist(ctx, accessJTI, s.jwtManager.GetAccessExpire()); err != nil {
			tracer.RecordError(span, err)
			// 记录错误但不返回
		}
	}

	// 将 refresh token 加入黑名单
	if refreshJTI != "" {
		if err := s.jwtManager.RevokeRefreshToken(ctx, refreshJTI); err != nil {
			tracer.RecordError(span, err)
			// 记录错误但不返回
		}
	}

	return nil
}

// GetUserByID 根据 ID 获取用户
func (s *service) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "service.user.GetUserByID")
	defer span.End()

	u, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		tracer.RecordError(span, err)
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, response.Err(response.CodeUserNotFound, "用户不存在")
		}
		return nil, response.Err(response.CodeDatabaseError, "查询用户失败")
	}

	return u, nil
}
