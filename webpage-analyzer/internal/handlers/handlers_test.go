package handlers

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"webpage-analyzer/internal/models"
)

type MockAnalyzerService struct {
	ShouldFail bool
	Result     *models.AnalysisResult
}

func (m *MockAnalyzerService) AnalyzeURL(ctx context.Context, url string) (*models.AnalysisResult, error) {
	if m.ShouldFail {
		return &models.AnalysisResult{
			URL:          url,
			ErrorMessage: "Mock error",
		}, nil
	}

	if m.Result != nil {
		return m.Result, nil
	}

	return &models.AnalysisResult{
		URL:       url,
		PageTitle: "Test Page",
	}, nil
}

func TestHandler_Analyze(t *testing.T) {
	tests := []struct {
		name           string
		formData       string
		mockService    *MockAnalyzerService
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid URL submission",
			formData:       "url=https://example.com",
			mockService:    &MockAnalyzerService{ShouldFail: false},
			expectedStatus: http.StatusOK,
			expectedBody:   "Test Page",
		},
		{
			name:           "Empty URL",
			formData:       "url=",
			mockService:    &MockAnalyzerService{ShouldFail: false},
			expectedStatus: http.StatusOK,
			expectedBody:   "Please provide a URL",
		},
		{
			name:           "Invalid URL",
			formData:       "url=invalid",
			mockService:    &MockAnalyzerService{ShouldFail: true},
			expectedStatus: http.StatusOK,
			expectedBody:   "Mock error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := template.Must(template.New("test").Parse(`{{if .Error}}{{.Error}}{{else if .Result}}{{.Result.PageTitle}}{{end}}`))
			handler := NewHandlerWithTemplate(tt.mockService, tmpl)

			req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(tt.formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			handler.Analyze(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
