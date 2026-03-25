package main

import (
	"fmt"
	"log"
	"net/http"
	stdhttp "net/http"

	"webpage-analyzer/internal/analyzer"
	"webpage-analyzer/internal/browser"
	"webpage-analyzer/internal/cache"
	"webpage-analyzer/internal/config"
	"webpage-analyzer/internal/handlers"
	httpclient "webpage-analyzer/internal/http"
	"webpage-analyzer/internal/logging"
	"webpage-analyzer/internal/services"
)

func main() {
	cfg, err := config.Load("config/app.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	errorLogger, err := buildLogger(cfg)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	httpClient := httpclient.NewDefaultHTTPClientWithConfig(
		httpclient.NewClientConfig(
			cfg.HTTPTimeout(),
			cfg.HTTPClient.MaxRedirects,
			cfg.HTTPClient.InsecureSkipVerify,
		),
	)

	var pageRenderer analyzer.PageRenderer
	if cfg.Browser.Enabled {
		pageRenderer = browser.NewChromeRendererWithConfig(cfg.Browser.Command, cfg.BrowserTimeout())
	}

	var resultCache cache.ResultCache
	if cfg.Cache.Enabled {
		resultCache = cache.NewRedisResultCache(
			cfg.Cache.Redis.Addr,
			cfg.Cache.Redis.Password,
			cfg.Cache.Redis.DB,
		)
	}

	analyzerService := services.NewAnalyzerService(
		analyzer.NewAnalyzer(httpClient, pageRenderer),
		resultCache,
		errorLogger,
		cfg.CacheTTL(),
	)

	handler, err := handlers.NewHandler(analyzerService, cfg.TemplatePaths.Index)
	if err != nil {
		log.Fatal("Failed to initialize handler:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/analyze", handler.Analyze)

	address := ":" + cfg.Server.Port

	log.Printf("Server starting on http://localhost:%s", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(address, mux))
}

func buildLogger(cfg *config.Config) (logging.Logger, error) {
	if !cfg.Logging.Enabled {
		return &logging.NoopLogger{}, nil
	}

	var sinks []logging.Logger

	for _, backend := range cfg.Logging.Backends {
		switch backend {
		case "file":
			logger, err := logging.NewFileLogger(cfg.Logging.File.Path)
			if err != nil {
				return nil, err
			}
			sinks = append(sinks, logger)
		case "db":
			logger, err := logging.NewSQLiteLogger(cfg.Logging.Database.DSN)
			if err != nil {
				return nil, err
			}
			sinks = append(sinks, logger)
		case "elasticsearch":
			sinks = append(sinks, logging.NewElasticsearchLogger(
				&stdhttp.Client{},
				cfg.Logging.Elasticsearch.URL,
				cfg.Logging.Elasticsearch.Index,
			))
		default:
			return nil, fmt.Errorf("unsupported logging backend: %s", backend)
		}
	}

	if len(sinks) == 0 {
		return &logging.NoopLogger{}, nil
	}

	return logging.NewMultiLogger(sinks...), nil
}
