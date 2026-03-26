package app

import (
	"net/http"

	"webpage-analyzer/internal/api"
	"webpage-analyzer/internal/config"
	"webpage-analyzer/internal/handlers"
)

func newWebApp(cfg *config.Config) (*App, error) {
	serviceClient := api.NewAnalysisClient(cfg.AnalysisAPI.BaseURL, &http.Client{
		Timeout: cfg.AnalysisAPITimeout(),
	})

	handler, err := handlers.NewHandler(serviceClient, cfg.TemplatePaths.Index)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/analyze", handler.Analyze)

	return newRuntimeApp(cfg.Server.Port, mux), nil
}
