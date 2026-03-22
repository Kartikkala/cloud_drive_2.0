package shared

import (
	"context"
	"io"
	"net/url"
	"time"
)

type ObjectStorage interface {
	GeneratePostUploadPolicy(ctx context.Context, bucket, dirKey string, expiry time.Time) (*url.URL, map[string]string, error)
	GeneratePresignedGetURL(ctx context.Context, bucket, key string) (*url.URL, error)
	Put(ctx context.Context, bucket, key string, data io.Reader, size int64) error
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	Copy(ctx context.Context, bucket, srcKey, destKey string) error
}
