package analyzer

import (
	"golang.org/x/net/html"
)

// HeadingExtractor handles heading extraction
type HeadingExtractor struct {
	headingLevels []string
}

// NewHeadingExtractor creates a new heading extractor
func NewHeadingExtractor() *HeadingExtractor {
	return &HeadingExtractor{
		headingLevels: []string{"h1", "h2", "h3", "h4", "h5", "h6"},
	}
}

// Extract extracts all headings from the document
func (e *HeadingExtractor) Extract(doc *html.Node) map[string]int {
	headings := make(map[string]int)

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, level := range e.headingLevels {
				if n.Data == level {
					headings[level]++
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return headings
}
