package http

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"time"
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Fetch(url string) (*http.Response, io.ReadCloser, error)
	Check(url string) (*http.Response, error)
}

// DefaultHTTPClient implements HTTPClient
type DefaultHTTPClient struct {
	client *http.Client
	config *ClientConfig
}

type ClientConfig struct {
	Timeout            time.Duration
	MaxRedirects       int
	InsecureSkipVerify bool
}

// NewDefaultHTTPClient creates a new HTTP client with default config
func NewDefaultHTTPClient() *DefaultHTTPClient {
	config := &ClientConfig{
		Timeout:            10 * time.Second,
		MaxRedirects:       10,
		InsecureSkipVerify: false,
	}

	return NewDefaultHTTPClientWithConfig(config)
}

func NewClientConfig(timeout time.Duration, maxRedirects int, insecureSkipVerify bool) *ClientConfig {
	return &ClientConfig{
		Timeout:            timeout,
		MaxRedirects:       maxRedirects,
		InsecureSkipVerify: insecureSkipVerify,
	}
}

// NewDefaultHTTPClientWithConfig creates a new HTTP client with custom config
func NewDefaultHTTPClientWithConfig(config *ClientConfig) *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: config.InsecureSkipVerify,
				},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= config.MaxRedirects {
					return errors.New("too many redirects")
				}
				return nil
			},
		},
		config: config,
	}
}

// Fetch implements the HTTPClient interface
func (c *DefaultHTTPClient) Fetch(url string) (*http.Response, io.ReadCloser, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return resp, nil, nil
	}

	return resp, resp.Body, nil
}

// Check performs a lightweight availability check for a URL.
func (c *DefaultHTTPClient) Check(targetURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodHead, targetURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
		resp.Body.Close()
		req, err = http.NewRequest(http.MethodGet, targetURL, nil)
		if err != nil {
			return nil, err
		}
		resp, err = c.client.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	return resp, nil
}
