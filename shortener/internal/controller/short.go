package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	shortenerpb "singkatin-api/proto/api/v1/proto/shortener"
	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/internal/dto/request"
	"singkatin-api/shortener/internal/model"
	"singkatin-api/shortener/internal/service"
	"singkatin-api/shortener/pkg/response"

	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// ShortController is an interface that has all the function to be implemented inside short controller
	ShortController interface {
		// grpc
		GetListShortenerByUserID(ctx context.Context, req *shortenerpb.ListShortenerRequest) (*shortenerpb.ListShortenerResponse, error)
		ExistsByShortURL(ctx context.Context, req *shortenerpb.ExistsByShortURLRequest) (*shortenerpb.ExistsByShortURLResponse, error)

		// http
		ClickShortener(ctx echo.Context) error
		GenerateQRCode(ctx echo.Context) error

		// rabbitmq
		ProcessCreateShortUser(ctx context.Context, msg *shortenerpb.CreateShortenerMessage) error
		ProcessUpdateVisitorCount(ctx context.Context, msg *shortenerpb.UpdateVisitorCountMessage) error
		ProcessUpdateShortUser(ctx context.Context, msg *shortenerpb.UpdateShortenerMessage) error
		ProcessDeleteShortUser(ctx context.Context, msg *shortenerpb.DeleteShortenerMessage) error
	}

	// ShortControllerImpl is an app short struct that consists of all the dependencies needed for short controller
	ShortControllerImpl struct {
		Config   *config.Config
		Tracer   *trace.TracerProvider
		ShortSvc service.ShortService
		shortenerpb.UnimplementedShortenerServiceServer
	}
)

// NewShortController return new instances short controller
func NewShortController(config *config.Config, tracer *trace.TracerProvider, shortSvc service.ShortService) *ShortControllerImpl {
	return &ShortControllerImpl{
		Config:   config,
		Tracer:   tracer,
		ShortSvc: shortSvc,
	}
}

func (c *ShortControllerImpl) GetListShortenerByUserID(ctx context.Context, req *shortenerpb.ListShortenerRequest) (*shortenerpb.ListShortenerResponse, error) {
	tr := c.Tracer.Tracer("Shortener-GetListShortenerByUserID Controller")
	_, span := tr.Start(ctx, "Start GetListShortenerByUserID")
	defer span.End()

	filter := &request.GetShortRequest{
		UserID: req.GetUserId(),
		Page:   req.GetPage(),
		Limit:  req.GetLimit(),
	}

	data, err := c.ShortSvc.GetListShortenerByUserID(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed Get List Shortener By UserID %s", err.Error())
	}

	if len(data.Shorts) < 1 {
		return &shortenerpb.ListShortenerResponse{}, nil
	}

	shorteners := make([]*shortenerpb.Shortener, len(data.Shorts))

	for i, q := range data.Shorts {
		shorteners[i] = &shortenerpb.Shortener{
			Id:       q.ID.Hex(),
			FullUrl:  q.FullURL,
			ShortUrl: q.ShortURL,
			Visited:  q.Visited,
		}
	}

	return &shortenerpb.ListShortenerResponse{
		Shorteners: shorteners,
		Limit:      data.Limit,
		Page:       data.Page,
		TotalCount: data.TotalCount,
	}, nil
}

func (c *ShortControllerImpl) ExistsByShortURL(ctx context.Context, req *shortenerpb.ExistsByShortURLRequest) (*shortenerpb.ExistsByShortURLResponse, error) {
	tr := c.Tracer.Tracer("Shortener-ExistsByShortURL Controller")
	_, span := tr.Start(ctx, "Start ExistsByShortURL")
	defer span.End()

	exists, err := c.ShortSvc.ExistsByShortURL(ctx, req.GetShortUrl(), req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed Exists By ShortURL %s", err.Error())
	}

	return &shortenerpb.ExistsByShortURLResponse{
		Exists: exists,
	}, nil
}

