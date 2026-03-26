package app

import (
	"net/http"

	"webpage-analyzer/internal/analyzer"
	"webpage-analyzer/internal/api"
	"webpage-analyzer/internal/config"
	"webpage-analyzer/internal/services"
)

func newAnalysisApp(cfg *config.Config) (*App, error) {
	errorLogger, err := buildLogger(cfg)
	if err != nil {
		return nil, err
	}

	analyzerService := services.NewAnalyzerService(
		analyzer.NewAnalyzer(newHTTPClient(cfg), newPageRenderer(cfg)),
		newResultCache(cfg),
		errorLogger,
		cfg.CacheTTL(),
	)

	apiHandler := api.NewAnalysisHandler(analyzerService)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/analyze", apiHandler.Analyze)
	mux.HandleFunc("/api/v1/health", apiHandler.Health)

	return newRuntimeApp(cfg.Server.Port, mux), nil
}
