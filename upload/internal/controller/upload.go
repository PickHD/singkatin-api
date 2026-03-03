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

	uploadControllerImpl struct {
		Config    *config.Config
		Tracer    *trace.TracerProvider
		UploadSvc service.UploadService
	}
)

// NewUploadController return new instances upload controller
func NewUploadController(config *config.Config, tracer *trace.TracerProvider, uploadSvc service.UploadService) UploadController {
	return &uploadControllerImpl{
		Config:    config,
		Tracer:    tracer,
		UploadSvc: uploadSvc,
	}
}

func (c *uploadControllerImpl) ProcessUploadAvatarUser(ctx context.Context, msg *uploadpb.UploadAvatarMessage) error {
	tr := c.Tracer.Tracer("Upload-ProcessUploadAvatarUser Controller")
	_, span := tr.Start(ctx, "Start ProcessUploadAvatarUser")
	defer span.End()

	req := &model.UploadAvatarRequest{
		FileName:    msg.GetFileName(),
		ContentType: msg.GetContentType(),
		Avatars:     msg.GetAvatars(),
	}

	err := c.UploadSvc.UploadAvatarUser(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
