package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"singkatin-api/user/pkg/logger"
	"singkatin-api/user/pkg/utils"

	shortenerpb "singkatin-api/proto/api/v1/proto/shortener"
	"singkatin-api/user/internal/config"
	"singkatin-api/user/internal/dto/request"
	"singkatin-api/user/internal/dto/response"
	"singkatin-api/user/internal/model"
	"singkatin-api/user/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UserService is an interface that has all the function to be implemented inside user service
	UserService interface {
		GetUserDetail(ctx context.Context, email string) (*model.User, error)
		GetUserShorts(ctx context.Context, req *request.GetShortRequest) (*response.GetShortResponse, error)
		GenerateUserShorts(ctx context.Context, userID string, req *request.ShortUserRequest) error
		UpdateUserProfile(ctx context.Context, userID string, req *request.EditProfileRequest) error
		UploadUserAvatar(ctx *fiber.Ctx, userID string) (*response.UploadAvatarResponse, error)
		UpdateUserShorts(ctx context.Context, shortID string, req *request.ShortUserRequest) error
		DeleteUserShorts(ctx context.Context, shortID string) error
	}

	// userServiceImpl is an app user struct that consists of all the dependencies needed for user service
	userServiceImpl struct {
		Config       *config.Config
		Tracer       *trace.TracerProvider
		UserRepo     repository.UserRepository
		ShortClients shortenerpb.ShortenerServiceClient
	}
)

// NewUserService return new instances user service
func NewUserService(config *config.Config, tracer *trace.TracerProvider, userRepo repository.UserRepository, shortClients shortenerpb.ShortenerServiceClient) UserService {
	return &userServiceImpl{
		Config:       config,
		Tracer:       tracer,
		UserRepo:     userRepo,
		ShortClients: shortClients,
	}
}

func (s *userServiceImpl) GetUserDetail(ctx context.Context, email string) (*model.User, error) {
	tr := s.Tracer.Tracer("User-GetUserDetail Service")
	ctx, span := tr.Start(ctx, "Start GetUserDetail")
	defer span.End()

	return s.UserRepo.FindByEmail(ctx, email)
}

func (s *userServiceImpl) GetUserShorts(ctx context.Context, req *request.GetShortRequest) (*response.GetShortResponse, error) {
	tr := s.Tracer.Tracer("User-GetUserShorts Service")
	ctx, span := tr.Start(ctx, "Start GetUserShorts")
	defer span.End()

	data, err := s.ShortClients.GetListShortenerByUserID(ctx, &shortenerpb.ListShortenerRequest{
		UserId: req.UserID,
		Page:   req.Page,
		Limit:  req.Limit})
	if err != nil {
		logger.Errorf("UserServiceImpl.GetUserShorts ShortClients ERROR, %v", err)
		return nil, err
	}

	if len(data.Shorteners) < 1 {
		return &response.GetShortResponse{}, nil
	}

	shorteners := make([]response.UserShorts, len(data.Shorteners))

	for i, q := range data.Shorteners {
		shorteners[i] = response.UserShorts{
			ID:       q.GetId(),
			FullURL:  q.GetFullUrl(),
			ShortURL: q.GetShortUrl(),
			Visited:  q.GetVisited(),
		}
	}

	return &response.GetShortResponse{
		Shorts:     shorteners,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalCount: data.TotalCount,
	}, nil
}

