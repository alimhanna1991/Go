package analyzer

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

// Analyzer is the main analyzer facade
type Analyzer struct {
	htmlVersionDetector *HTMLVersionDetector
	titleExtractor      *TitleExtractor
	headingExtractor    *HeadingExtractor
	loginFormDetector   *LoginFormDetector
	linkAnalyzer        *LinkAnalyzer
	httpClient          HTTPClient
}

// NewAnalyzer creates a new analyzer with all dependencies
func NewAnalyzer(httpClient HTTPClient) *Analyzer {
	return &Analyzer{
		htmlVersionDetector: NewHTMLVersionDetector(),
		titleExtractor:      NewTitleExtractor(),
		headingExtractor:    NewHeadingExtractor(),
		loginFormDetector:   NewLoginFormDetector(),
		linkAnalyzer:        NewLinkAnalyzer(httpClient),
		httpClient:          httpClient,
	}
}

// Analyze performs complete analysis of a webpage
func (a *Analyzer) Analyze(targetURL string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		URL:      targetURL,
		Headings: make(map[string]int),
	}

	// Validate and normalize URL
	normalizedURL, err := a.normalizeURL(targetURL)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result, err
	}
	result.URL = normalizedURL

	// Fetch the page
	resp, body, err := a.httpClient.Fetch(normalizedURL)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to fetch URL: %v", err)
		return result, err
	}

	if resp != nil {
		result.StatusCode = resp.StatusCode
	}

	if body == nil {
		if resp != nil && resp.StatusCode != 200 {
			result.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return result, nil
	}
	defer body.Close()

	// Parse HTML
	doc, err := html.Parse(body)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to parse HTML: %v", err)
		return result, err
	}

	// Perform analysis
	result.HTMLVersion = a.htmlVersionDetector.Detect(doc, resp.Header.Get("Content-Type"))
	result.PageTitle = a.titleExtractor.Extract(doc)
	result.Headings = a.headingExtractor.Extract(doc)
	result.HasLoginForm = a.loginFormDetector.Detect(doc)

	// Analyze links
	baseURL := a.getBaseURL(normalizedURL)
	internal, external, links := a.linkAnalyzer.Analyze(doc, baseURL)
	result.InternalLinks = internal
	result.ExternalLinks = external
	result.Links = links
	result.InaccessibleLinks = a.linkAnalyzer.CheckAccessibility(links)

	return result, nil
}

func (a *Analyzer) normalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %v", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	return parsedURL.String(), nil
}

func (a *Analyzer) getBaseURL(rawURL string) string {
	parsedURL, _ := url.Parse(rawURL)
	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}
