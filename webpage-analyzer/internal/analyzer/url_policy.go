package analyzer

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func normalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %v", err)
	}

	if parsedURL.Host == "" && parsedURL.Scheme == "" && parsedURL.Path != "" {
		if !looksLikeHost(parsedURL.Path) {
			return "", fmt.Errorf("invalid URL format: missing host")
		}
		parsedURL, err = url.Parse("http://" + rawURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %v", err)
		}
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("invalid URL scheme: only http and https are supported")
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("invalid URL format: missing host")
	}

	if err := validateTargetHost(parsedURL.Hostname()); err != nil {
		return "", err
	}

	return parsedURL.String(), nil
}

func baseURL(rawURL string) string {
	parsedURL, _ := url.Parse(rawURL)
	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}

func looksLikeHost(value string) bool {
	if value == "localhost" {
		return true
	}

	return strings.Contains(value, ".")
}

func validateTargetHost(host string) error {
	normalizedHost := strings.TrimSpace(strings.ToLower(host))
	if normalizedHost == "" {
		return fmt.Errorf("invalid URL format: missing host")
	}

	if normalizedHost == "localhost" {
		return fmt.Errorf("target host is not allowed")
	}

	ip := net.ParseIP(normalizedHost)
	if ip == nil {
		return nil
	}

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return fmt.Errorf("target host is not allowed")
	}

	return nil
}
