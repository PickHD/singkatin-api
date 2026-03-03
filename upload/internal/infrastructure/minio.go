package infrastructure

import (
	"context"

	"singkatin-api/upload/internal/config"
	"singkatin-api/upload/pkg/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConnectionProvider struct {
	client *minio.Client
}

func NewMinIOConnection(ctx context.Context, cfg *config.Config) *MinIOConnectionProvider {
	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		logger.Errorf("Failed to connect to MinIO: %v", err)
	}

	return &MinIOConnectionProvider{client: minioClient}
}

func (m *MinIOConnectionProvider) GetClient() *minio.Client {
	return m.client
}
