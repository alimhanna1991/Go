package analyzer

import (
	"fmt"
	"net/http"
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

	normalizedURL, err := normalizeURL(targetURL)
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
	a.detectLoginForm(result, doc, normalizedURL)

	internal, external, links := a.linkAnalyzer.Analyze(doc, baseURL(normalizedURL))
	result.InternalLinks = internal
	result.ExternalLinks = external
	result.Links = links
	result.InaccessibleLinks = a.linkAnalyzer.CheckAccessibility(result.Links)

	return result, nil
}

func (a *Analyzer) detectLoginForm(result *models.AnalysisResult, doc *html.Node, targetURL string) {
	result.HasLoginForm = a.loginFormDetector.Detect(doc, targetURL, result.PageTitle)
	if result.HasLoginForm {
		return
	}

	renderedDoc, err := a.renderDocument(targetURL)
	if err != nil || renderedDoc == nil {
		return
	}

	renderedTitle := a.titleExtractor.Extract(renderedDoc)
	if renderedTitle != "" {
		result.PageTitle = renderedTitle
	}
	result.HasLoginForm = a.loginFormDetector.Detect(renderedDoc, targetURL, result.PageTitle)
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
