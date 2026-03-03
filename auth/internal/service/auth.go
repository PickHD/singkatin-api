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
		Context  context.Context
		Config   *config.Config
		Tracer   *trace.TracerProvider
		Mailer   *infrastructure.EmailProvider
		AuthRepo repository.AuthRepository
		JWT      *infrastructure.JwtProvider
	}
)

// NewAuthService return new instances auth service
func NewAuthService(ctx context.Context, config *config.Config, tracer *trace.TracerProvider, mailer *infrastructure.EmailProvider, authRepo repository.AuthRepository, jwt *infrastructure.JwtProvider) AuthService {
	return &authServiceImpl{
		Context:  ctx,
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

	emailLink := fmt.Sprintf("<h1><a href='%s'>%s</a><h1>", "http://localhost:8080/v1/register/verify?code="+codeVerification, "Verification Link")

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

	emailLink := fmt.Sprintf("<h1><a href='%s'>%s</a><h1>", "http://localhost:8080/v1/forgot-password/verify?code="+codeVerification, "Verification Link")

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
	if len(req.FullName) < 3 {
		return model.NewError(model.Validation, "full name must more than 3")
	}

	if req.FullName == "" {
		return model.NewError(model.Validation, "full name required")
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
