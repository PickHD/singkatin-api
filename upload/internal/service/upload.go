package service

import (
	"context"

	"singkatin-api/upload/internal/config"
	"singkatin-api/upload/internal/dto/request"
	"singkatin-api/upload/internal/repository"

	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UploadService is an interface that has all the function to be implemented inside upload service
	UploadService interface {
		UploadAvatarUser(ctx context.Context, req *request.UploadAvatarRequest) error
	}

	uploadServiceImpl struct {
		Config     *config.Config
		Tracer     *trace.TracerProvider
		UploadRepo repository.UploadRepository
	}
)

// NewUploadService return new instances upload repository
func NewUploadService(config *config.Config, tracer *trace.TracerProvider, uploadRepo repository.UploadRepository) UploadService {
	return &uploadServiceImpl{
		Config:     config,
		Tracer:     tracer,
		UploadRepo: uploadRepo,
	}
}

func (s *uploadServiceImpl) UploadAvatarUser(ctx context.Context, req *request.UploadAvatarRequest) error {
	tr := s.Tracer.Tracer("Upload-UploadAvatarUser Service")
	_, span := tr.Start(ctx, "Start UploadObject")
	defer span.End()

	return s.UploadRepo.UploadObject(ctx, req)
}
