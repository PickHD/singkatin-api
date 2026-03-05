package request

import "time"

type (
	CreateShortRequest struct {
		UserID    string     `json:"user_id"`
		FullURL   string     `json:"full_url"`
		ShortURL  string     `json:"short_url"`
		CustomURL string     `json:"custom_url"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	GetShortRequest struct {
		UserID string `json:"user_id"`
		Page   int64  `json:"page"`
		Limit  int64  `json:"limit"`
	}

	UpdateVisitorRequest struct {
		ShortURL  string `json:"short_url"`
		UserAgent string `json:"user_agent"`
		IPAddress string `json:"ip_address"`
		Referer   string `json:"referer"`
	}

	UpdateShortRequest struct {
		ID      string `json:"id"`
		FullURL string `json:"full_url"`
	}

	DeleteShortRequest struct {
		ID string `json:"id"`
	}
)
