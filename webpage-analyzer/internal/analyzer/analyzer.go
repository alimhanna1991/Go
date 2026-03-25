package analyzer

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"webpage-analyzer/internal/models"
)

// Analyzer is the main analyzer facade
type Analyzer struct {
	htmlVersionDetector *HTMLVersionDetector
	titleExtractor      *TitleExtractor
	headingExtractor    *HeadingExtractor
	loginFormDetector   *LoginFormDetector
	linkAnalyzer        *LinkAnalyzer
	httpClient          HTTPClient
	pageRenderer        PageRenderer
}

// NewAnalyzer creates a new analyzer with all dependencies
func NewAnalyzer(httpClient HTTPClient, pageRenderer PageRenderer) *Analyzer {
	return &Analyzer{
		htmlVersionDetector: NewHTMLVersionDetector(),
		titleExtractor:      NewTitleExtractor(),
		headingExtractor:    NewHeadingExtractor(),
		loginFormDetector:   NewLoginFormDetector(),
		linkAnalyzer:        NewLinkAnalyzer(httpClient),
		httpClient:          httpClient,
		pageRenderer:        pageRenderer,
	}
}

// Analyze performs complete analysis of a webpage
func (a *Analyzer) Analyze(targetURL string) (*models.AnalysisResult, error) {
	result := &models.AnalysisResult{
		URL:      targetURL,
		Headings: make(map[string]int),
	}

	normalizedURL, err := a.normalizeURL(targetURL)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result, err
	}
	result.URL = normalizedURL

	resp, body, err := a.httpClient.Fetch(normalizedURL)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to fetch URL: %v", err)
		return result, err
	}

	if resp != nil {
		result.StatusCode = resp.StatusCode
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		result.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if body == nil {
		return result, nil
	}
	defer body.Close()

	doc, err := html.Parse(body)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to parse HTML: %v", err)
		return result, err
	}

	contentType := ""
	if resp != nil {
		contentType = resp.Header.Get("Content-Type")
	}
	result.HTMLVersion = a.htmlVersionDetector.Detect(doc, contentType)
	result.PageTitle = a.titleExtractor.Extract(doc)
	result.Headings = a.headingExtractor.Extract(doc)
	result.HasLoginForm = a.loginFormDetector.Detect(doc, normalizedURL, result.PageTitle)

	if !result.HasLoginForm {
		if renderedDoc, err := a.renderDocument(normalizedURL); err == nil && renderedDoc != nil {
			result.HasLoginForm = a.loginFormDetector.Detect(renderedDoc, normalizedURL, result.PageTitle)
		}
	}

	baseURL := a.getBaseURL(normalizedURL)
	internal, external, links := a.linkAnalyzer.Analyze(doc, baseURL)
	result.InternalLinks = internal
	result.ExternalLinks = external
	result.Links = links
	result.InaccessibleLinks = a.linkAnalyzer.CheckAccessibility(result.Links)

	return result, nil
}

func (a *Analyzer) normalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %v", err)
	}

	if parsedURL.Host == "" && parsedURL.Scheme == "" && parsedURL.Path != "" {
		if !looksLikeHost(parsedURL.Path) {
			return "", fmt.Errorf("invalid URL format: missing host")
		}
		parsedURL, err = url.Parse("http://" + rawURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %v", err)
		}
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("invalid URL format: missing host")
	}

	return parsedURL.String(), nil
}

func (a *Analyzer) getBaseURL(rawURL string) string {
	parsedURL, _ := url.Parse(rawURL)
	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}

func looksLikeHost(value string) bool {
	if value == "localhost" {
		return true
	}

	if strings.Contains(value, ".") {
		return true
	}

	return false
}

func (a *Analyzer) renderDocument(targetURL string) (*html.Node, error) {
	if a.pageRenderer == nil {
		return nil, nil
	}

	renderedHTML, err := a.pageRenderer.RenderHTML(targetURL)
	if err != nil || renderedHTML == "" {
		return nil, err
	}

	return html.Parse(strings.NewReader(renderedHTML))
}
