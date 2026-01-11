package model

import (
	"regexp"
	"time"
)

type (
	// RegisterRequest consist request data for registering as users
	RegisterRequest struct {
		FullName string `json:"fullname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// RegisterResponse consist response of success registering as users
	RegisterResponse struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		IsVerified bool   `json:"is_verified"`
	}

	// LoginRequest consist request data for login as users
	LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// LoginResponse consist response of success login as users
	LoginResponse struct {
		AccessToken string    `json:"access_token"`
		Type        string    `json:"type"`
		ExpireAt    time.Time `json:"expired_at"`
	}

	VerifyCodeResponse struct {
		IsVerified bool `json:"is_verified"`
	}

	// ForgotPasswordRequest consist request of forgot password users
	ForgotPasswordRequest struct {
		Email string `json:"email"`
	}

	// VerificationType consist type of verification
	VerificationType string

	ResetPasswordRequest struct {
		NewPassword string `json:"new_password"`
	}
)

const (
	RegisterVerification       VerificationType = "register_verification"
	ForgotPasswordVerification VerificationType = "forgot_password_verification"
)

var (
	IsValidEmail, _ = regexp.Compile(`^(?P<name>[a-zA-Z0-9.!#$%&'*+/=?^_ \x60{|}~-]+)@(?P<domain>[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*)$`)
)
