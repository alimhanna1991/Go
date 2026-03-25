package analyzer

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

// MockHTTPClient for testing
type MockHTTPClient struct {
	ResponseBody string
	StatusCode   int
	Error        error
}

type MockPageRenderer struct {
	HTML  string
	Error error
}

func (m *MockHTTPClient) Fetch(url string) (*http.Response, io.ReadCloser, error) {
	if m.Error != nil {
		return nil, nil, m.Error
	}

	resp := &http.Response{
		StatusCode: m.StatusCode,
		Header:     make(http.Header),
	}

	if m.StatusCode == 0 {
		resp.StatusCode = http.StatusOK
	}

	body := io.NopCloser(bytes.NewReader([]byte(m.ResponseBody)))
	return resp, body, nil
}

func (m *MockHTTPClient) Check(url string) (*http.Response, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	statusCode := m.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
	}, nil
}

func (m *MockPageRenderer) RenderHTML(url string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.HTML, nil
}

func TestAnalyzer_Analyze(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		mockClient *MockHTTPClient
		wantErr    bool
		validate   func(*testing.T, *AnalysisResult)
	}{
		{
			name: "Successful analysis",
			url:  "https://example.com",
			mockClient: &MockHTTPClient{
				StatusCode: http.StatusOK,
				ResponseBody: `
                    <!DOCTYPE html>
                    <html>
                    <head><title>Test Page</title></head>
                    <body>
                        <h1>Main Heading</h1>
                        <h2>Sub Heading</h2>
                        <a href="https://example.com/internal">Internal</a>
                        <a href="https://external.com">External</a>
                        <form><input type="password"></form>
                    </body>
                    </html>
                `,
			},
			wantErr: false,
			validate: func(t *testing.T, result *AnalysisResult) {
				if result.PageTitle != "Test Page" {
					t.Errorf("Expected title 'Test Page', got '%s'", result.PageTitle)
				}
				if result.Headings["h1"] != 1 {
					t.Errorf("Expected 1 h1, got %d", result.Headings["h1"])
				}
				if result.Headings["h2"] != 1 {
					t.Errorf("Expected 1 h2, got %d", result.Headings["h2"])
				}
				if !result.HasLoginForm {
					t.Error("Expected login form to be detected")
				}
			},
		},
		{
			name: "404 error handling",
			url:  "https://example.com/404",
			mockClient: &MockHTTPClient{
				StatusCode:   http.StatusNotFound,
				ResponseBody: "",
			},
			wantErr: false,
			validate: func(t *testing.T, result *AnalysisResult) {
				if result.StatusCode != http.StatusNotFound {
					t.Errorf("Expected status code %d, got %d", http.StatusNotFound, result.StatusCode)
				}
				if result.ErrorMessage == "" {
					t.Error("Expected error message for 404")
				}
			},
		},
		{
			name:       "Invalid URL",
			url:        "not-a-valid-url",
			mockClient: &MockHTTPClient{},
			wantErr:    true,
			validate: func(t *testing.T, result *AnalysisResult) {
				if result.ErrorMessage == "" {
					t.Error("Expected error message for invalid URL")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(tt.mockClient, nil)
			result, err := analyzer.Analyze(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("Analyze() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestAnalyzer_Analyze_UsesRenderedDOMForLoginDetection(t *testing.T) {
	rawHTML := `<!DOCTYPE html><html><head><title>Spring Academy</title></head><body><div id="app"></div></body></html>`
	renderedHTML := `<!DOCTYPE html><html><body><form action="/auth"><input type="password"><button>Sign in</button></form></body></html>`

	analyzer := NewAnalyzer(&MockHTTPClient{
		StatusCode:   http.StatusOK,
		ResponseBody: rawHTML,
	}, &MockPageRenderer{
		HTML: renderedHTML,
	})

	result, err := analyzer.Analyze("https://spring.academy/auth")
	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}
	if !result.HasLoginForm {
		t.Fatal("expected login form to be detected from rendered DOM")
	}
}
