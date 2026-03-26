package app

import (
	"fmt"
	"net/http"
	stdhttp "net/http"

	"webpage-analyzer/internal/analyzer"
	"webpage-analyzer/internal/api"
	"webpage-analyzer/internal/browser"
	"webpage-analyzer/internal/cache"
	"webpage-analyzer/internal/config"
	"webpage-analyzer/internal/handlers"
	httpclient "webpage-analyzer/internal/http"
	"webpage-analyzer/internal/logging"
	"webpage-analyzer/internal/services"
)

// App contains the runtime HTTP dependencies for the application.
type App struct {
	Handler http.Handler
	Address string
	Port    string
}

// New builds the application from runtime configuration.
func New(cfg *config.Config) (*App, error) {
	switch cfg.Service.Role {
	case "web":
		return newWebApp(cfg)
	case "analysis":
		return newAnalysisApp(cfg)
	default:
		return nil, fmt.Errorf("unsupported service role: %s", cfg.Service.Role)
	}
}

func newWebApp(cfg *config.Config) (*App, error) {
	serviceClient := api.NewAnalysisClient(cfg.AnalysisAPI.BaseURL, &http.Client{
		Timeout: cfg.AnalysisAPITimeout(),
	})

	handler, err := handlers.NewHandler(serviceClient, cfg.TemplatePaths.Index)
	if err != nil {
		return nil, fmt.Errorf("initialize handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/analyze", handler.Analyze)

	return newRuntimeApp(cfg.Server.Port, mux), nil
}

func newAnalysisApp(cfg *config.Config) (*App, error) {
	errorLogger, err := buildLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize logger: %w", err)
	}

	mux := http.NewServeMux()
	analyzerService := services.NewAnalyzerService(
		analyzer.NewAnalyzer(newHTTPClient(cfg), newPageRenderer(cfg)),
		newResultCache(cfg),
		errorLogger,
		cfg.CacheTTL(),
	)
	apiHandler := api.NewAnalysisHandler(analyzerService)
	mux.HandleFunc("/api/v1/analyze", apiHandler.Analyze)
	mux.HandleFunc("/api/v1/health", apiHandler.Health)

	return newRuntimeApp(cfg.Server.Port, mux), nil
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

func newRuntimeApp(port string, handler http.Handler) *App {
	return &App{
		Handler: handler,
		Address: ":" + port,
		Port:    port,
	}
}

func newHTTPClient(cfg *config.Config) analyzer.HTTPClient {
	return httpclient.NewDefaultHTTPClientWithConfig(
		httpclient.NewClientConfig(
			cfg.HTTPTimeout(),
			cfg.HTTPClient.MaxRedirects,
			cfg.HTTPClient.InsecureSkipVerify,
		),
	)
}

func newPageRenderer(cfg *config.Config) analyzer.PageRenderer {
	if !cfg.Browser.Enabled {
		return nil
	}

	return browser.NewChromeRendererWithConfig(cfg.Browser.Command, cfg.BrowserTimeout())
}

func newResultCache(cfg *config.Config) cache.ResultCache {
	if !cfg.Cache.Enabled {
		return nil
	}

	return cache.NewRedisResultCache(
		cfg.Cache.Redis.Addr,
		cfg.Cache.Redis.Password,
		cfg.Cache.Redis.DB,
	)
}
