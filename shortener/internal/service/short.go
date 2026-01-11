package service

import (
	"context"
	"net/url"
	"time"

	"github.com/PickHD/singkatin-revamp/shortener/internal/config"
	"github.com/PickHD/singkatin-revamp/shortener/internal/model"
	"github.com/PickHD/singkatin-revamp/shortener/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// ShortService is an interface that has all the function to be implemented inside short service
	ShortService interface {
		GetListShortenerByUserID(ctx context.Context, userID string) ([]model.Short, error)
		CreateShort(ctx context.Context, req *model.CreateShortRequest) error
		ClickShort(shortURL string) (*model.ClickShortResponse, error)
		UpdateVisitorShort(ctx context.Context, req *model.UpdateVisitorRequest) error
		UpdateShort(ctx context.Context, req *model.UpdateShortRequest) error
		DeleteShort(ctx context.Context, req *model.DeleteShortRequest) error
	}

	// ShortServiceImpl is an app short struct that consists of all the dependencies needed for short repository
	ShortServiceImpl struct {
		Context   context.Context
		Config    *config.Configuration
		Logger    *logrus.Logger
		Tracer    *trace.TracerProvider
		ShortRepo repository.ShortRepository
	}
)

// NewShortService return new instances short service
func NewShortService(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, shortRepo repository.ShortRepository) *ShortServiceImpl {
	return &ShortServiceImpl{
		Context:   ctx,
		Config:    config,
		Logger:    logger,
		Tracer:    tracer,
		ShortRepo: shortRepo,
	}
}

func (ss *ShortServiceImpl) GetListShortenerByUserID(ctx context.Context, userID string) ([]model.Short, error) {
	tr := ss.Tracer.Tracer("Shortener-GetListShortenerByUserID Service")
	ctx, span := tr.Start(ctx, "Start GetListShortenerByUserID")
	defer span.End()

	data, err := ss.ShortRepo.GetListShortenerByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (ss *ShortServiceImpl) CreateShort(ctx context.Context, req *model.CreateShortRequest) error {
	tr := ss.Tracer.Tracer("Shortener-CreateShort Service")
	ctx, span := tr.Start(ctx, "Start CreateShort")
	defer span.End()

	err := ss.validateCreateShort(req)
	if err != nil {
		return err
	}

	return ss.ShortRepo.Create(ctx, &model.Short{
		FullURL:  req.FullURL,
		ShortURL: req.ShortURL,
		UserID:   req.UserID,
	})
}

func (ss *ShortServiceImpl) ClickShort(shortURL string) (*model.ClickShortResponse, error) {
	var (
		redisTTLDuration = time.Minute * time.Duration(ss.Config.Redis.TTL)
	)

	tr := ss.Tracer.Tracer("Shortener-ClickShort Service")
	ctx, span := tr.Start(ss.Context, "Start ClickShort")
	defer span.End()

	req := &model.UpdateVisitorRequest{ShortURL: shortURL}

	err := ss.validateClickShort(req)
	if err != nil {
		return nil, err
	}

	cachedFullURL, err := ss.ShortRepo.GetFullURLByKey(ctx, req.ShortURL)
	if err != nil {
		if err == redis.Nil {
			ss.Logger.Info("get data from default databases....")

			data, err := ss.ShortRepo.GetByShortURL(ss.Context, req.ShortURL)
			if err != nil {
				return nil, err
			}

			err = ss.ShortRepo.SetFullURLByKey(ss.Context, req.ShortURL, data.FullURL, redisTTLDuration)
			if err != nil {
				return nil, err
			}

			err = ss.ShortRepo.PublishUpdateVisitorCount(ss.Context, req)
			if err != nil {
				return nil, err
			}

			return &model.ClickShortResponse{FullURL: data.FullURL}, nil
		}

		return nil, err
	}

	ss.Logger.Info("get data from caching....")

	err = ss.ShortRepo.PublishUpdateVisitorCount(ctx, req)
	if err != nil {
		return nil, err
	}

	return &model.ClickShortResponse{FullURL: cachedFullURL}, nil
}

func (ss *ShortServiceImpl) UpdateVisitorShort(ctx context.Context, req *model.UpdateVisitorRequest) error {
	tr := ss.Tracer.Tracer("Shortener-UpdateVisitorShort Service")
	ctx, span := tr.Start(ctx, "Start UpdateVisitorShort")
	defer span.End()

	data, err := ss.ShortRepo.GetByShortURL(ctx, req.ShortURL)
	if err != nil {
		return err
	}

	err = ss.ShortRepo.UpdateVisitorByShortURL(ctx, req, data.Visited)
	if err != nil {
		return err
	}

	return nil
}

func (ss *ShortServiceImpl) UpdateShort(ctx context.Context, req *model.UpdateShortRequest) error {
	tr := ss.Tracer.Tracer("Shortener-UpdateShort Service")
	ctx, span := tr.Start(ctx, "Start UpdateShort")
	defer span.End()

	if _, err := url.ParseRequestURI(req.FullURL); err != nil {
		return model.NewError(model.Validation, err.Error())
	}

	data, err := ss.ShortRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	// delete cache if any
	err = ss.ShortRepo.DeleteFullURLByKey(ctx, data.ShortURL)
	if err != nil {
		return err
	}

	return ss.ShortRepo.UpdateFullURLByID(ctx, req)
}

func (ss *ShortServiceImpl) DeleteShort(ctx context.Context, req *model.DeleteShortRequest) error {
	tr := ss.Tracer.Tracer("Shortener-DeleteShort Service")
	ctx, span := tr.Start(ctx, "Start DeleteShort")
	defer span.End()

	data, err := ss.ShortRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	// delete cache if any
	err = ss.ShortRepo.DeleteFullURLByKey(ctx, data.ShortURL)
	if err != nil {
		return err
	}

	return ss.ShortRepo.DeleteByID(ctx, req)
}

func (ss *ShortServiceImpl) validateCreateShort(req *model.CreateShortRequest) error {
	if _, err := url.ParseRequestURI(req.FullURL); err != nil {
		return model.NewError(model.Validation, err.Error())
	}

	return nil
}

func (ss *ShortServiceImpl) validateClickShort(req *model.UpdateVisitorRequest) error {
	if req.ShortURL == "" {
		return model.NewError(model.Validation, "short URL cannot be empty")
	}

	if len(req.ShortURL) != 8 {
		return model.NewError(model.Validation, "short URL length must be 8")
	}

	return nil
}
