package user

import (
	domain "arch3/internal/domain/user"
	"arch3/pkg/sqlx"
)

// toDomain 将 DAO 实体转换为领域模型
func toDomain(entity *Entity) *domain.User {
	return &domain.User{
		ID:           entity.ID,
		UserID:       entity.UserID,
		GroupID:      sqlx.NullStringToPtr(entity.GroupID),
		UserName:     entity.UserName,
		RealName:     sqlx.NullStringToPtr(entity.RealName),
		PasswordHash: entity.PasswordHash,
		Email:        sqlx.NullStringToPtr(entity.Email),
		PhoneNumber:  entity.PhoneNumber,
		AvatarURL:    sqlx.NullStringToPtr(entity.AvatarURL),
		Gender:       entity.Gender,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
		Status:       entity.Status,
		IDNumber:     sqlx.NullStringToPtr(entity.IDNumber),
		Source:       sqlx.NullStringToPtr(entity.Source),
		DeviceID:     sqlx.NullStringToPtr(entity.DeviceID),
	}
}

// toEntity 将领域模型转换为 DAO 实体
func toEntity(u *domain.User) *Entity {
	return &Entity{
		ID:           u.ID,
		UserID:       u.UserID,
		GroupID:      sqlx.PtrToNullString(u.GroupID),
		UserName:     u.UserName,
		RealName:     sqlx.PtrToNullString(u.RealName),
		PasswordHash: u.PasswordHash,
		Email:        sqlx.PtrToNullString(u.Email),
		PhoneNumber:  u.PhoneNumber,
		AvatarURL:    sqlx.PtrToNullString(u.AvatarURL),
		Gender:       u.Gender,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Status:       u.Status,
		IDNumber:     sqlx.PtrToNullString(u.IDNumber),
		Source:       sqlx.PtrToNullString(u.Source),
		DeviceID:     sqlx.PtrToNullString(u.DeviceID),
	}
}
