package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PickHD/singkatin-revamp/auth/internal/config"
	"github.com/PickHD/singkatin-revamp/auth/internal/helper"
	"github.com/PickHD/singkatin-revamp/auth/internal/model"
	"github.com/PickHD/singkatin-revamp/auth/internal/repository"
	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
	"gopkg.in/gomail.v2"
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

	// AuthServiceImpl is an app auth struct that consists of all the dependencies needed for auth service
	AuthServiceImpl struct {
		Context  context.Context
		Config   *config.Configuration
		Logger   *logrus.Logger
		Tracer   *trace.TracerProvider
		Mailer   *gomail.Dialer
		AuthRepo repository.AuthRepository
	}
)

// NewAuthService return new instances auth service
func NewAuthService(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, mailer *gomail.Dialer, authRepo repository.AuthRepository) *AuthServiceImpl {
	return &AuthServiceImpl{
		Context:  ctx,
		Config:   config,
		Logger:   logger,
		Tracer:   tracer,
		Mailer:   mailer,
		AuthRepo: authRepo,
	}
}

func (as *AuthServiceImpl) RegisterUser(ctx context.Context, req *model.RegisterRequest) (*model.RegisterResponse, error) {
	tr := as.Tracer.Tracer("Auth-RegisterUser service")
	ctx, span := tr.Start(ctx, "Start RegisterUser")
	defer span.End()

	err := validateRegisterUser(req)
	if err != nil {
		return nil, err
	}

	hashPass, err := helper.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	data, err := as.AuthRepo.CreateUser(ctx, &model.User{
		FullName:   req.FullName,
		Email:      req.Email,
		Password:   hashPass,
		IsVerified: false,
	})
	if err != nil {
		return nil, err
	}

	codeVerification := helper.RandomStringBytesMaskImprSrcSB(9)
	expiredCodeDuration := time.Minute * time.Duration(as.Config.Redis.TTL)

	err = as.AuthRepo.SetVerificationByEmail(ctx, req.Email, codeVerification, expiredCodeDuration, model.RegisterVerification)
	if err != nil {
		return nil, err
	}

	emailLink := fmt.Sprintf("<h1><a href='%s'>%s</a><h1>", "http://localhost:8080/v1/register/verify?code="+codeVerification, "Verification Link")

	err = as.sendMail(as.Config.Mailer.Sender, []string{req.Email}, req.Email, "Registration Confirmations", "Please Complete the Verification of your Request Registration", emailLink)
	if err != nil {
		as.Logger.Error(err)
		return nil, err
	}

	return &model.RegisterResponse{
		ID:         data.ID.Hex(),
		Email:      data.Email,
		IsVerified: false,
	}, nil
}

func (as *AuthServiceImpl) LoginUser(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	tr := as.Tracer.Tracer("Auth-LoginUser service")
	ctx, span := tr.Start(ctx, "Start LoginUser")
	defer span.End()

	err := validateLoginUser(req)
	if err != nil {
		return nil, err
	}

	user, err := as.AuthRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// verify user password by comparing incoming request password with crypted password stored in database
	if !helper.CheckPasswordHash(user.Password, req.Password) {
		return nil, model.NewError(model.Validation, "invalid password")
	}

	// generate access token jwt
	token, expiredAt, err := as.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken: token,
		Type:        "Bearer",
		ExpireAt:    time.Now().Add(expiredAt),
	}, nil
}

func (as *AuthServiceImpl) VerifyCode(ctx context.Context, code string, verificationType model.VerificationType) (*model.VerifyCodeResponse, error) {
	tr := as.Tracer.Tracer("Auth-VerifyCode service")
	ctx, span := tr.Start(ctx, "Start VerifyCode")
	defer span.End()

	getEmail, err := as.AuthRepo.GetVerificationByCode(ctx, code, verificationType)
	if err != nil {
		if err == redis.Nil {
			return nil, model.NewError(model.NotFound, "code not found / expired")
		}

		return nil, err
	}

	switch verificationType {
	case model.RegisterVerification:
		err = as.AuthRepo.UpdateVerifyStatusByEmail(ctx, getEmail)
		if err != nil {
			return nil, err
		}
	case model.ForgotPasswordVerification:
	}

	return &model.VerifyCodeResponse{
		IsVerified: true,
	}, nil
}

func (as *AuthServiceImpl) ForgotPasswordUser(ctx context.Context, req *model.ForgotPasswordRequest) error {
	tr := as.Tracer.Tracer("Auth-ForgotPasswordUser service")
	ctx, span := tr.Start(ctx, "Start ForgotPasswordUser")
	defer span.End()

	if !model.IsValidEmail.MatchString(req.Email) {
		return model.NewError(model.Validation, "invalid email")
	}

	_, err := as.AuthRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	codeVerification := helper.RandomStringBytesMaskImprSrcSB(10)
	expiredCodeDuration := time.Minute * time.Duration(as.Config.Redis.TTL)

	err = as.AuthRepo.SetVerificationByEmail(ctx, req.Email, codeVerification, expiredCodeDuration, model.ForgotPasswordVerification)
	if err != nil {
		return err
	}

	emailLink := fmt.Sprintf("<h1><a href='%s'>%s</a><h1>", "http://localhost:8080/v1/forgot-password/verify?code="+codeVerification, "Verification Link")

	err = as.sendMail(as.Config.Mailer.Sender, []string{req.Email}, req.Email, "Forgot Password Confirmations", "Please Complete the Verification of your Request Forgot Password", emailLink)
	if err != nil {
		as.Logger.Error(err)
		return err
	}

	return nil
}

func (as *AuthServiceImpl) ResetPasswordUser(ctx context.Context, req *model.ResetPasswordRequest, code string) error {
	tr := as.Tracer.Tracer("Auth-ResetPasswordUser service")
	ctx, span := tr.Start(ctx, "Start ResetPasswordUser")
	defer span.End()

	if req.NewPassword == "" {
		return model.NewError(model.Validation, "password required")
	}

	ok := helper.IsValid(req.NewPassword)
	if !ok {
		return model.NewError(model.Validation, "password must min length 7, and at least has 1 each upper,lower,number,special")
	}

	getEmail, err := as.AuthRepo.GetVerificationByCode(ctx, code, model.ForgotPasswordVerification)
	if err != nil {
		if err == redis.Nil {
			return model.NewError(model.NotFound, "code not found / expired")
		}

		return err
	}

	hashedNewPassword, err := helper.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	err = as.AuthRepo.UpdatePasswordByEmail(ctx, getEmail, hashedNewPassword)
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

	ok := helper.IsValid(req.Password)
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

func (as *AuthServiceImpl) generateJWT(user *model.User) (string, time.Duration, error) {
	var (
		payloadUserID   = "user_id"
		payloadFullName = "full_name"
		payloadEmail    = "email"
		payloadExpires  = "exp"
		JWTExpire       = time.Duration(as.Config.Common.JWTExpire) * time.Hour
	)

	claims := jwt.MapClaims{}
	claims[payloadUserID] = user.ID.Hex()
	claims[payloadFullName] = user.FullName
	claims[payloadEmail] = user.Email
	claims[payloadExpires] = time.Now().Add(JWTExpire).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(as.Config.Secret.JWTSecret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, JWTExpire, nil
}

func (as *AuthServiceImpl) sendMail(from string, to []string, cc string, ccTitle string, subject string, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", from)
	mailer.SetHeader("To", to...)
	mailer.SetAddressHeader("Cc", cc, ccTitle)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	err := as.Mailer.DialAndSend(mailer)
	if err != nil {
		return err
	}

	return nil
}
