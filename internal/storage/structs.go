package storage

import (
	"gorm.io/gorm"
	"context"
	"io"
	"time"
)
type NodeInfo struct {
	Bytes uint64
	ContentType string
	LastModified time.Time
	Key string
	OwnerID uint64
	ParentID string
}

type ObjectStorage interface {
	Put(ctx context.Context, bucket, key string, data io.Reader, size int64) error
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
}

type Service struct {
	DB *gorm.DB
	Client ObjectStorage
}
