package storage

import (
	"context"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioStorage(endpoint, accessKey, secretKey string, useSSL bool) (*MinioStorage, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		return nil, err
	}

	return &MinioStorage{client: minioClient}, nil
}

func (m *MinioStorage) Put(
	ctx context.Context,
	bucket, key string,
	data io.Reader,
	size int64,
) error {

	_, err := m.client.PutObject(
		ctx,
		bucket,
		key,
		data,
		size,
		minio.PutObjectOptions{},
	)
	return err
}

func (m *MinioStorage) Get(
	ctx context.Context,
	bucket, key string,
) (io.ReadCloser, error) {
	stream, err := m.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})

	if err != nil {
		log.Printf("Error encountered in Minio GET():\n %v", err)
		return nil, err
	}

	// Validate if object actually exists in MINIO
	if _, err = stream.Stat(); err != nil {
		stream.Close()
		return nil, err
	}

	return stream, nil
}

func (m *MinioStorage) Delete(
	ctx context.Context,
	bucket, key string,
) error {

	err := m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("Error encountered in Minio REMOVE(): %v\n", err)
	}
	return err
}

func (m *MinioStorage) Copy(
	ctx context.Context,
	bucket, srcKey, destKey string,
) error {
	var src minio.CopySrcOptions = minio.CopySrcOptions{
		Bucket: bucket,
		Object: srcKey,
	}

	var dest minio.CopyDestOptions = minio.CopyDestOptions{
		Bucket: bucket,
		Object: destKey,
	}
	_, err := m.client.CopyObject(ctx, dest, src)

	if err != nil {
		return err
	}
	return nil
}

func (m *MinioStorage) GeneratePresignedGetURL(
	ctx context.Context,
	bucket, key string,
) (*url.URL, error) {
	return m.client.PresignedGetObject(
		ctx,
		bucket,
		key,
		time.Hour * 2,
		url.Values{},
	)
}

func (m *MinioStorage) GeneratePostUploadPolicy(
	ctx context.Context,
	bucket, dirKey string,
	expiry time.Time,
) (*url.URL, map[string]string, error) {

	policy := minio.NewPostPolicy()

	policy.SetBucket(bucket)
	policy.SetKeyStartsWith(dirKey + "/")
	policy.SetExpires(expiry)

	url, formData, err := m.client.PresignedPostPolicy(ctx, policy)

	if err != nil {
		return nil, nil, err
	}

	return url, formData, nil
}
