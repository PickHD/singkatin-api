package infrastructure

import (
	"errors"
	"fmt"
	"singkatin-api/auth/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtProvider struct {
	secretKey      string
	issuer         string
	expireDuration time.Duration
}

func NewJWTProvider(cfg *config.Config) *JwtProvider {
	return &JwtProvider{
		secretKey:      cfg.Secret.JWTSecret,
		issuer:         cfg.Server.AppName,
		expireDuration: time.Hour * time.Duration(cfg.Common.JWTExpire),
	}
}

type MyClaims struct {
	UserID   string `json:"user_id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func (p *JwtProvider) GenerateToken(userID string, fullName string, email string) (string, error) {
	claims := &MyClaims{
		UserID:   userID,
		FullName: fullName,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.expireDuration)),
			Issuer:    p.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.secretKey))
}

func (p *JwtProvider) ValidateToken(tokenString string) (*MyClaims, error) {
	claims := &MyClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
