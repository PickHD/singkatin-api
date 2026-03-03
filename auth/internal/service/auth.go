package service

import (
	"context"
	"fmt"
	"time"

	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/internal/infrastructure"
	"singkatin-api/auth/internal/model"
	"singkatin-api/auth/internal/repository"
	"singkatin-api/auth/pkg/logger"
	"singkatin-api/auth/pkg/utils"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// AuthService is an interface that has all the function to be implemented inside auth service
	AuthService interface {
		RegisterUser(ctx context.Context, req *model.RegisterRequest) (*model.RegisterResponse, error)
		LoginUser(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
		VerifyCode(ctx context.Context, code string, verificationType model.VerificationType) (*model.VerifyCodeResponse, error)
		ForgotPasswordUser(ctx context.Context, req *model.ForgotPasswordRequest) error
		ResetPasswordUser(ctx context.Context, req *model.ResetPasswordRequest, code string) error
	}

	// authServiceImpl is an app auth struct that consists of all the dependencies needed for auth service
	authServiceImpl struct {
		Config   *config.Config
		Tracer   *trace.TracerProvider
		Mailer   *infrastructure.EmailProvider
		AuthRepo repository.AuthRepository
		JWT      *infrastructure.JwtProvider
	}
)

// NewAuthService return new instances auth service
func NewAuthService(config *config.Config, tracer *trace.TracerProvider, mailer *infrastructure.EmailProvider, authRepo repository.AuthRepository, jwt *infrastructure.JwtProvider) AuthService {
	return &authServiceImpl{
		Config:   config,
		Tracer:   tracer,
		Mailer:   mailer,
		AuthRepo: authRepo,
		JWT:      jwt,
	}
}

func (s *authServiceImpl) RegisterUser(ctx context.Context, req *model.RegisterRequest) (*model.RegisterResponse, error) {
	tr := s.Tracer.Tracer("Auth-RegisterUser service")
	ctx, span := tr.Start(ctx, "Start RegisterUser")
	defer span.End()

	err := validateRegisterUser(req)
	if err != nil {
		return nil, err
	}

	hashPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	data, err := s.AuthRepo.CreateUser(ctx, &model.User{
		FullName:   req.FullName,
		Email:      req.Email,
		Password:   hashPass,
		IsVerified: false,
	})
	if err != nil {
		return nil, err
	}

	codeVerification := utils.RandomStringBytesMaskImprSrcSB(9)
	expiredCodeDuration := time.Minute * time.Duration(s.Config.Redis.TTL)

	err = s.AuthRepo.SetVerificationByEmail(ctx, req.Email, codeVerification, expiredCodeDuration, model.RegisterVerification)
	if err != nil {
		return nil, err
	}

	verifyURL := s.Config.Server.BaseURL + "/register/verify?code=" + codeVerification

	emailLink := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	</head>
	<body style="margin:0;padding:0;background-color:#f4f4f7;font-family:'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
		<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f7;padding:40px 0;">
			<tr><td align="center">
				<table width="480" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.08);">
					<!-- Header -->
					<tr><td style="background:linear-gradient(135deg,#6366f1,#8b5cf6);padding:32px 40px;text-align:center;">
						<h1 style="margin:0;color:#ffffff;font-size:24px;font-weight:700;letter-spacing:-0.5px;">Singkatin</h1>
						<p style="margin:8px 0 0;color:rgba(255,255,255,0.85);font-size:14px;">URL Shortener</p>
					</td></tr>
					<!-- Body -->
					<tr><td style="padding:40px;">
						<h2 style="margin:0 0 12px;color:#1a1a2e;font-size:20px;font-weight:600;">Verify Your Email</h2>
						<p style="margin:0 0 24px;color:#555770;font-size:15px;line-height:1.6;">Thanks for signing up! Please click the button below to verify your email address and activate your account.</p>
						<table width="100%%" cellpadding="0" cellspacing="0"><tr><td align="center" style="padding:8px 0 24px;">
							<a href="%s" style="display:inline-block;padding:14px 36px;background:linear-gradient(135deg,#6366f1,#8b5cf6);color:#ffffff;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;letter-spacing:0.3px;">Verify Email Address</a>
						</td></tr></table>
						<p style="margin:0 0 8px;color:#9ca3af;font-size:13px;">This link will expire in %d minutes.</p>
						<p style="margin:0;color:#9ca3af;font-size:13px;">If you didn't create an account, you can safely ignore this email.</p>
					</td></tr>
					<!-- Footer -->
					<tr><td style="padding:24px 40px;background-color:#f9fafb;border-top:1px solid #e5e7eb;text-align:center;">
						<p style="margin:0;color:#9ca3af;font-size:12px;">&copy; 2026 Singkatin. All rights reserved.</p>
					</td></tr>
				</table>
			</td></tr>
		</table>
	</body>
	</html>`, verifyURL, s.Config.Redis.TTL)

	err = s.Mailer.SendEmail(ctx, req.Email, "Registration Confirmations", emailLink)
	if err != nil {
		logger.Errorf("AuthServiceImpl.RegisterUser() sendMail ERROR, %v", err)
		return nil, err
	}

	return &model.RegisterResponse{
		ID:         data.ID.Hex(),
		Email:      data.Email,
		IsVerified: false,
	}, nil
}

func (s *authServiceImpl) LoginUser(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	tr := s.Tracer.Tracer("Auth-LoginUser service")
	ctx, span := tr.Start(ctx, "Start LoginUser")
	defer span.End()

	err := validateLoginUser(req)
	if err != nil {
		return nil, err
	}

	user, err := s.AuthRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// verify user password by comparing incoming request password with crypted password stored in database
	if !utils.CheckPasswordHash(user.Password, req.Password) {
		return nil, model.NewError(model.Validation, "invalid password")
	}

	// generate access token jwt
	token, err := s.JWT.GenerateToken(user.ID.Hex(), user.FullName, user.Email)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken: token,
		Type:        "Bearer",
	}, nil
}

func (s *authServiceImpl) VerifyCode(ctx context.Context, code string, verificationType model.VerificationType) (*model.VerifyCodeResponse, error) {
	tr := s.Tracer.Tracer("Auth-VerifyCode service")
	ctx, span := tr.Start(ctx, "Start VerifyCode")
	defer span.End()

	getEmail, err := s.AuthRepo.GetVerificationByCode(ctx, code, verificationType)
	if err != nil {
		if err == redis.Nil {
			return nil, model.NewError(model.NotFound, "code not found / expired")
		}

		return nil, err
	}

	switch verificationType {
	case model.RegisterVerification:
		err = s.AuthRepo.UpdateVerifyStatusByEmail(ctx, getEmail)
		if err != nil {
			return nil, err
		}
	case model.ForgotPasswordVerification:
	}

	return &model.VerifyCodeResponse{
		IsVerified: true,
	}, nil
}

func (s *authServiceImpl) ForgotPasswordUser(ctx context.Context, req *model.ForgotPasswordRequest) error {
	tr := s.Tracer.Tracer("Auth-ForgotPasswordUser service")
	ctx, span := tr.Start(ctx, "Start ForgotPasswordUser")
	defer span.End()

	if !model.IsValidEmail.MatchString(req.Email) {
		return model.NewError(model.Validation, "invalid email")
	}

	_, err := s.AuthRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	codeVerification := utils.RandomStringBytesMaskImprSrcSB(10)
	expiredCodeDuration := time.Minute * time.Duration(s.Config.Redis.TTL)

	err = s.AuthRepo.SetVerificationByEmail(ctx, req.Email, codeVerification, expiredCodeDuration, model.ForgotPasswordVerification)
	if err != nil {
		return err
	}

	verifyURL := s.Config.Server.BaseURL + "/forgot-password/verify?code=" + codeVerification

	emailLink := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	</head>
	<body style="margin:0;padding:0;background-color:#f4f4f7;font-family:'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
		<table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f7;padding:40px 0;">
			<tr><td align="center">
				<table width="480" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.08);">
					<!-- Header -->
					<tr><td style="background:linear-gradient(135deg,#ef4444,#f97316);padding:32px 40px;text-align:center;">
						<h1 style="margin:0;color:#ffffff;font-size:24px;font-weight:700;letter-spacing:-0.5px;">Singkatin</h1>
						<p style="margin:8px 0 0;color:rgba(255,255,255,0.85);font-size:14px;">URL Shortener</p>
					</td></tr>
					<!-- Body -->
					<tr><td style="padding:40px;">
						<h2 style="margin:0 0 12px;color:#1a1a2e;font-size:20px;font-weight:600;">Reset Your Password</h2>
						<p style="margin:0 0 24px;color:#555770;font-size:15px;line-height:1.6;">We received a request to reset your password. Click the button below to choose a new password.</p>
						<table width="100%%" cellpadding="0" cellspacing="0"><tr><td align="center" style="padding:8px 0 24px;">
							<a href="%s" style="display:inline-block;padding:14px 36px;background:linear-gradient(135deg,#ef4444,#f97316);color:#ffffff;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;letter-spacing:0.3px;">Reset Password</a>
						</td></tr></table>
						<p style="margin:0 0 8px;color:#9ca3af;font-size:13px;">This link will expire in %d minutes.</p>
						<p style="margin:0;color:#9ca3af;font-size:13px;">If you didn't request a password reset, you can safely ignore this email.</p>
					</td></tr>
					<!-- Footer -->
					<tr><td style="padding:24px 40px;background-color:#f9fafb;border-top:1px solid #e5e7eb;text-align:center;">
						<p style="margin:0;color:#9ca3af;font-size:12px;">&copy; 2026 Singkatin. All rights reserved.</p>
					</td></tr>
				</table>
			</td></tr>
		</table>
	</body>
	</html>`, verifyURL, s.Config.Redis.TTL)

	err = s.Mailer.SendEmail(ctx, req.Email, "Forgot Password Confirmations", emailLink)
	if err != nil {
		logger.Errorf("AuthServiceImpl.ForgotPasswordUser() sendMail ERROR, %v", err)
		return err
	}

	return nil
}

