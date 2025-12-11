package user

import (
	"context"

	domain "arch3/internal/domain/user"
)

// Repository 用户仓储接口（由使用方定义）
type Repository interface {
	// FindByUserID 根据业务 ID 查询用户
	FindByUserID(ctx context.Context, userID string) (*domain.User, error)
	// FindByPhoneNumber 根据手机号查询用户
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*domain.User, error)
	// Create 创建用户
	Create(ctx context.Context, user *domain.User) error
	// Update 更新用户
	Update(ctx context.Context, user *domain.User) error
}
