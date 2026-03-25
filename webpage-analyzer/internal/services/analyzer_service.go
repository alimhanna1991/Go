package services

import (
	"webpage-analyzer/internal/analyzer"
	"webpage-analyzer/internal/browser"
	"webpage-analyzer/internal/http"
	"webpage-analyzer/internal/models"
)

// AnalyzerService defines the service interface
type AnalyzerService interface {
	AnalyzeURL(url string) (*models.AnalysisResult, error)
}

// DefaultAnalyzerService implements AnalyzerService
type DefaultAnalyzerService struct {
	analyzer *analyzer.Analyzer
}

// NewAnalyzerService creates a new analyzer service
func NewAnalyzerService() *DefaultAnalyzerService {
	httpClient := http.NewDefaultHTTPClient()
	pageRenderer := browser.NewChromeRenderer()
	return &DefaultAnalyzerService{
		analyzer: analyzer.NewAnalyzer(httpClient, pageRenderer),
	}
}

// AnalyzeURL performs URL analysis
func (s *DefaultAnalyzerService) AnalyzeURL(url string) (*models.AnalysisResult, error) {
	return s.analyzer.Analyze(url)
}
