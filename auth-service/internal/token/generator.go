package token

import (
	"auth-service/pkg/consts"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

type ServiceToken struct {
	SecretKey string
}

func NewTokenService() *ServiceToken {
	return &ServiceToken{
		SecretKey: os.Getenv("SECRET_KEY"),
	}
}

// GenerateTokens when called generates access tokens.
func (ts *ServiceToken) GenerateTokens(clientIP string) (string, string, error) {
	accessToken, err := ts.GenerateAccessToken(clientIP)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := GenerateRefreshToken(clientIP)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken generates access tokens.
func (ts *ServiceToken) GenerateAccessToken(clientIP string) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(consts.AccessTokenExpireTime).Unix(),
		"iat": time.Now().Unix(),
		"ip":  clientIP,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedToken, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", fmt.Errorf("failed to sign the token: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken generates refresh tokens.
func GenerateRefreshToken(clientIP string) (string, error) {
	tokenBytes := make([]byte, consts.RefreshTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	tokenData := fmt.Sprintf("%s|%s", clientIP, base64.URLEncoding.EncodeToString(tokenBytes))

	return tokenData, nil
}
