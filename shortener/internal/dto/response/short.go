package response

import "singkatin-api/shortener/internal/model"

type (
	ClickShortResponse struct {
		FullURL   string `json:"full_url"`
		Permanent bool   `json:"permanent"`
	}

	GetShortResponse struct {
		Shorts     []model.Short
		TotalCount int64
		Page       int64
		Limit      int64
	}
)
