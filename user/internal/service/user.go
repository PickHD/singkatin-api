package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"

	"singkatin-api/user/pkg/logger"
	"singkatin-api/user/pkg/utils"

	"singkatin-api/user/internal/config"
	"singkatin-api/user/internal/model"
	"singkatin-api/user/internal/repository"
	shortenerpb "singkatin-api/user/pkg/api/v1/proto/shortener"

	"github.com/gofiber/fiber/v2"
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

	// userServiceImpl is an app user struct that consists of all the dependencies needed for user service
	userServiceImpl struct {
		Context      context.Context
		Config       *config.Config
		Tracer       *trace.TracerProvider
		UserRepo     repository.UserRepository
		ShortClients shortenerpb.ShortenerServiceClient
	}
)

// NewUserService return new instances user service
func NewUserService(ctx context.Context, config *config.Config, tracer *trace.TracerProvider, userRepo repository.UserRepository, shortClients shortenerpb.ShortenerServiceClient) UserService {
	return &userServiceImpl{
		Context:      ctx,
		Config:       config,
		Tracer:       tracer,
		UserRepo:     userRepo,
		ShortClients: shortClients,
	}
}

func (s *userServiceImpl) GetUserDetail(email string) (*model.User, error) {
	tr := s.Tracer.Tracer("User-GetUserDetail Service")
	_, span := tr.Start(s.Context, "Start GetUserDetail")
	defer span.End()

	return s.UserRepo.FindByEmail(s.Context, email)
}

func (s *userServiceImpl) GetUserShorts(userID string) ([]model.UserShorts, error) {
	tr := s.Tracer.Tracer("User-GetUserShorts Service")
	_, span := tr.Start(s.Context, "Start GetUserShorts")
	defer span.End()

	data, err := s.ShortClients.GetListShortenerByUserID(s.Context, &shortenerpb.ListShortenerRequest{
		UserId: userID})
	if err != nil {
		logger.Errorf("UserServiceImpl.GetUserShorts ShortClients ERROR, %v", err)
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

func (s *userServiceImpl) GenerateUserShorts(userID string, req *model.ShortUserRequest) (*model.ShortUserResponse, error) {
	tr := s.Tracer.Tracer("User-GenerateUserShorts Service")
	_, span := tr.Start(s.Context, "Start GenerateUserShorts")
	defer span.End()

	msg := model.GenerateShortUserMessage{
		FullURL:  req.FullURL,
		UserID:   userID,
		ShortURL: utils.RandomStringBytesMaskImprSrcSB(8),
	}

	err := s.UserRepo.PublishCreateUserShortener(s.Context, &msg)
	if err != nil {
		return nil, err
	}

	return &model.ShortUserResponse{
		ShortURL: fmt.Sprintf("%s/%s", s.Config.HttpService.ShortenerBaseAPIURL, msg.ShortURL),
		Method:   "GET",
	}, nil
}

func (s *userServiceImpl) UpdateUserProfile(userID string, req *model.EditProfileRequest) error {
	tr := s.Tracer.Tracer("User-UpdateUserProfile Service")
	_, span := tr.Start(s.Context, "Start UpdateUserProfile")
	defer span.End()

	if req.FullName == "" {
		return model.NewError(model.Validation, "Full Name Required")
	}

	if len(req.FullName) < 3 {
		return model.NewError(model.Validation, "Full Name must more than 3")
	}

	err := s.UserRepo.UpdateProfileByID(s.Context, userID, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *userServiceImpl) UploadUserAvatar(ctx *fiber.Ctx, userID string) (*model.UploadAvatarResponse, error) {
	tr := s.Tracer.Tracer("User-UploadUserAvatar Service")
	_, span := tr.Start(s.Context, "Start UploadUserAvatar")
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
	err = s.UserRepo.PublishUploadAvatarUser(s.Context, &model.UploadAvatarRequest{
		FileName:    fileName,
		ContentType: contentType,
		Avatars:     buf.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	fileUrl := url.URL{
		Scheme: "http",
		Path:   fmt.Sprintf("%s/%s/%s", s.Config.MinIO.Endpoint, s.Config.MinIO.Bucket, fileName),
	}

	// update avatar_url users to db
	err = s.UserRepo.UpdateAvatarUserByID(s.Context, fileUrl.String(), userID)
	if err != nil {
		return nil, err
	}

	return &model.UploadAvatarResponse{
		FileURL: fileUrl.String(),
	}, nil
}

func (s *userServiceImpl) UpdateUserShorts(shortID string, req *model.ShortUserRequest) (*model.ShortUserResponse, error) {
	tr := s.Tracer.Tracer("User-UpdateUserShorts Service")
	_, span := tr.Start(s.Context, "Start UpdateUserShorts")
	defer span.End()

	err := s.UserRepo.PublishUpdateUserShortener(s.Context, shortID, req)
	if err != nil {
		return nil, err
	}

	return &model.ShortUserResponse{}, nil
}

func (s *userServiceImpl) DeleteUserShorts(shortID string) (*model.ShortUserResponse, error) {
	tr := s.Tracer.Tracer("User-DeleteUserShorts Service")
	_, span := tr.Start(s.Context, "Start DeleteUserShorts")
	defer span.End()

	err := s.UserRepo.PublishDeleteUserShortener(s.Context, shortID)
	if err != nil {
		return nil, err
	}

	return &model.ShortUserResponse{}, nil
}
