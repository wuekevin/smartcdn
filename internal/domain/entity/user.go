package entity

import "time"

type User struct {
	ID               uint64     `gorm:"primarykey;autoIncrement"`
	Username         string     `gorm:"uniqueIndex;size:64;not null"`
	Password         string     `gorm:"not null"`
	Email            string     `gorm:"size:255"`
	IsActive         bool       `gorm:"not null;default:true"`
	LoginFailedCount int        `gorm:"not null;default:0"`
	LockedUntil      *time.Time `gorm:"column:locked_until"`
	LastLoginAt      *time.Time `gorm:"column:last_login_at"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `gorm:"index"`
}

type Role struct {
	ID          uint64     `gorm:"primarykey;autoIncrement"`
	Name        string     `gorm:"uniqueIndex;size:64;not null"`
	Description string     `gorm:"size:255"`
	IsSystem    bool       `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `gorm:"index"`
}

type Permission struct {
	RoleID   uint64 `gorm:"primarykey;column:role_id"`
	Resource string `gorm:"primarykey;size:128"`
	Action   string `gorm:"primarykey;size:64"`
}
