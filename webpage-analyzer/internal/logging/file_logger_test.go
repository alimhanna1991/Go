package logging

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileLogger_LogError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "errors.jsonl")
	logger, err := NewFileLogger(path)
	if err != nil {
		t.Fatalf("NewFileLogger() returned error: %v", err)
	}

	entry := Entry{
		Timestamp: time.Now().UTC(),
		Operation: "test.operation",
		URL:       "https://example.com",
		Message:   "something failed",
		Source:    "unit-test",
	}

	if err := logger.LogError(context.Background(), entry); err != nil {
		t.Fatalf("LogError() returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logLine := string(content)
	if !strings.Contains(logLine, `"operation":"test.operation"`) {
		t.Fatalf("expected operation field in log line, got %s", logLine)
	}
	if !strings.HasSuffix(logLine, "\n") {
		t.Fatalf("expected newline-terminated log entry, got %q", logLine)
	}
}
