package analyzer

import (
	"strings"

	"golang.org/x/net/html"
)

// HTMLVersionDetector handles HTML version detection
type HTMLVersionDetector struct{}

// NewHTMLVersionDetector creates a new HTML version detector
func NewHTMLVersionDetector() *HTMLVersionDetector {
	return &HTMLVersionDetector{}
}

// Detect detects the HTML version from the document
func (d *HTMLVersionDetector) Detect(doc *html.Node, contentType string) string {
	doctype := d.extractDoctype(doc)

	if doctype == "" {
		return d.guessVersion(contentType)
	}

	return d.parseDoctype(doctype)
}

func (d *HTMLVersionDetector) extractDoctype(doc *html.Node) string {
	var doctype string
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if n.Type == html.DoctypeNode {
			parts := []string{n.Data}
			for _, attr := range n.Attr {
				parts = append(parts, attr.Key, attr.Val)
			}
			doctype = strings.Join(parts, " ")
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return doctype
}

func (d *HTMLVersionDetector) parseDoctype(doctype string) string {
	doctypeLower := strings.ToLower(doctype)

	switch {
	case strings.Contains(doctypeLower, "html") && (strings.Contains(doctypeLower, "5") || doctypeLower == "html"):
		return "HTML5"
	case strings.Contains(doctypeLower, "4.01"):
		return "HTML 4.01"
	case strings.Contains(doctypeLower, "1.0"):
		return "XHTML 1.0"
	default:
		return "Unknown HTML version"
	}
}

func (d *HTMLVersionDetector) guessVersion(contentType string) string {
	if strings.Contains(contentType, "xhtml") {
		return "XHTML (assumed from content type)"
	}
	return "HTML5 (assumed - no doctype found)"
}
