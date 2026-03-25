package http

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDefaultHTTPClient_Fetch(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		statusCode int
		wantErr    bool
		client     *DefaultHTTPClient
	}{
		{
			name:       "Successful fetch",
			url:        "http://example.com",
			statusCode: http.StatusOK,
			wantErr:    false,
			client: newTestClient(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("test response")),
					Header:     make(http.Header),
				}, nil
			}),
		},
		{
			name:    "Invalid URL",
			url:     "not-a-valid-url",
			wantErr: true,
			client:  NewDefaultHTTPClient(),
		},
		{
			name:    "Timeout test",
			url:     "http://localhost:9999", // Non-existent
			wantErr: true,
			client: newTestClient(func(req *http.Request) (*http.Response, error) {
				return nil, &net.OpError{Err: errors.New("connection refused")}
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := tt.client.Fetch(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && body != nil {
				defer body.Close()
				if resp.StatusCode != tt.statusCode {
					t.Errorf("Expected status code %d, got %d", tt.statusCode, resp.StatusCode)
				}
			}
		})
	}
}

func TestDefaultHTTPClient_Timeout(t *testing.T) {
	config := &ClientConfig{
		Timeout:      1 * time.Second,
		MaxRedirects: 10,
	}

	client := NewDefaultHTTPClientWithConfig(config)
	client.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, &timeoutError{}
	})
	_, _, err := client.Fetch("http://example.com")

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestDefaultHTTPClient_Check(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodHead {
			t.Fatalf("expected HEAD request, got %s", req.Method)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})

	resp, err := client.Check("http://example.com")
	if err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDefaultHTTPClient_CheckFallsBackToGet(t *testing.T) {
	callCount := 0
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 && req.Method == http.MethodHead {
			return &http.Response{
				StatusCode: http.StatusMethodNotAllowed,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}

		if callCount == 2 && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}, nil
		}

		t.Fatalf("unexpected request #%d with method %s", callCount, req.Method)
		return nil, nil
	})

	resp, err := client.Check("http://example.com")
	if err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func newTestClient(fn roundTripFunc) *DefaultHTTPClient {
	client := NewDefaultHTTPClient()
	client.client.Transport = fn
	return client
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type timeoutError struct{}

func (e *timeoutError) Error() string {
	return "request timeout"
}

func (e *timeoutError) Timeout() bool {
	return true
}

func (e *timeoutError) Temporary() bool {
	return true
}
