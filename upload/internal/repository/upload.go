package repository

import (
	"bytes"
	"context"

	"github.com/PickHD/singkatin-revamp/upload/internal/config"
	"github.com/PickHD/singkatin-revamp/upload/internal/model"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UploadRepository is an interface that has all the function to be implemented inside upload repository
	UploadRepository interface {
		UploadObject(ctx context.Context, req *model.UploadAvatarRequest) error
	}

	// UploadRepositoryImpl is an app upload struct that consists of all the dependencies needed for upload repository
	UploadRepositoryImpl struct {
		Context context.Context
		Config  *config.Configuration
		Logger  *logrus.Logger
		Tracer  *trace.TracerProvider
		MinIO   *minio.Client
	}
)

// NewUploadRepository return new instances upload repository
func NewUploadRepository(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, minioCli *minio.Client) *UploadRepositoryImpl {
	return &UploadRepositoryImpl{
		Context: ctx,
		Config:  config,
		Logger:  logger,
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
		ur.Logger.Error("UploadRepositoryImpl.UploadObject PutObject ERROR,", err)

		return err
	}

	return nil
}
