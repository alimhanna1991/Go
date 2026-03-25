package handlers

import (
	"bytes"
	"html/template"
	"net/http"

	"webpage-analyzer/internal/models"
	"webpage-analyzer/internal/services"
)

// Handler handles HTTP requests
type Handler struct {
	analyzerService services.AnalyzerService
	template        *template.Template
}

// NewHandler creates a new handler with dependencies
func NewHandler(analyzerService services.AnalyzerService) (*Handler, error) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		return nil, err
	}

	return NewHandlerWithTemplate(analyzerService, tmpl), nil
}

// NewHandlerWithTemplate creates a handler with an injected template for testability.
func NewHandlerWithTemplate(analyzerService services.AnalyzerService, tmpl *template.Template) *Handler {
	return &Handler{
		analyzerService: analyzerService,
		template:        tmpl,
	}
}

// Home handles the home page
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	h.render(w, newPageData("", nil))
}

// Analyze handles the analyze request
func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.render(w, newPageData("", &models.AnalysisResult{
			ErrorMessage: "Failed to parse form submission",
		}))
		return
	}

	targetURL := r.FormValue("url")
	if targetURL == "" {
		h.render(w, newPageData("", &models.AnalysisResult{
			ErrorMessage: "Please provide a URL",
		}))
		return
	}

	result, err := h.analyzerService.AnalyzeURL(targetURL)
	if err != nil && (result == nil || result.ErrorMessage == "") {
		result = &models.AnalysisResult{
			URL:          targetURL,
			ErrorMessage: err.Error(),
		}
	}

	h.render(w, newPageData(targetURL, result))
}

func (h *Handler) render(w http.ResponseWriter, data *PageData) {
	if h.template != nil {
		if err := h.template.Execute(w, data); err == nil {
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var buf bytes.Buffer
	if data.Error != "" {
		buf.WriteString(data.Error)
	} else if data.Result != nil {
		buf.WriteString(data.Result.PageTitle)
		if data.Result.PageTitle == "" {
			buf.WriteString(data.Result.URL)
		}
	}
	_, _ = w.Write(buf.Bytes())
}

// PageData is the template view model for the home and result pages.
type PageData struct {
	FormURL string
	Result  *models.AnalysisResult
	Error   string
}

func newPageData(formURL string, result *models.AnalysisResult) *PageData {
	data := &PageData{
		FormURL: formURL,
		Result:  result,
	}

	if result != nil {
		data.Error = result.ErrorMessage
		if data.FormURL == "" {
			data.FormURL = result.URL
		}
	}

	return data
}
