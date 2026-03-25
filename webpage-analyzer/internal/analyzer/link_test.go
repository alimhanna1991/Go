package analyzer

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestLinkAnalyzer_Analyze(t *testing.T) {
	mockClient := &MockHTTPClient{}
	linkAnalyzer := NewLinkAnalyzer(mockClient)

	tests := []struct {
		name             string
		html             string
		baseURL          string
		expectedInternal int
		expectedExternal int
	}{
		{
			name: "Mixed internal and external links",
			html: `
                <a href="/internal">Internal 1</a>
                <a href="https://example.com/internal2">Internal 2</a>
                <a href="https://external.com">External 1</a>
                <a href="https://another.com">External 2</a>
            `,
			baseURL:          "https://example.com",
			expectedInternal: 2,
			expectedExternal: 2,
		},
		{
			name: "Only internal links",
			html: `
                <a href="/page1">Page 1</a>
                <a href="/page2">Page 2</a>
            `,
			baseURL:          "https://example.com",
			expectedInternal: 2,
			expectedExternal: 0,
		},
		{
			name: "Skip anchor and javascript links",
			html: `
                <a href="#section">Anchor</a>
                <a href="javascript:void(0)">JS Link</a>
                <a href="mailto:test@example.com">Email</a>
                <a href="/valid">Valid</a>
            `,
			baseURL:          "https://example.com",
			expectedInternal: 1, // Only /valid
			expectedExternal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			internal, external, _ := linkAnalyzer.Analyze(doc, tt.baseURL)

			if internal != tt.expectedInternal {
				t.Errorf("Expected %d internal links, got %d", tt.expectedInternal, internal)
			}

			if external != tt.expectedExternal {
				t.Errorf("Expected %d external links, got %d", tt.expectedExternal, external)
			}
		})
	}
}
