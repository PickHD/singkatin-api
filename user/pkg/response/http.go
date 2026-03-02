package response

import "github.com/gofiber/fiber/v2"

type (
	BaseResponse struct {
		Messages string `json:"messages"`
		Data     any    `json:"data"`
		Error    error  `json:"error,omitempty"`
		Meta     *Meta  `json:"meta,omitempty"`
	}

	Meta struct {
		Page      int64 `json:"page,omitempty"`
		TotalPage int64 `json:"total_page,omitempty"`
		TotalData int64 `json:"total_data,omitempty"`
	}
)

// NewResponses return dynamic JSON responses
func NewResponses[T any](ctx *fiber.Ctx, statusCode int, message string, data T, err error, meta *Meta) error {
	if statusCode < 400 {
		return ctx.Status(statusCode).JSON(&BaseResponse{
			Messages: message,
			Data:     data,
			Error:    nil,
			Meta:     meta,
		})
	}

	return ctx.Status(statusCode).JSON(&BaseResponse{
		Messages: message,
		Data:     data,
		Error:    err,
		Meta:     nil,
	})
}
