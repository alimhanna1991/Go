package analyzer

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"webpage-analyzer/internal/models"
)

// LinkAnalyzer handles link analysis
type LinkAnalyzer struct {
	httpClient HTTPClient
}

// NewLinkAnalyzer creates a new link analyzer
func NewLinkAnalyzer(client HTTPClient) *LinkAnalyzer {
	return &LinkAnalyzer{
		httpClient: client,
	}
}

// Analyze analyzes links in the document
func (a *LinkAnalyzer) Analyze(doc *html.Node, baseURL string) (internal, external int, links []models.LinkInfo) {
	rawLinks := a.extractLinks(doc, baseURL)
	return a.categorizeAndCheckLinks(rawLinks, baseURL)
}

func (a *LinkAnalyzer) extractLinks(doc *html.Node, baseURL string) []string {
	var links []string
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && attr.Val != "" && !a.shouldSkipLink(attr.Val) {
					resolved := a.resolveURL(attr.Val, baseURL)
					if resolved != "" {
						links = append(links, resolved)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return a.uniqueLinks(links)
}

func (a *LinkAnalyzer) shouldSkipLink(link string) bool {
	skipPrefixes := []string{"#", "javascript:", "mailto:", "tel:"}
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(link, prefix) {
			return true
		}
	}
	return false
}

func (a *LinkAnalyzer) resolveURL(link, baseURL string) string {
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	parsedLink, err := url.Parse(link)
	if err != nil {
		return ""
	}

	resolved := parsedBase.ResolveReference(parsedLink)
	return resolved.String()
}

func (a *LinkAnalyzer) uniqueLinks(links []string) []string {
	seen := make(map[string]bool)
	unique := make([]string, 0, len(links))

	for _, link := range links {
		if !seen[link] {
			seen[link] = true
			unique = append(unique, link)
		}
	}

	return unique
}

func (a *LinkAnalyzer) categorizeAndCheckLinks(links []string, baseURL string) (internal, external int, linkInfos []models.LinkInfo) {
	parsedBase, _ := url.Parse(baseURL)

	for _, link := range links {
		parsedLink, err := url.Parse(link)
		if err != nil {
			continue
		}

		info := models.LinkInfo{URL: link}

		if parsedLink.Host == parsedBase.Host {
			internal++
		} else {
			external++
		}

		linkInfos = append(linkInfos, info)
	}

	return internal, external, linkInfos
}

// CheckAccessibility checks which links are accessible
func (a *LinkAnalyzer) CheckAccessibility(links []models.LinkInfo) int {
	inaccessible := 0

	for i := range links {
		if a.isSkippableLink(links[i].URL) {
			links[i].Accessible = true
			continue
		}

		accessible, statusCode, err := a.checkLinkAccessibility(links[i].URL)
		links[i].Accessible = accessible
		links[i].StatusCode = statusCode

		if err != nil {
			links[i].Error = err.Error()
		}

		if !accessible {
			inaccessible++
		}
	}

	return inaccessible
}

func (a *LinkAnalyzer) isSkippableLink(link string) bool {
	skippable := []string{"mailto:", "tel:", "javascript:"}
	for _, prefix := range skippable {
		if strings.HasPrefix(link, prefix) {
			return true
		}
	}
	return false
}

func (a *LinkAnalyzer) checkLinkAccessibility(link string) (bool, int, error) {
	resp, err := a.httpClient.Check(link)
	if err != nil {
		return false, 0, err
	}

	if resp != nil {
		accessible := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest
		return accessible, resp.StatusCode, nil
	}

	return false, 0, nil
}
