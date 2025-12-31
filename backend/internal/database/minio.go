package database

import (
	"context"
	"fmt"

	"github.com/Chelaran/mayoku/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ConnectMinIO создает подключение к MinIO
func ConnectMinIO(cfg *config.Config) (*minio.Client, error) {
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKeyID, cfg.MinIO.SecretAccessKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Проверка подключения и создание bucket если не существует
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIO.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.MinIO.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return client, nil
}
