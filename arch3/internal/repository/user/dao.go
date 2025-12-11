package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// ErrNotFound 记录不存在错误
var ErrNotFound = errors.New("record not found")

// DAO 用户数据访问对象
type DAO struct {
	db *gorm.DB
}

// NewDAO 创建用户 DAO
func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

// FindByUserID 根据业务 ID 查询用户
func (d *DAO) FindByUserID(ctx context.Context, userID string) (*Entity, error) {
	var entity Entity
	err := d.db.WithContext(ctx).Where("user_id = ?", userID).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// FindByPhoneNumber 根据手机号查询用户
func (d *DAO) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*Entity, error) {
	var entity Entity
	err := d.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// Create 创建用户
func (d *DAO) Create(ctx context.Context, entity *Entity) error {
	return d.db.WithContext(ctx).Create(entity).Error
}

// Update 更新用户
func (d *DAO) Update(ctx context.Context, entity *Entity) error {
	return d.db.WithContext(ctx).Save(entity).Error
}