func (s *authServiceImpl) ResetPasswordUser(ctx context.Context, req *model.ResetPasswordRequest, code string) error {
	tr := s.Tracer.Tracer("Auth-ResetPasswordUser service")
	ctx, span := tr.Start(ctx, "Start ResetPasswordUser")
	defer span.End()

	if req.NewPassword == "" {
		return model.NewError(model.Validation, "password required")
	}

	ok := utils.IsValid(req.NewPassword)
	if !ok {
		return model.NewError(model.Validation, "password must min length 7, and at least has 1 each upper,lower,number,special")
	}

	getEmail, err := s.AuthRepo.GetVerificationByCode(ctx, code, model.ForgotPasswordVerification)
	if err != nil {
		if err == redis.Nil {
			return model.NewError(model.NotFound, "code not found / expired")
		}

		return err
	}

	hashedNewPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	err = s.AuthRepo.UpdatePasswordByEmail(ctx, getEmail, hashedNewPassword)
	if err != nil {
		return err
	}

	return nil
}

func validateRegisterUser(req *model.RegisterRequest) error {
	if req.FullName == "" {
		return model.NewError(model.Validation, "full name required")
	}

	if len(req.FullName) < 3 {
		return model.NewError(model.Validation, "full name must have at least 3 characters")
	}

	if req.Password == "" {
		return model.NewError(model.Validation, "password required")
	}

	ok := utils.IsValid(req.Password)
	if !ok {
		return model.NewError(model.Validation, "password must min length 7, and at least has 1 each upper,lower,number,special")
	}

	if !model.IsValidEmail.MatchString(req.Email) {
		return model.NewError(model.Validation, "invalid email")
	}

	return nil
}

func validateLoginUser(req *model.LoginRequest) error {
	if !model.IsValidEmail.MatchString(req.Email) {
		return model.NewError(model.Validation, "invalid email")
	}

	return nil
}
