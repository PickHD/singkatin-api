package response

type (
	// UserShorts consist data of user shorts
	UserShorts struct {
		ID       string `json:"id"`
		FullURL  string `json:"full_url"`
		ShortURL string `json:"short_url"`
		Visited  int64  `json:"visited"`
	}

	GetShortResponse struct {
		Shorts     []UserShorts `json:"shorts"`
		Page       int64        `json:"page"`
		Limit      int64        `json:"limit"`
		TotalCount int64        `json:"total_count"`
	}

	// GenerateShortUserMessage consist message short users to publish
	GenerateShortUserMessage struct {
		FullURL   string `json:"full_url"`
		ShortURL  string `json:"short_url"`
		UserID    string `json:"user_id"`
		ExpiresAt int64  `json:"expires_at"`
	}

	// UploadAvatarResponse consist response data when success upload avatar users
	UploadAvatarResponse struct {
		FileURL string
	}
)
