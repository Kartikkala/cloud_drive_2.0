package storage

import (
	"context"
	"io"
	"log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct{
	client *minio.Client
}

func NewMinioStorage(endpoint, accessKey, secretKey string) (*MinioStorage, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
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
