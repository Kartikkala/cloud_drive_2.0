package storage

import (
	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService(DB *gorm.DB) *Service {
	return &Service{
		DB : DB,
	}
}