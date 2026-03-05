package service

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/internal/dto/request"
	"singkatin-api/shortener/internal/dto/response"
	"singkatin-api/shortener/internal/model"
	"singkatin-api/shortener/internal/repository"
	"singkatin-api/shortener/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// ShortService is an interface that has all the function to be implemented inside short service
	ShortService interface {
		GetListShortenerByUserID(ctx context.Context, req *request.GetShortRequest) (*response.GetShortResponse, error)
		CreateShort(ctx context.Context, req *request.CreateShortRequest) error
		ClickShort(ctx context.Context, req *request.UpdateVisitorRequest) (*response.ClickShortResponse, error)
		UpdateVisitorShort(ctx context.Context, req *request.UpdateVisitorRequest) error
		UpdateShort(ctx context.Context, req *request.UpdateShortRequest) error
		DeleteShort(ctx context.Context, req *request.DeleteShortRequest) error
		ExistsByShortURL(ctx context.Context, shortURL string, id string) (bool, error)
		GetByShortURL(ctx context.Context, shortURL string) (*model.Short, error)
	}

	// shortServiceImpl is an app short struct that consists of all the dependencies needed for short repository
	shortServiceImpl struct {
		Config    *config.Config
		Tracer    *trace.TracerProvider
		ShortRepo repository.ShortRepository
	}
)

// NewShortService return new instances short service
func NewShortService(config *config.Config, tracer *trace.TracerProvider, shortRepo repository.ShortRepository) ShortService {
	return &shortServiceImpl{
		Config:    config,
		Tracer:    tracer,
		ShortRepo: shortRepo,
	}
}

func (s *shortServiceImpl) GetListShortenerByUserID(ctx context.Context, req *request.GetShortRequest) (*response.GetShortResponse, error) {
	tr := s.Tracer.Tracer("Shortener-GetListShortenerByUserID Service")
	ctx, span := tr.Start(ctx, "Start GetListShortenerByUserID")
	defer span.End()

	data, err := s.ShortRepo.GetListShortenerByUserID(ctx, req)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *shortServiceImpl) CreateShort(ctx context.Context, req *request.CreateShortRequest) error {
	tr := s.Tracer.Tracer("Shortener-CreateShort Service")
	ctx, span := tr.Start(ctx, "Start CreateShort")
	defer span.End()

	err := s.validateCreateShort(req)
	if err != nil {
		return err
	}

	// check if full url already exists
	exists, err := s.ShortRepo.GetByFullURLAndUserID(ctx, req)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if exists != nil {
		return model.NewError(model.Duplicate, fmt.Sprintf("URL already shortened as %s", exists.ShortURL))
	}

	// check if custom url already exists
	if req.CustomURL != "" {
		exists, err := s.ShortRepo.ExistsByShortURL(ctx, req.CustomURL, "")
		if err != nil {
			return err
		}
		if exists {
			return model.NewError(model.Validation, "custom URL already exists")
		}
	}

	return s.ShortRepo.Create(ctx, &model.Short{
		FullURL:   req.FullURL,
		ShortURL:  req.ShortURL,
		UserID:    req.UserID,
		ExpiresAt: req.ExpiresAt,
	})
}

