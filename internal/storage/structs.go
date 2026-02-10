package storage

import (
	"context"
	"io"

	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"gorm.io/gorm"
)

type ObjectStorage interface {
	Put(ctx context.Context, bucket, key string, data io.Reader, size int64) error
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	Copy(ctx context.Context, bucket, srcKey, destKey string) error
}

type Service struct {
	DB          *gorm.DB
	Client      ObjectStorage
	EventBroker *events.Broker[*events.Job]
}

type NodeWithPermission struct {
	Node
	PermissionType *PermissionType
}
