package storage

import (
	"gorm.io/gorm"
	"context"
	"io"
	"time"
)
type ObjectInfo struct {
	Bytes uint64
	ContentType string
	LastModified time.Time
	Key string
	Bucket string
	Metadata map[string] string
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
