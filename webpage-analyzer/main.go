package main

import (
	"log"
	"net/http"

	"webpage-analyzer/internal/app"
	"webpage-analyzer/internal/config"
)

func main() {
	cfg, err := config.Load("config/app.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	runtimeApp, err := app.New(cfg)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	log.Printf("Server starting on http://localhost:%s", runtimeApp.Port)
	log.Fatal(http.ListenAndServe(runtimeApp.Address, runtimeApp.Handler))
}
