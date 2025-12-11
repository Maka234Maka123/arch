package user

import (
	"errors"
	"time"
)

// ErrUserNotFound 用户不存在错误
var ErrUserNotFound = errors.New("user not found")

// User 用户领域模型
type User struct {
	ID           uint
	UserID       string
	GroupID      *string
	UserName     string
	RealName     *string
	PasswordHash string
	Email        *string
	PhoneNumber  string
	AvatarURL    *string
	Gender       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Status       string
	IDNumber     *string
	Source       *string
	DeviceID     *string
}
