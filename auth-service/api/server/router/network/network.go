package network

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	"auth-service/api/server/middleware"
	"auth-service/internal/service"
)

// SetupRoutes set up the Routes
// @BasePath.
func SetupRoutes(svc *service.RewardService) http.Handler {
	r := chi.NewRouter()

	r.Group(func(secure chi.Router) {
		secure.Use(middleware.Auth())

		secure.Get("/users/{id}/status", svc.RetrieveOne)
		secure.Get("/users/leaderboard", svc.GetLeaderboard)
		secure.Get("/refresh/{id}", svc.Refresh)
	})

	r.Post("/authenticate", svc.Authenticate)
	r.Post("/registrate", svc.Registrate)
	r.Get("/provide/{id}", svc.Provide)

	return r
}
