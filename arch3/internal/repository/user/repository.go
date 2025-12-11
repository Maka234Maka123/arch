package user

import (
	"context"
	"errors"

	domain "arch3/internal/domain/user"
	userservice "arch3/internal/service/user"
)

// Repository 用户仓储实现
type Repository struct {
	dao *DAO
}

// NewRepository 创建用户仓储实例
func NewRepository(dao *DAO) userservice.Repository {
	return &Repository{dao: dao}
}

// FindByUserID 根据业务 ID 查询用户
func (r *Repository) FindByUserID(ctx context.Context, userID string) (*domain.User, error) {
	entity, err := r.dao.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomain(entity), nil
}

// FindByPhoneNumber 根据手机号查询用户
func (r *Repository) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*domain.User, error) {
	entity, err := r.dao.FindByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomain(entity), nil
}

// Create 创建用户
func (r *Repository) Create(ctx context.Context, u *domain.User) error {
	entity := toEntity(u)
	if err := r.dao.Create(ctx, entity); err != nil {
		return err
	}
	// 回填生成的字段
	u.CreatedAt = entity.CreatedAt
	u.UpdatedAt = entity.UpdatedAt
	return nil
}

// Update 更新用户
func (r *Repository) Update(ctx context.Context, u *domain.User) error {
	entity := toEntity(u)
	if err := r.dao.Update(ctx, entity); err != nil {
		return err
	}
	u.UpdatedAt = entity.UpdatedAt
	return nil
}
