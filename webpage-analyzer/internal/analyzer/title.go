package analyzer

import (
	"golang.org/x/net/html"
)

// TitleExtractor handles page title extraction
type TitleExtractor struct{}

// NewTitleExtractor creates a new title extractor
func NewTitleExtractor() *TitleExtractor {
	return &TitleExtractor{}
}

// Extract extracts the page title from the document
func (e *TitleExtractor) Extract(doc *html.Node) string {
	var title string
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return title
}
