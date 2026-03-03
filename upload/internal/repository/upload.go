package repository

import (
	"bytes"
	"context"

	"singkatin-api/upload/internal/config"
	"singkatin-api/upload/internal/model"
	"singkatin-api/upload/pkg/logger"

	"github.com/minio/minio-go/v7"

	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UploadRepository is an interface that has all the function to be implemented inside upload repository
	UploadRepository interface {
		UploadObject(ctx context.Context, req *model.UploadAvatarRequest) error
	}

	uploadRepositoryImpl struct {
		Config *config.Config
		Tracer *trace.TracerProvider
		MinIO  *minio.Client
	}
)

// NewUploadRepository return new instances upload repository
func NewUploadRepository(config *config.Config, tracer *trace.TracerProvider, minioCli *minio.Client) UploadRepository {
	return &uploadRepositoryImpl{
		Config: config,
		Tracer: tracer,
		MinIO:  minioCli,
	}
}

func (r *uploadRepositoryImpl) UploadObject(ctx context.Context, req *model.UploadAvatarRequest) error {
	tr := r.Tracer.Tracer("Upload-UploadObject Repository")
	_, span := tr.Start(ctx, "Start UploadObject")
	defer span.End()

	b := bytes.NewBuffer(req.Avatars)

	_, err := r.MinIO.PutObject(ctx, r.Config.MinIO.Bucket, req.FileName, b,
		int64(b.Len()), minio.PutObjectOptions{ContentType: req.ContentType})
	if err != nil {
		logger.Errorf("UploadRepositoryImpl.UploadObject PutObject ERROR, %v", err)

		return err
	}

	return nil
}
