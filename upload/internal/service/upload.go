package service

import (
	"context"

	"github.com/PickHD/singkatin-revamp/upload/internal/config"
	"github.com/PickHD/singkatin-revamp/upload/internal/model"
	"github.com/PickHD/singkatin-revamp/upload/internal/repository"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UploadService is an interface that has all the function to be implemented inside upload service
	UploadService interface {
		UploadAvatarUser(ctx context.Context, req *model.UploadAvatarRequest) error
	}

	// UploadServiceImpl is an app upload struct that consists of all the dependencies needed for upload service
	UploadServiceImpl struct {
		Context    context.Context
		Config     *config.Configuration
		Logger     *logrus.Logger
		Tracer     *trace.TracerProvider
		UploadRepo repository.UploadRepository
	}
)

// NewUploadService return new instances upload repository
func NewUploadService(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, uploadRepo repository.UploadRepository) *UploadServiceImpl {
	return &UploadServiceImpl{
		Context:    ctx,
		Config:     config,
		Logger:     logger,
		Tracer:     tracer,
		UploadRepo: uploadRepo,
	}
}

func (us *UploadServiceImpl) UploadAvatarUser(ctx context.Context, req *model.UploadAvatarRequest) error {
	tr := us.Tracer.Tracer("Upload-UploadAvatarUser Service")
	ctx, span := tr.Start(ctx, "Start UploadObject")
	defer span.End()

	return us.UploadRepo.UploadObject(ctx, req)
}
