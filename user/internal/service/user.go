package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/PickHD/singkatin-revamp/user/internal/config"
	"github.com/PickHD/singkatin-revamp/user/internal/helper"
	"github.com/PickHD/singkatin-revamp/user/internal/model"
	"github.com/PickHD/singkatin-revamp/user/internal/repository"
	shortenerpb "github.com/PickHD/singkatin-revamp/user/pkg/api/v1/proto/shortener"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UserService is an interface that has all the function to be implemented inside user service
	UserService interface {
		GetUserDetail(email string) (*model.User, error)
		GetUserShorts(userID string) ([]model.UserShorts, error)
		GenerateUserShorts(userID string, req *model.ShortUserRequest) (*model.ShortUserResponse, error)
		UpdateUserProfile(userID string, req *model.EditProfileRequest) error
		UploadUserAvatar(ctx *fiber.Ctx, userID string) (*model.UploadAvatarResponse, error)
		UpdateUserShorts(shortID string, req *model.ShortUserRequest) (*model.ShortUserResponse, error)
		DeleteUserShorts(shortID string) (*model.ShortUserResponse, error)
	}

	// UserServiceImpl is an app user struct that consists of all the dependencies needed for user service
	UserServiceImpl struct {
		Context      context.Context
		Config       *config.Configuration
		Logger       *logrus.Logger
		Tracer       *trace.TracerProvider
		UserRepo     repository.UserRepository
		ShortClients shortenerpb.ShortenerServiceClient
	}
)

// NewUserService return new instances user service
func NewUserService(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, userRepo repository.UserRepository, shortClients shortenerpb.ShortenerServiceClient) *UserServiceImpl {
	return &UserServiceImpl{
		Context:      ctx,
		Config:       config,
		Logger:       logger,
		Tracer:       tracer,
		UserRepo:     userRepo,
		ShortClients: shortClients,
	}
}

func (us *UserServiceImpl) GetUserDetail(email string) (*model.User, error) {
	tr := us.Tracer.Tracer("User-GetUserDetail Service")
	_, span := tr.Start(us.Context, "Start GetUserDetail")
	defer span.End()

	return us.UserRepo.FindByEmail(us.Context, email)
}

func (us *UserServiceImpl) GetUserShorts(userID string) ([]model.UserShorts, error) {
	tr := us.Tracer.Tracer("User-GetUserShorts Service")
	_, span := tr.Start(us.Context, "Start GetUserShorts")
	defer span.End()

	data, err := us.ShortClients.GetListShortenerByUserID(us.Context, &shortenerpb.ListShortenerRequest{
		UserId: userID})
	if err != nil {
		us.Logger.Error("UserServiceImpl.GetUserShorts ShortClients ERROR, ", err)
		return nil, err
	}

	if len(data.Shorteners) < 1 {
		return nil, nil
	}

	shorteners := make([]model.UserShorts, len(data.Shorteners))

	for i, q := range data.Shorteners {
		shorteners[i] = model.UserShorts{
			ID:       q.GetId(),
			FullURL:  q.GetFullUrl(),
			ShortURL: q.GetShortUrl(),
			Visited:  q.GetVisited(),
		}
	}

	return shorteners, nil
}

func (us *UserServiceImpl) GenerateUserShorts(userID string, req *model.ShortUserRequest) (*model.ShortUserResponse, error) {
	tr := us.Tracer.Tracer("User-GenerateUserShorts Service")
	_, span := tr.Start(us.Context, "Start GenerateUserShorts")
	defer span.End()

	msg := model.GenerateShortUserMessage{
		FullURL:  req.FullURL,
		UserID:   userID,
		ShortURL: helper.RandomStringBytesMaskImprSrcSB(8),
	}

	err := us.UserRepo.PublishCreateUserShortener(us.Context, &msg)
	if err != nil {
		return nil, err
	}

	return &model.ShortUserResponse{
		ShortURL: fmt.Sprintf("%s/%s", us.Config.HttpService.ShortenerBaseAPIURL, msg.ShortURL),
		Method:   "GET",
	}, nil
}

func (us *UserServiceImpl) UpdateUserProfile(userID string, req *model.EditProfileRequest) error {
	tr := us.Tracer.Tracer("User-UpdateUserProfile Service")
	_, span := tr.Start(us.Context, "Start UpdateUserProfile")
	defer span.End()

	if req.FullName == "" {
		return model.NewError(model.Validation, "Full Name Required")
	}

	if len(req.FullName) < 3 {
		return model.NewError(model.Validation, "Full Name must more than 3")
	}

	err := us.UserRepo.UpdateProfileByID(us.Context, userID, req)
	if err != nil {
		return err
	}

	return nil
}

func (us *UserServiceImpl) UploadUserAvatar(ctx *fiber.Ctx, userID string) (*model.UploadAvatarResponse, error) {
	tr := us.Tracer.Tracer("User-UploadUserAvatar Service")
	_, span := tr.Start(us.Context, "Start UploadUserAvatar")
	defer span.End()

	file, err := ctx.FormFile("file")
	if err != nil {
		return nil, err
	}

	buffer, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer buffer.Close()

	contentType := file.Header["Content-Type"][0]
	fileName := fmt.Sprintf("%s-%s", file.Filename, userID)

	// detect content type & validate only allow images
	switch contentType {
	case "image/jpeg", "image/png":
	default:
		return nil, model.NewError(model.Validation, "invalid file, only accept file with extension image/jpeg or image/png")
	}

	// copy to new buffer
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, buffer); err != nil {
		return nil, err
	}

	// publish to queue async
	err = us.UserRepo.PublishUploadAvatarUser(us.Context, &model.UploadAvatarRequest{
		FileName:    fileName,
		ContentType: contentType,
		Avatars:     buf.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	fileUrl := url.URL{
		Scheme: "http",
		Path:   fmt.Sprintf("%s/%s/%s", us.Config.MinIO.Endpoint, us.Config.MinIO.Bucket, fileName),
	}

	// update avatar_url users to db
	err = us.UserRepo.UpdateAvatarUserByID(us.Context, fileUrl.String(), userID)
	if err != nil {
		return nil, err
	}

	return &model.UploadAvatarResponse{
		FileURL: fileUrl.String(),
	}, nil
}

func (us *UserServiceImpl) UpdateUserShorts(shortID string, req *model.ShortUserRequest) (*model.ShortUserResponse, error) {
	tr := us.Tracer.Tracer("User-UpdateUserShorts Service")
	_, span := tr.Start(us.Context, "Start UpdateUserShorts")
	defer span.End()

	err := us.UserRepo.PublishUpdateUserShortener(us.Context, shortID, req)
	if err != nil {
		return nil, err
	}

	return &model.ShortUserResponse{}, nil
}

func (us *UserServiceImpl) DeleteUserShorts(shortID string) (*model.ShortUserResponse, error) {
	tr := us.Tracer.Tracer("User-DeleteUserShorts Service")
	_, span := tr.Start(us.Context, "Start DeleteUserShorts")
	defer span.End()

	err := us.UserRepo.PublishDeleteUserShortener(us.Context, shortID)
	if err != nil {
		return nil, err
	}

	return &model.ShortUserResponse{}, nil
}
