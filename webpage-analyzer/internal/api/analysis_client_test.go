package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestAnalysisClient_AnalyzeURL(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/api/v1/analyze" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"url":"https://example.com","page_title":"Remote Result"}`)),
			}, nil
		}),
	}

	client := NewAnalysisClient("http://analysis-service", httpClient)
	result, err := client.AnalyzeURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("AnalyzeURL() returned error: %v", err)
	}
	if result.PageTitle != "Remote Result" {
		t.Fatalf("expected remote result, got %+v", result)
	}
}

func TestAnalysisClient_AnalyzeURL_PropagatesRemoteErrors(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"url":"https://example.com","error_message":"upstream failure"}`)),
			}, nil
		}),
	}

	client := NewAnalysisClient("http://analysis-service", httpClient)
	result, err := client.AnalyzeURL(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result == nil || result.ErrorMessage != "upstream failure" {
		t.Fatalf("expected remote error payload, got %+v", result)
	}
}
