package controller

import (
	"context"
	"net/http"
	"strings"

	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/internal/model"
	"singkatin-api/shortener/internal/service"
	shortenerpb "singkatin-api/shortener/pkg/api/v1/proto/shortener"
	"singkatin-api/shortener/pkg/response"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// ShortController is an interface that has all the function to be implemented inside short controller
	ShortController interface {
		// grpc
		GetListShortenerByUserID(ctx context.Context, req *shortenerpb.ListShortenerRequest) (*shortenerpb.ListShortenerResponse, error)

		// http
		ClickShortener(ctx echo.Context) error

		// rabbitmq
		ProcessCreateShortUser(ctx context.Context, msg *shortenerpb.CreateShortenerMessage) error
		ProcessUpdateVisitorCount(ctx context.Context, msg *shortenerpb.UpdateVisitorCountMessage) error
		ProcessUpdateShortUser(ctx context.Context, msg *shortenerpb.UpdateShortenerMessage) error
		ProcessDeleteShortUser(ctx context.Context, msg *shortenerpb.DeleteShortenerMessage) error
	}

	// ShortControllerImpl is an app short struct that consists of all the dependencies needed for short controller
	ShortControllerImpl struct {
		Context  context.Context
		Config   *config.Configuration
		Tracer   *trace.TracerProvider
		ShortSvc service.ShortService
		shortenerpb.UnimplementedShortenerServiceServer
	}
)

// NewShortController return new instances short controller
func NewShortController(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, shortSvc service.ShortService) *ShortControllerImpl {
	return &ShortControllerImpl{
		Context:  ctx,
		Config:   config,
		Tracer:   tracer,
		ShortSvc: shortSvc,
	}
}

func (c *ShortControllerImpl) GetListShortenerByUserID(ctx context.Context, req *shortenerpb.ListShortenerRequest) (*shortenerpb.ListShortenerResponse, error) {
	tr := c.Tracer.Tracer("Shortener-GetListShortenerByUserID Controller")
	_, span := tr.Start(ctx, "Start GetListShortenerByUserID")
	defer span.End()

	data, err := c.ShortSvc.GetListShortenerByUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed Get List Shortener By UserID %s", err.Error())
	}

	if len(data) < 1 {
		return &shortenerpb.ListShortenerResponse{}, nil
	}

	shorteners := make([]*shortenerpb.Shortener, len(data))

	for i, q := range data {
		shorteners[i] = &shortenerpb.Shortener{
			Id:       q.ID.Hex(),
			FullUrl:  q.FullURL,
			ShortUrl: q.ShortURL,
			Visited:  q.Visited,
		}
	}

	return &shortenerpb.ListShortenerResponse{
		Shorteners: shorteners,
	}, nil
}

func (c *ShortControllerImpl) ClickShortener(ctx echo.Context) error {
	tr := c.Tracer.Tracer("Shortener-ClickShortener Controller")
	_, span := tr.Start(c.Context, "Start ClickShortener")
	defer span.End()

	data, err := c.ShortSvc.ClickShort(ctx.Param("short_url"))
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			return response.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), ctx.Param("short_url"), err, nil)
		}

		if strings.Contains(err.Error(), string(model.NotFound)) {
			return response.NewResponses[any](ctx, http.StatusNotFound, err.Error(), ctx.Param("short_url"), err, nil)
		}

		return response.NewResponses[any](ctx, http.StatusInternalServerError, "failed click shortener", ctx.Param("short_url"), err, nil)
	}

	return ctx.Redirect(http.StatusTemporaryRedirect, data.FullURL)
}

func (c *ShortControllerImpl) ProcessCreateShortUser(ctx context.Context, msg *shortenerpb.CreateShortenerMessage) error {
	tr := c.Tracer.Tracer("Shortener-ProcessCreateShortUser Controller")
	_, span := tr.Start(c.Context, "Start ProcessCreateShortUser")
	defer span.End()

	req := &model.CreateShortRequest{
		UserID:   msg.GetUserId(),
		FullURL:  msg.GetFullUrl(),
		ShortURL: msg.GetShortUrl(),
	}

	err := c.ShortSvc.CreateShort(ctx, req)
	if err != nil {
		return model.NewError(model.Internal, err.Error())
	}

	return nil
}

func (c *ShortControllerImpl) ProcessUpdateVisitorCount(ctx context.Context, msg *shortenerpb.UpdateVisitorCountMessage) error {
	tr := c.Tracer.Tracer("Shortener-ProcessUpdateVisitorCount Controller")
	_, span := tr.Start(c.Context, "Start ProcessUpdateVisitorCount")
	defer span.End()

	req := &model.UpdateVisitorRequest{
		ShortURL: msg.GetShortUrl(),
	}

	err := c.ShortSvc.UpdateVisitorShort(ctx, req)
	if err != nil {
		return model.NewError(model.Internal, err.Error())
	}

	return nil
}

func (c *ShortControllerImpl) ProcessUpdateShortUser(ctx context.Context, msg *shortenerpb.UpdateShortenerMessage) error {
	tr := c.Tracer.Tracer("Shortener-ProcessUpdateShortUser Controller")
	_, span := tr.Start(c.Context, "Start ProcessUpdateShortUser")
	defer span.End()

	req := &model.UpdateShortRequest{
		ID:      msg.GetId(),
		FullURL: msg.GetFullUrl(),
	}

	err := c.ShortSvc.UpdateShort(ctx, req)
	if err != nil {
		return model.NewError(model.Internal, err.Error())
	}

	return nil
}

func (c *ShortControllerImpl) ProcessDeleteShortUser(ctx context.Context, msg *shortenerpb.DeleteShortenerMessage) error {
	tr := c.Tracer.Tracer("Shortener-ProcessDeleteShortUser Controller")
	_, span := tr.Start(c.Context, "Start ProcessDeleteShortUser")
	defer span.End()

	req := &model.DeleteShortRequest{
		ID: msg.GetId(),
	}

	err := c.ShortSvc.DeleteShort(ctx, req)
	if err != nil {
		return model.NewError(model.Internal, err.Error())
	}

	return nil
}
