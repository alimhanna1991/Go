package analyzer

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestHTMLVersionDetector_Detect(t *testing.T) {
	detector := NewHTMLVersionDetector()

	tests := []struct {
		name        string
		html        string
		contentType string
		expected    string
	}{
		{
			name:     "HTML5 doctype",
			html:     `<!DOCTYPE html><html></html>`,
			expected: "HTML5",
		},
		{
			name:     "HTML 4.01",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN">`,
			expected: "HTML 4.01",
		},
		{
			name:     "XHTML 1.0",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN">`,
			expected: "XHTML 1.0",
		},
		{
			name:     "No doctype - assume HTML5",
			html:     `<html></html>`,
			expected: "HTML5 (assumed - no doctype found)",
		},
		{
			name:        "XHTML content type",
			html:        `<html></html>`,
			contentType: "application/xhtml+xml",
			expected:    "XHTML (assumed from content type)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := detector.Detect(doc, tt.contentType)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
