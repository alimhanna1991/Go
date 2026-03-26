package app

import (
	"fmt"
	stdhttp "net/http"

	"webpage-analyzer/internal/analyzer"
	"webpage-analyzer/internal/browser"
	"webpage-analyzer/internal/cache"
	"webpage-analyzer/internal/config"
	httpclient "webpage-analyzer/internal/http"
	"webpage-analyzer/internal/logging"
)

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
