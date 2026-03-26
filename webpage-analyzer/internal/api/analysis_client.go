package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"webpage-analyzer/internal/models"
)

// AnalysisClient calls the analysis service over HTTP.
type AnalysisClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAnalysisClient creates a service client for the analysis API.
func NewAnalysisClient(baseURL string, httpClient *http.Client) *AnalysisClient {
	normalizedBaseURL := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &AnalysisClient{
		baseURL:    normalizedBaseURL,
		httpClient: httpClient,
	}
}

// AnalyzeURL requests webpage analysis from the remote service.
func (c *AnalysisClient) AnalyzeURL(ctx context.Context, targetURL string) (*models.AnalysisResult, error) {
	requestBody, err := json.Marshal(models.AnalysisRequest{URL: targetURL})
	if err != nil {
		return nil, fmt.Errorf("marshal analysis request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/analyze", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create analysis request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call analysis service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read analysis response: %w", err)
	}

	var result models.AnalysisResult
	if len(body) > 0 {
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("decode analysis response: %w", err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		if result.ErrorMessage == "" {
			result.ErrorMessage = fmt.Sprintf("analysis service returned status %d", resp.StatusCode)
		}
		return &result, fmt.Errorf(result.ErrorMessage)
	}

	return &result, nil
}
