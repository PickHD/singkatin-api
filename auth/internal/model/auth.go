package model

import (
	"regexp"
)

type (
	// VerificationType consist type of verification
	VerificationType string
)

const (
	RegisterVerification       VerificationType = "register_verification"
	ForgotPasswordVerification VerificationType = "forgot_password_verification"
)

var (
	IsValidEmail, _ = regexp.Compile(`^(?P<name>[a-zA-Z0-9.!#$%&'*+/=?^_ \x60{|}~-]+)@(?P<domain>[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*)$`)
)
