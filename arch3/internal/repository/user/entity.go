package user

import (
	"database/sql"
	"time"
)

// Entity 用户数据库实体
type Entity struct {
	ID           uint           `gorm:"column:id;primaryKey;autoIncrement"`
	UserID       string         `gorm:"column:user_id;type:varchar(32);uniqueIndex;not null"`
	GroupID      sql.NullString `gorm:"column:group_id;type:varchar(32)"`
	UserName     string         `gorm:"column:user_name;type:varchar(50);not null;index"`
	RealName     sql.NullString `gorm:"column:real_name;type:varchar(100);index"`
	PasswordHash string         `gorm:"column:password_hash;type:char(64);not null"`
	Email        sql.NullString `gorm:"column:email;type:varchar(254)"`
	PhoneNumber  string         `gorm:"column:phone_number;type:varchar(15);uniqueIndex;not null"`
	AvatarURL    sql.NullString `gorm:"column:avatar_url;type:varchar(255)"`
	Gender       string         `gorm:"column:gender;type:enum('male','female','other');default:other"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	Status       string         `gorm:"column:status;type:enum('real_name_verified','real_name_unverified','banned','under_review');default:real_name_unverified;not null"`
	IDNumber     sql.NullString `gorm:"column:id_number;type:varchar(30);uniqueIndex"`
	Source       sql.NullString `gorm:"column:source;type:varchar(50);index"`
	DeviceID     sql.NullString `gorm:"column:device_id;type:varchar(128)"`
}

// TableName 返回表名
func (Entity) TableName() string {
	return "users"
}
