package main

import (
	"auth-service/api/server"
	"auth-service/api/server/router/network"
	_ "auth-service/docs"
	"log"
)

// @title Auth Service API
// @version 1.0
// @description API for user rewards management system
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @schemes http https.
func main() {
	cfg, err := network.Load()
	if err != nil {
		log.Fatalf("Failed to load configs: %v", err)

		return
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to init server: %v", err)

		return
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server stopped: %v", err)

		return
	}
}
