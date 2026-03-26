package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"webpage-analyzer/internal/logging"
	"webpage-analyzer/internal/models"
)

type stubAnalyzer struct {
	result *models.AnalysisResult
	err    error
	calls  int
}

func (a *stubAnalyzer) Analyze(url string) (*models.AnalysisResult, error) {
	a.calls++
	return a.result, a.err
}

type stubCache struct {
	result   *models.AnalysisResult
	found    bool
	getErr   error
	setErr   error
	setCalls int
}

func (c *stubCache) Get(ctx context.Context, url string) (*models.AnalysisResult, bool, error) {
	return c.result, c.found, c.getErr
}

func (c *stubCache) Set(ctx context.Context, url string, result *models.AnalysisResult, ttl time.Duration) error {
	c.setCalls++
	return c.setErr
}

type stubLogger struct {
	entries []logging.Entry
}

func (l *stubLogger) LogError(ctx context.Context, entry logging.Entry) error {
	l.entries = append(l.entries, entry)
	return nil
}

func TestAnalyzerService_UsesCacheHit(t *testing.T) {
	expected := &models.AnalysisResult{URL: "https://example.com", PageTitle: "Cached"}
	cache := &stubCache{result: expected, found: true}
	analyzer := &stubAnalyzer{}

	service := NewAnalyzerService(analyzer, cache, &stubLogger{}, 5*time.Minute)
	result, err := service.AnalyzeURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("AnalyzeURL() returned error: %v", err)
	}
	if result.PageTitle != "Cached" {
		t.Fatalf("expected cached result, got %+v", result)
	}
	if analyzer.calls != 0 {
		t.Fatalf("expected analyzer not to run on cache hit, got %d calls", analyzer.calls)
	}
}

func TestAnalyzerService_CachesAnalyzerResult(t *testing.T) {
	analyzer := &stubAnalyzer{
		result: &models.AnalysisResult{URL: "https://example.com", PageTitle: "Fresh"},
	}
	cache := &stubCache{}

	service := NewAnalyzerService(analyzer, cache, &stubLogger{}, 5*time.Minute)
	result, err := service.AnalyzeURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("AnalyzeURL() returned error: %v", err)
	}
	if result.PageTitle != "Fresh" {
		t.Fatalf("expected analyzer result, got %+v", result)
	}
	if analyzer.calls != 1 {
		t.Fatalf("expected analyzer to run once, got %d calls", analyzer.calls)
	}
	if cache.setCalls != 1 {
		t.Fatalf("expected cache set once, got %d", cache.setCalls)
	}
}

func TestAnalyzerService_LogsAnalyzerErrors(t *testing.T) {
	expectedErr := errors.New("analysis failed")
	analyzer := &stubAnalyzer{err: expectedErr}
	logger := &stubLogger{}

	service := NewAnalyzerService(analyzer, nil, logger, 5*time.Minute)
	_, err := service.AnalyzeURL(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logger.entries))
	}
	if logger.entries[0].Message != expectedErr.Error() {
		t.Fatalf("expected log message %q, got %q", expectedErr.Error(), logger.entries[0].Message)
	}
}

func TestAnalyzerService_UsesProvidedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cache := &stubCache{getErr: context.Canceled}
	logger := &stubLogger{}
	service := NewAnalyzerService(&stubAnalyzer{result: &models.AnalysisResult{}}, cache, logger, 5*time.Minute)

	_, err := service.AnalyzeURL(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("expected cache cancellation to be logged but not returned, got %v", err)
	}
	if len(logger.entries) == 0 || logger.entries[0].Message != context.Canceled.Error() {
		t.Fatalf("expected cancellation log entry, got %+v", logger.entries)
	}
}
