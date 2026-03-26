package api

import (
	"encoding/json"
	"net/http"

	"webpage-analyzer/internal/models"
	"webpage-analyzer/internal/services"
)

// AnalysisHandler exposes the analysis use case over HTTP.
type AnalysisHandler struct {
	service services.AnalyzerService
}

// NewAnalysisHandler creates a new API handler.
func NewAnalysisHandler(service services.AnalyzerService) *AnalysisHandler {
	return &AnalysisHandler{service: service}
}

// Analyze accepts analysis requests and returns JSON results.
func (h *AnalysisHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, &models.AnalysisResult{
			ErrorMessage: "invalid JSON payload",
		})
		return
	}

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, &models.AnalysisResult{
			ErrorMessage: "url is required",
		})
		return
	}

	result, err := h.service.AnalyzeURL(r.Context(), req.URL)
	if err != nil && (result == nil || result.ErrorMessage == "") {
		result = &models.AnalysisResult{
			URL:          req.URL,
			ErrorMessage: err.Error(),
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// Health returns a minimal readiness response.
func (h *AnalysisHandler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
