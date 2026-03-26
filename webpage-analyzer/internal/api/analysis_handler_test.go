package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"webpage-analyzer/internal/models"
)

type stubAnalyzerService struct {
	result *models.AnalysisResult
	err    error
}

func (s *stubAnalyzerService) AnalyzeURL(ctx context.Context, url string) (*models.AnalysisResult, error) {
	if s.result != nil {
		return s.result, s.err
	}
	return &models.AnalysisResult{URL: url, PageTitle: "Remote"}, s.err
}

func TestAnalysisHandler_Analyze(t *testing.T) {
	handler := NewAnalysisHandler(&stubAnalyzerService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewBufferString(`{"url":"https://example.com"}`))
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result models.AnalysisResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if result.PageTitle != "Remote" {
		t.Fatalf("expected remote title, got %+v", result)
	}
}

func TestAnalysisHandler_Analyze_RejectsMissingURL(t *testing.T) {
	handler := NewAnalysisHandler(&stubAnalyzerService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewBufferString(`{"url":""}`))
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAnalysisHandler_Analyze_UsesErrorResultPayload(t *testing.T) {
	handler := NewAnalysisHandler(&stubAnalyzerService{
		result: &models.AnalysisResult{
			URL:          "https://example.com",
			ErrorMessage: "analysis failed",
		},
		err: errors.New("analysis failed"),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewBufferString(`{"url":"https://example.com"}`))
	w := httptest.NewRecorder()

	handler.Analyze(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result models.AnalysisResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if result.ErrorMessage != "analysis failed" {
		t.Fatalf("expected propagated error payload, got %+v", result)
	}
}
