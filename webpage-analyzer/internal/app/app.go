package app

import (
	"fmt"
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

// App contains the runtime HTTP dependencies for the application.
type App struct {
	Handler http.Handler
	Address string
	Port    string
}

// New builds the application from runtime configuration.
func New(cfg *config.Config) (*App, error) {
	errorLogger, err := buildLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize logger: %w", err)
	}

	handler, err := buildHandler(cfg, errorLogger)
	if err != nil {
		return nil, fmt.Errorf("initialize handler: %w", err)
	}

	return &App{
		Handler: handler,
		Address: ":" + cfg.Server.Port,
		Port:    cfg.Server.Port,
	}, nil
}

func buildHandler(cfg *config.Config, errorLogger logging.Logger) (http.Handler, error) {
	httpClient := httpclient.NewDefaultHTTPClientWithConfig(
		httpclient.NewClientConfig(
			cfg.HTTPTimeout(),
			cfg.HTTPClient.MaxRedirects,
			cfg.HTTPClient.InsecureSkipVerify,
		),
	)

	analyzerService := services.NewAnalyzerService(
		analyzer.NewAnalyzer(httpClient, newPageRenderer(cfg)),
		newResultCache(cfg),
		errorLogger,
		cfg.CacheTTL(),
	)

	handler, err := handlers.NewHandler(analyzerService, cfg.TemplatePaths.Index)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/analyze", handler.Analyze)

	return mux, nil
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
