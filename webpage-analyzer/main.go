package main

import (
	"log"
	"net/http"
	"os"

	"webpage-analyzer/internal/handlers"
	"webpage-analyzer/internal/services"
)

func main() {
	analyzerService := services.NewAnalyzerService()
	handler, err := handlers.NewHandler(analyzerService)
	if err != nil {
		log.Fatal("Failed to initialize handler:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/analyze", handler.Analyze)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := ":" + port

	log.Printf("Server starting on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(address, mux))
}
