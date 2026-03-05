package request

type (
	// ShortUserRequest consist request data generate/update short users
	ShortUserRequest struct {
		FullURL       string `json:"full_url"`
		CustomURL     string `json:"custom_url"`
		ExpiresInDays *int   `json:"expires_in_days"`
	}

	// EditProfileRequest consist request data edit profile users
	EditProfileRequest struct {
		FullName string `json:"full_name"`
	}

	// UploadAvatarRequest consist request data upload avatar users
	UploadAvatarRequest struct {
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
		Avatars     []byte `json:"avatars"`
	}

	GenerateShortUserMessage struct {
		FullURL   string `json:"full_url"`
		ShortURL  string `json:"short_url"`
		UserID    string `json:"user_id"`
		ExpiresAt int64  `json:"expires_at"`
	}

	GetShortRequest struct {
		UserID string `json:"-"`
		Page   int64  `json:"-"`
		Limit  int64  `json:"-"`
	}
)