func (s *userServiceImpl) GenerateUserShorts(ctx context.Context, userID string, req *request.ShortUserRequest) error {
	tr := s.Tracer.Tracer("User-GenerateUserShorts Service")
	ctx, span := tr.Start(ctx, "Start GenerateUserShorts")
	defer span.End()

	shortUrl := ""
	if req.CustomURL != "" {
		if len(req.CustomURL) < 3 {
			return model.NewError(model.Validation, "custom URL must be at least 3 characters")
		}

		exists, err := s.ShortClients.ExistsByShortURL(ctx, &shortenerpb.ExistsByShortURLRequest{
			ShortUrl: req.CustomURL,
			Id:       "",
		})
		if err != nil {
			return err
		}
		if exists.GetExists() {
			return model.NewError(model.Validation, "custom URL already exists")
		}

		shortUrl = req.CustomURL
	} else {
		shortUrl = utils.RandomStringBytesMaskImprSrcSB(8)
	}

	var expiresAt int64
	if req.ExpiresInDays != nil {
		expiresAt = time.Now().AddDate(0, 0, *req.ExpiresInDays).Unix()
	}

	msg := request.GenerateShortUserMessage{
		FullURL:   req.FullURL,
		UserID:    userID,
		ShortURL:  shortUrl,
		ExpiresAt: expiresAt,
	}

	err := s.UserRepo.PublishCreateUserShortener(ctx, &msg)
	if err != nil {
		return err
	}

	return nil
}

func (s *userServiceImpl) UpdateUserProfile(ctx context.Context, userID string, req *request.EditProfileRequest) error {
	tr := s.Tracer.Tracer("User-UpdateUserProfile Service")
	ctx, span := tr.Start(ctx, "Start UpdateUserProfile")
	defer span.End()

	if req.FullName == "" {
		return model.NewError(model.Validation, "Full Name Required")
	}

	if len(req.FullName) < 3 {
		return model.NewError(model.Validation, "Full Name must have at least 3 characters")
	}

	err := s.UserRepo.UpdateProfileByID(ctx, userID, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *userServiceImpl) UploadUserAvatar(c *fiber.Ctx, userID string) (*response.UploadAvatarResponse, error) {
	ctx := c.UserContext()
	tr := s.Tracer.Tracer("User-UploadUserAvatar Service")
	ctx, span := tr.Start(ctx, "Start UploadUserAvatar")
	defer span.End()

	file, err := c.FormFile("file")
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

	// Limit uploaded file buffer to 5MB to avoid OOM
	limitReader := io.LimitReader(buffer, 5<<20)

	// copy to new buffer
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, limitReader); err != nil {
		return nil, err
	}

	// publish to queue async
	err = s.UserRepo.PublishUploadAvatarUser(ctx, &request.UploadAvatarRequest{
		FileName:    fileName,
		ContentType: contentType,
		Avatars:     buf.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	fileUrl := url.URL{
		Scheme: "http",
		Host:   s.Config.MinIO.Endpoint,
		Path:   fmt.Sprintf("/%s/%s", s.Config.MinIO.Bucket, fileName),
	}

	// update avatar_url users to db
	err = s.UserRepo.UpdateAvatarUserByID(ctx, fileUrl.String(), userID)
	if err != nil {
		return nil, err
	}

	return &response.UploadAvatarResponse{
		FileURL: fileUrl.String(),
	}, nil
}

func (s *userServiceImpl) UpdateUserShorts(ctx context.Context, shortID string, req *request.ShortUserRequest) error {
	tr := s.Tracer.Tracer("User-UpdateUserShorts Service")
	ctx, span := tr.Start(ctx, "Start UpdateUserShorts")
	defer span.End()

	if req.CustomURL != "" {
		if len(req.CustomURL) < 3 {
			return model.NewError(model.Validation, "custom URL must be at least 3 characters")
		}

		exists, err := s.ShortClients.ExistsByShortURL(ctx, &shortenerpb.ExistsByShortURLRequest{
			ShortUrl: req.CustomURL,
			Id:       shortID,
		})

		if err != nil {
			return err
		}
		if exists.GetExists() {
			return model.NewError(model.Validation, "custom URL already exists")
		}
	}

	err := s.UserRepo.PublishUpdateUserShortener(ctx, shortID, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *userServiceImpl) DeleteUserShorts(ctx context.Context, shortID string) error {
	tr := s.Tracer.Tracer("User-DeleteUserShorts Service")
	ctx, span := tr.Start(ctx, "Start DeleteUserShorts")
	defer span.End()

	err := s.UserRepo.PublishDeleteUserShortener(ctx, shortID)
	if err != nil {
		return err
	}

	return nil
}
