package store

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

type MinioStore struct {
	client     *minio.Client
	bucketName string
}

func NewMinioStore(client *minio.Client, bucket string) *MinioStore {
	return &MinioStore{
		client:     client,
		bucketName: bucket,
	}
}

func (m *MinioStore) Upload(ctx context.Context, fileName string, content io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucketName, fileName, content, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return fileName, nil
}
