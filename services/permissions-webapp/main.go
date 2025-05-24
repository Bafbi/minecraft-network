package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bafbi/minecraft-network/services/permissions-webapp/internal/config"
	"github.com/bafbi/minecraft-network/services/permissions-webapp/internal/core"
	"github.com/bafbi/minecraft-network/services/permissions-webapp/internal/web"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	if err := core.InitNats(cfg.NatsURL); err != nil {
		log.Fatalf("Error initializing NATS: %v", err)
	}
	defer core.CloseNats()

	if err := core.InitCasbin(cfg); err != nil { // Pass full cfg
		log.Fatalf("Error initializing Casbin: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)    // Basic request logging
	r.Use(middleware.Recoverer) // Recover from panics

	web.RegisterRoutes(r) // Pass logger if handlers need it

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001" // Default port for the webapp
	}

	log.Printf("Permissions webapp starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
