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

	UploadRepositoryImpl struct {
		Context context.Context
		Config  *config.Configuration
		Tracer  *trace.TracerProvider
		MinIO   *minio.Client
	}
)

// NewUploadRepository return new instances upload repository
func NewUploadRepository(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, minioCli *minio.Client) *UploadRepositoryImpl {
	return &UploadRepositoryImpl{
		Context: ctx,
		Config:  config,
		Tracer:  tracer,
		MinIO:   minioCli,
	}
}

func (ur *UploadRepositoryImpl) UploadObject(ctx context.Context, req *model.UploadAvatarRequest) error {
	tr := ur.Tracer.Tracer("Upload-UploadObject Repository")
	ctx, span := tr.Start(ctx, "Start UploadObject")
	defer span.End()

	b := bytes.NewBuffer(req.Avatars)

	_, err := ur.MinIO.PutObject(ctx, ur.Config.MinIO.Bucket, req.FileName, b,
		int64(b.Len()), minio.PutObjectOptions{ContentType: req.ContentType})
	if err != nil {
		logger.Errorf("UploadRepositoryImpl.UploadObject PutObject ERROR, %v", err)

		return err
	}

	return nil
}
