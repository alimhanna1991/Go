package analyzer

import (
	"io"
	"net/http"

	"webpage-analyzer/internal/models"
)

// HTTPClient abstracts HTTP fetches for the analyzer package.
type HTTPClient interface {
	Fetch(url string) (*http.Response, io.ReadCloser, error)
	Check(url string) (*http.Response, error)
}

// PageRenderer renders client-side HTML when raw server responses are insufficient.
type PageRenderer interface {
	RenderHTML(url string) (string, error)
}

// Type aliases keep analyzer package tests and callers simple.
type AnalysisResult = models.AnalysisResult
type LinkInfo = models.LinkInfo
