package response

type (
	// RegisterResponse consist response of success registering as users
	RegisterResponse struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		IsVerified bool   `json:"is_verified"`
	}

	// LoginResponse consist response of success login as users
	LoginResponse struct {
		AccessToken string `json:"access_token"`
		Type        string `json:"type"`
	}

	VerifyCodeResponse struct {
		IsVerified bool `json:"is_verified"`
	}
)
