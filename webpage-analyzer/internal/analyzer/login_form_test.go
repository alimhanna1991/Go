package analyzer

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestLoginFormDetector_Detect(t *testing.T) {
	detector := NewLoginFormDetector()

	tests := []struct {
		name     string
		document string
		expected bool
	}{
		{
			name: "password and username fields",
			document: `
				<html><body>
					<form action="/login">
						<input type="text" name="username">
						<input type="password" name="password">
						<button type="submit">Log in</button>
					</form>
				</body></html>
			`,
			expected: true,
		},
		{
			name: "password and submit hint only",
			document: `
				<html><body>
					<form action="/signin">
						<input type="password" name="password">
						<button type="submit">Sign in</button>
					</form>
				</body></html>
			`,
			expected: true,
		},
		{
			name: "password reset form is not login",
			document: `
				<html><body>
					<form action="/reset-password">
						<input type="email" name="email">
						<button type="submit">Send reset link</button>
					</form>
				</body></html>
			`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.document))
			if err != nil {
				t.Fatalf("failed to parse html: %v", err)
			}

			got := detector.Detect(doc, "", "")
			if got != tt.expected {
				t.Fatalf("expected %t, got %t", tt.expected, got)
			}
		})
	}
}

func TestLoginFormDetector_DetectAuthPageHeuristics(t *testing.T) {
	detector := NewLoginFormDetector()
	doc, err := html.Parse(strings.NewReader(`<html><body><div id="app"><button>Sign in</button></div></body></html>`))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	if !detector.Detect(doc, "https://spring.academy/auth", "Spring Academy Sign In") {
		t.Fatal("expected auth page heuristics to detect login")
	}
}
