package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault_Validate(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected default config to be valid, got %v", err)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")

	content := []byte(`
server:
  port: "9090"
http_client:
  timeout_seconds: 12
  max_redirects: 5
  insecure_skip_verify: true
browser:
  enabled: false
  command: "custom-chrome"
  timeout_seconds: 20
logging:
  enabled: true
  backends: ["file"]
  file:
    path: "logs/test.jsonl"
cache:
  enabled: true
  ttl_seconds: 600
  redis:
    addr: "redis:6379"
template_paths:
  index: "web/templates/index.html"
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Port != "9090" {
		t.Fatalf("expected port 9090, got %q", cfg.Server.Port)
	}
	if !cfg.Cache.Enabled {
		t.Fatal("expected cache to be enabled")
	}
	if cfg.Browser.Command != "custom-chrome" {
		t.Fatalf("expected browser command override, got %q", cfg.Browser.Command)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := Default()
	cfg.Server.Port = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty server port")
	}
}
