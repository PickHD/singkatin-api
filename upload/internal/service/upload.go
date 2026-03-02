package service

import (
	"context"

	"singkatin-api/upload/internal/config"
	"singkatin-api/upload/internal/model"
	"singkatin-api/upload/internal/repository"

	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UploadService is an interface that has all the function to be implemented inside upload service
	UploadService interface {
		UploadAvatarUser(ctx context.Context, req *model.UploadAvatarRequest) error
	}

	UploadServiceImpl struct {
		Context    context.Context
		Config     *config.Configuration
		Tracer     *trace.TracerProvider
		UploadRepo repository.UploadRepository
	}
)

// NewUploadService return new instances upload repository
func NewUploadService(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, uploadRepo repository.UploadRepository) *UploadServiceImpl {
	return &UploadServiceImpl{
		Context:    ctx,
		Config:     config,
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
