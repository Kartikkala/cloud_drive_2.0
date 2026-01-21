package auth

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"unique;not null" json:"email"`
	Username     string    `gorm:"unique;not null" json:"username"`
	Password     string    `gorm:"not null" json:"-"`
	CreationDate time.Time `gorm:"column:creation_date;autoCreateTime" json:"creation_date"`
	Role         string    `gorm:"default:user" json:"role"`
}