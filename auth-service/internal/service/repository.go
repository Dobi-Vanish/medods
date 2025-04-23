package service

import (
	"auth-service/internal/postgres/repository"
	"net/http"
)

type RewardServiceInterface interface {
	RetrieveOne(w http.ResponseWriter, r *http.Request)
	GetLeaderboard(w http.ResponseWriter, r *http.Request)
	Authenticate(w http.ResponseWriter, r *http.Request)
	Registrate(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
}

type RewardService struct {
	RewardServiceInterface
	Repo   repository.Repository
	Client *http.Client
}
