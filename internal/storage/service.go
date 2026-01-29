package storage

import (
	"gorm.io/gorm"
)

func NewService(DB *gorm.DB, storageClient ObjectStorage) *Service {
	return &Service{
		DB : DB,
		Client: storageClient,
	}
}