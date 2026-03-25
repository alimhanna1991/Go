package services

import (
	"context"
	"time"

	"webpage-analyzer/internal/cache"
	"webpage-analyzer/internal/logging"
	"webpage-analyzer/internal/models"
)

// AnalyzerService defines the service interface
type AnalyzerService interface {
	AnalyzeURL(url string) (*models.AnalysisResult, error)
}

// DefaultAnalyzerService implements AnalyzerService
type DefaultAnalyzerService struct {
	analyzer analyzerEngine
	cache    cache.ResultCache
	logger   logging.Logger
	cacheTTL time.Duration
}

type analyzerEngine interface {
	Analyze(url string) (*models.AnalysisResult, error)
}

// NewAnalyzerService creates a new analyzer service.
func NewAnalyzerService(engine analyzerEngine, resultCache cache.ResultCache, logger logging.Logger, cacheTTL time.Duration) *DefaultAnalyzerService {
	if logger == nil {
		logger = &logging.NoopLogger{}
	}

	return &DefaultAnalyzerService{
		analyzer: engine,
		cache:    resultCache,
		logger:   logger,
		cacheTTL: cacheTTL,
	}
}

// AnalyzeURL performs URL analysis
func (s *DefaultAnalyzerService) AnalyzeURL(url string) (*models.AnalysisResult, error) {
	ctx := context.Background()

	if s.cache != nil {
		cached, found, err := s.cache.Get(ctx, url)
		if err != nil {
			_ = s.logger.LogError(ctx, logging.Entry{
				Timestamp: time.Now().UTC(),
				Operation: "cache.get",
				URL:       url,
				Message:   err.Error(),
				Source:    "redis",
			})
		}
		if found {
			return cached, nil
		}
	}

	result, err := s.analyzer.Analyze(url)
	if err != nil {
		_ = s.logger.LogError(ctx, logging.Entry{
			Timestamp: time.Now().UTC(),
			Operation: "analyzer.analyze",
			URL:       url,
			Message:   err.Error(),
			Source:    "service",
		})
		return result, err
	}

	if s.cache != nil && result != nil {
		if err := s.cache.Set(ctx, url, result, s.cacheTTL); err != nil {
			_ = s.logger.LogError(ctx, logging.Entry{
				Timestamp: time.Now().UTC(),
				Operation: "cache.set",
				URL:       url,
				Message:   err.Error(),
				Source:    "redis",
			})
		}
	}

	return result, nil
}
