package middleware

import (
	"auth-service/internal/token"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Auth middleware checks JWT token from cookies.
func Auth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessCookie, err := r.Cookie("accessToken")
			if err != nil {
				handleAuthError(w, "missing access token")

				return
			}

			ip := r.RemoteAddr
			if strings.Contains(ip, ":") {
				ip = strings.Split(ip, ":")[0]
			}

			tokenService := token.NewTokenService()

			claims, err := tokenService.ValidateAccessToken(accessCookie.Value)
			if err != nil {
				fmt.Println("Validation error details:", err)
				handleAuthError(w, "invalid access token: "+err.Error())

				return
			}

			fmt.Println("Successful validation, claims:", claims)

			ctx := context.WithValue(r.Context(), "clientIP", ip) //nolint: revive,staticcheck
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// handleAuthError handle errors from Auth middleware.
func handleAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   true,
		"message": "Authentication failed: " + message,
	})

	if err != nil {
		return
	}
}