func (c *ShortControllerImpl) ClickShortener(ctx echo.Context) error {
	tr := c.Tracer.Tracer("Shortener-ClickShortener Controller")
	userCtxValue := ctx.Request().Context()
	userCtxValue, span := tr.Start(userCtxValue, "Start ClickShortener")
	defer span.End()

	req := &request.UpdateVisitorRequest{
		ShortURL:  ctx.Param("short_url"),
		UserAgent: ctx.Request().UserAgent(),
		IPAddress: ctx.RealIP(),
		Referer:   ctx.Request().Referer(),
	}

	data, err := c.ShortSvc.ClickShort(userCtxValue, req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			return response.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), ctx.Param("short_url"), err, nil)
		}

		if strings.Contains(err.Error(), string(model.NotFound)) {
			return response.NewResponses[any](ctx, http.StatusNotFound, err.Error(), ctx.Param("short_url"), err, nil)
		}

		if strings.Contains(err.Error(), string(model.Gone)) {
			return response.NewResponses[any](ctx, http.StatusGone, err.Error(), ctx.Param("short_url"), err, nil)
		}

		return response.NewResponses[any](ctx, http.StatusInternalServerError, "failed click shortener", ctx.Param("short_url"), err, nil)
	}

	statusCode := http.StatusMovedPermanently
	if !data.Permanent {
		statusCode = http.StatusFound
	}

	return ctx.Redirect(statusCode, data.FullURL)
}

func (c *ShortControllerImpl) GenerateQRCode(ctx echo.Context) error {
	tr := c.Tracer.Tracer("Shortener-GenerateQRCode Controller")
	userCtxValue := ctx.Request().Context()
	userCtxValue, span := tr.Start(userCtxValue, "Start GenerateQRCode")
	defer span.End()

	data, err := c.ShortSvc.GetByShortURL(userCtxValue, ctx.Param("short_url"))
	if err != nil {
		return response.NewResponses[any](ctx, http.StatusInternalServerError, "failed generate QR code", ctx.Param("short_url"), err, nil)
	}

	if data == nil {
		return response.NewResponses[any](ctx, http.StatusNotFound, "short URL not found", ctx.Param("short_url"), nil, nil)
	}

	shortLink := fmt.Sprintf("%s/%s", c.Config.Server.BaseURL, ctx.Param("short_url"))
	png, err := qrcode.Encode(shortLink, qrcode.Medium, 512)
	if err != nil {
		return response.NewResponses[any](ctx, http.StatusInternalServerError, "failed generate QR code", ctx.Param("short_url"), err, nil)
	}

	return ctx.Blob(http.StatusOK, "image/png", png)
}

func (c *ShortControllerImpl) ProcessCreateShortUser(ctx context.Context, msg *shortenerpb.CreateShortenerMessage) error {
	tr := c.Tracer.Tracer("Shortener-ProcessCreateShortUser Controller")
	ctx, span := tr.Start(ctx, "Start ProcessCreateShortUser")
	defer span.End()

	var expiresAt *time.Time
	if msg.GetExpiresAt() > 0 {
		t := time.Unix(msg.GetExpiresAt(), 0)
		expiresAt = &t
	}

	req := &request.CreateShortRequest{
		UserID:    msg.GetUserId(),
		FullURL:   msg.GetFullUrl(),
		ShortURL:  msg.GetShortUrl(),
		ExpiresAt: expiresAt,
	}

	err := c.ShortSvc.CreateShort(ctx, req)
	if err != nil {
		return model.NewError(model.Internal, err.Error())
	}

	return nil
}

func (c *ShortControllerImpl) ProcessUpdateVisitorCount(ctx context.Context, msg *shortenerpb.UpdateVisitorCountMessage) error {
	tr := c.Tracer.Tracer("Shortener-ProcessUpdateVisitorCount Controller")
	ctx, span := tr.Start(ctx, "Start ProcessUpdateVisitorCount")
	defer span.End()

	req := &request.UpdateVisitorRequest{
		ShortURL:  msg.GetShortUrl(),
		UserAgent: msg.GetUserAgent(),
		IPAddress: msg.GetIpAddress(),
		Referer:   msg.GetReferer(),
	}

	err := c.ShortSvc.UpdateVisitorShort(ctx, req)
	if err != nil {
		return model.NewError(model.Internal, err.Error())
	}

	return nil
}

func (c *ShortControllerImpl) ProcessUpdateShortUser(ctx context.Context, msg *shortenerpb.UpdateShortenerMessage) error {
	tr := c.Tracer.Tracer("Shortener-ProcessUpdateShortUser Controller")
	ctx, span := tr.Start(ctx, "Start ProcessUpdateShortUser")
	defer span.End()

	req := &request.UpdateShortRequest{
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
	ctx, span := tr.Start(ctx, "Start ProcessDeleteShortUser")
	defer span.End()

	req := &request.DeleteShortRequest{
		ID: msg.GetId(),
	}

	err := c.ShortSvc.DeleteShort(ctx, req)
	if err != nil {
		return model.NewError(model.Internal, err.Error())
	}

	return nil
}
