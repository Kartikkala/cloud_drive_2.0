package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"gorm.io/gorm"
)

type ObjectStorage interface {
	GeneratePostUploadPolicy(ctx context.Context, bucket, dirKey string, expiry time.Time) (*url.URL, map[string]string, error)
	Put(ctx context.Context, bucket, key string, data io.Reader, size int64) error
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	Copy(ctx context.Context, bucket, srcKey, destKey string) error
}

type Service struct {
	DB                *gorm.DB
	Client            ObjectStorage
	NewJobEventBroker *events.Broker[*events.Job]
}

type NodeWithPermission struct {
	Node
	PermissionType *PermissionType
}

type MinioStorage struct {
	client *minio.Client
}

type UploadPolicy struct {
	URL       string
	Fields    map[string]string
	KeyPrefix string
}
