package controller

import (
	"context"

	"github.com/PickHD/singkatin-revamp/upload/internal/config"
	"github.com/PickHD/singkatin-revamp/upload/internal/model"
	"github.com/PickHD/singkatin-revamp/upload/internal/service"
	uploadpb "github.com/PickHD/singkatin-revamp/upload/pkg/api/v1/proto/upload"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UploadController is an interface that has all the function to be implemented inside upload controller
	UploadController interface {
		ProcessUploadAvatarUser(ctx context.Context, msg *uploadpb.UploadAvatarMessage) error
	}

	// UploadControllerImpl is an app upload struct that consists of all the dependencies needed for upload controller
	UploadControllerImpl struct {
		Context   context.Context
		Config    *config.Configuration
		Logger    *logrus.Logger
		Tracer    *trace.TracerProvider
		UploadSvc service.UploadService
	}
)

// NewUploadController return new instances upload controller
func NewUploadController(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, uploadSvc service.UploadService) *UploadControllerImpl {
	return &UploadControllerImpl{
		Context:   ctx,
		Config:    config,
		Logger:    logger,
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
