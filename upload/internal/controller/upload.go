package controller

import (
	"context"

	"singkatin-api/upload/internal/config"
	"singkatin-api/upload/internal/model"
	"singkatin-api/upload/internal/service"
	uploadpb "singkatin-api/upload/pkg/api/v1/proto/upload"

	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UploadController is an interface that has all the function to be implemented inside upload controller
	UploadController interface {
		ProcessUploadAvatarUser(ctx context.Context, msg *uploadpb.UploadAvatarMessage) error
	}

	UploadControllerImpl struct {
		Context   context.Context
		Config    *config.Configuration
		Tracer    *trace.TracerProvider
		UploadSvc service.UploadService
	}
)

// NewUploadController return new instances upload controller
func NewUploadController(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, uploadSvc service.UploadService) *UploadControllerImpl {
	return &UploadControllerImpl{
		Context:   ctx,
		Config:    config,
		Tracer:    tracer,
		UploadSvc: uploadSvc,
	}
}

func (uc *UploadControllerImpl) ProcessUploadAvatarUser(ctx context.Context, msg *uploadpb.UploadAvatarMessage) error {
	tr := uc.Tracer.Tracer("Upload-ProcessUploadAvatarUser Controller")
	_, span := tr.Start(uc.Context, "Start ProcessUploadAvatarUser")
	defer span.End()

	req := &model.UploadAvatarRequest{
		FileName:    msg.GetFileName(),
		ContentType: msg.GetContentType(),
		Avatars:     msg.GetAvatars(),
	}

	err := uc.UploadSvc.UploadAvatarUser(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
