package logging

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestElasticsearchLogger_LogError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/errors/_doc" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	logger := NewElasticsearchLogger(client, "http://elastic.local", "errors")
	err := logger.LogError(context.Background(), Entry{
		Timestamp: time.Now().UTC(),
		Operation: "analyze",
		Message:   "failed",
	})
	if err != nil {
		t.Fatalf("LogError() returned error: %v", err)
	}
}

func TestElasticsearchLogger_LogError_StatusFailure(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("bad request")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	logger := NewElasticsearchLogger(client, "http://elastic.local", "errors")
	err := logger.LogError(context.Background(), Entry{
		Timestamp: time.Now().UTC(),
		Operation: "analyze",
		Message:   "failed",
	})
	if err == nil {
		t.Fatal("expected error for failing Elasticsearch status")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
