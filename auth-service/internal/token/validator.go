package token

import (
	"auth-service/pkg/errormsg"
	"fmt"
	"github.com/golang-jwt/jwt"
	"strings"
	"time"
)

type Validator struct {
	SecretKey string
}

// ValidateAccessToken validate provided access token.
func (ts *ServiceToken) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	tokenString = strings.TrimSpace(tokenString)
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, fmt.Errorf("%w: %v", errormsg.ErrUnexpectedSigningMethod, t.Header["alg"])
		}

		return []byte(ts.SecretKey), nil
	})

	if err != nil {
		return nil, errormsg.ErrTokenValidation
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errormsg.ErrInvalidToken
	}

	now := time.Now().Unix()

	expVal, ok := claims["exp"].(float64)
	if !ok {
		return nil, errormsg.ErrInvalidToken
	}

	exp := int64(expVal)

	iatVal, ok := claims["iat"].(float64)
	if !ok {
		return nil, errormsg.ErrInvalidToken
	}

	iat := int64(iatVal)

	fmt.Printf("Current: %d, Issued: %d, Expires: %d\n", now, iat, exp)

	if !token.Valid {
		return nil, errormsg.ErrInvalidToken
	}

	return claims, nil
}