func (s *shortServiceImpl) ClickShort(ctx context.Context, req *request.UpdateVisitorRequest) (*response.ClickShortResponse, error) {
	var (
		redisTTLDuration = time.Minute * time.Duration(s.Config.Redis.TTL)
	)

	tr := s.Tracer.Tracer("Shortener-ClickShort Service")
	ctx, span := tr.Start(ctx, "Start ClickShort")
	defer span.End()

	err := s.validateClickShort(req)
	if err != nil {
		return nil, err
	}

	cachedFullURL, isPermanent, err := s.ShortRepo.GetFullURLByKey(ctx, req.ShortURL)
	if err != nil {
		if err == redis.Nil {
			logger.Info("get data from default databases....")

			data, err := s.ShortRepo.GetByShortURL(ctx, req.ShortURL)
			if err != nil {
				return nil, err
			}

			if data.ExpiresAt != nil && data.ExpiresAt.Before(time.Now()) {
				return nil, model.NewError(model.Gone, "link has expired")
			}

			ttl := redisTTLDuration
			if data.ExpiresAt != nil {
				expirationTTL := time.Until(*data.ExpiresAt)
				if expirationTTL < ttl {
					ttl = expirationTTL
				}
			}

			err = s.ShortRepo.SetFullURLByKey(ctx, req.ShortURL, data.FullURL, data.ExpiresAt == nil, ttl)
			if err != nil {
				return nil, err
			}

			err = s.ShortRepo.PublishUpdateVisitorCount(ctx, req)
			if err != nil {
				return nil, err
			}

			return &response.ClickShortResponse{FullURL: data.FullURL, Permanent: data.ExpiresAt == nil}, nil
		}

		return nil, err
	}

	logger.Info("get data from caching....")

	err = s.ShortRepo.PublishUpdateVisitorCount(ctx, req)
	if err != nil {
		return nil, err
	}

	return &response.ClickShortResponse{FullURL: cachedFullURL, Permanent: isPermanent}, nil
}

func (s *shortServiceImpl) UpdateVisitorShort(ctx context.Context, req *request.UpdateVisitorRequest) error {
	tr := s.Tracer.Tracer("Shortener-UpdateVisitorShort Service")
	ctx, span := tr.Start(ctx, "Start UpdateVisitorShort")
	defer span.End()

	return s.ShortRepo.UpdateVisitorByShortURL(ctx, req)
}

func (s *shortServiceImpl) UpdateShort(ctx context.Context, req *request.UpdateShortRequest) error {
	tr := s.Tracer.Tracer("Shortener-UpdateShort Service")
	ctx, span := tr.Start(ctx, "Start UpdateShort")
	defer span.End()

	if _, err := url.ParseRequestURI(req.FullURL); err != nil {
		return model.NewError(model.Validation, err.Error())
	}

	data, err := s.ShortRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	// delete cache if any
	err = s.ShortRepo.DeleteFullURLByKey(ctx, data.ShortURL)
	if err != nil {
		return err
	}

	return s.ShortRepo.UpdateFullURLByID(ctx, req)
}

func (s *shortServiceImpl) DeleteShort(ctx context.Context, req *request.DeleteShortRequest) error {
	tr := s.Tracer.Tracer("Shortener-DeleteShort Service")
	ctx, span := tr.Start(ctx, "Start DeleteShort")
	defer span.End()

	data, err := s.ShortRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	// delete cache if any
	err = s.ShortRepo.DeleteFullURLByKey(ctx, data.ShortURL)
	if err != nil {
		return err
	}

	return s.ShortRepo.DeleteByID(ctx, req)
}

func (s *shortServiceImpl) ExistsByShortURL(ctx context.Context, shortURL string, id string) (bool, error) {
	tr := s.Tracer.Tracer("Shortener-ExistsByShortURL Service")
	ctx, span := tr.Start(ctx, "Start ExistsByShortURL")
	defer span.End()

	return s.ShortRepo.ExistsByShortURL(ctx, shortURL, id)
}

func (s *shortServiceImpl) GetByShortURL(ctx context.Context, shortURL string) (*model.Short, error) {
	tr := s.Tracer.Tracer("Shortener-GetByShortURL Service")
	ctx, span := tr.Start(ctx, "Start GetByShortURL")
	defer span.End()

	return s.ShortRepo.GetByShortURL(ctx, shortURL)
}

func (s *shortServiceImpl) validateCreateShort(req *request.CreateShortRequest) error {
	if _, err := url.ParseRequestURI(req.FullURL); err != nil {
		return model.NewError(model.Validation, err.Error())
	}

	return nil
}

func (s *shortServiceImpl) validateClickShort(req *request.UpdateVisitorRequest) error {
	if req.ShortURL == "" {
		return model.NewError(model.Validation, "short URL cannot be empty")
	}

	if len(req.ShortURL) < 3 {
		return model.NewError(model.Validation, "short URL must be at least 3 characters")
	}

	return nil
}
