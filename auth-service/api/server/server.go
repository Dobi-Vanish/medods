package server

import (
	"auth-service/api/server/router/network"
	"auth-service/internal/postgres/models"
	"auth-service/internal/service"
	"auth-service/migrations"
	"auth-service/pkg/consts"
	"auth-service/pkg/db"
	"auth-service/pkg/errormsg"
	"fmt"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"time"
)

type Server struct {
	cfg    *network.Config
	router *chi.Mux
}

func NewServer(cfg *network.Config) (*Server, error) {
	conn, err := db.Connect(cfg.DB.DSN)
	if err != nil {
		return nil, errormsg.ErrConnectDB
	}

	if err := migrations.Apply(conn); err != nil {
		return nil, errormsg.ErrApplyMigrations
	}

	repo := models.NewPostgresRepository(conn)
	svc := service.NewRewardService(repo)

	router := chi.NewRouter()
	router.Use(network.CORS())
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	handler := network.SetupRoutes(svc)
	router.Mount("/", handler)

	return &Server{
		cfg:    cfg,
		router: router,
	}, nil
}

func (s *Server) Start() error {
	server := &http.Server{
		Addr:         ":" + s.cfg.Server.Port,
		Handler:      s.router,
		ReadTimeout:  consts.ReadTimeout * time.Second,
		WriteTimeout: consts.WriteTimeout * time.Second,
		IdleTimeout:  consts.IdleTimeout * time.Second,
	}

	log.Printf("Server started on :%s", s.cfg.Server.Port)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}
