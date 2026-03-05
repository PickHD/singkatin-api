package request

type (
	// RegisterRequest consist request data for registering as users
	RegisterRequest struct {
		FullName string `json:"fullname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// LoginRequest consist request data for login as users
	LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// ForgotPasswordRequest consist request of forgot password users
	ForgotPasswordRequest struct {
		Email string `json:"email"`
	}

	ResetPasswordRequest struct {
		NewPassword string `json:"new_password"`
	}
)