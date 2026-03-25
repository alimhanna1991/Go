package logging

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteLogger_LogError(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "errors.db")
	logger, err := NewSQLiteLogger(dsn)
	if err != nil {
		t.Fatalf("NewSQLiteLogger() returned error: %v", err)
	}

	entry := Entry{
		Timestamp: time.Now().UTC(),
		Operation: "cache.set",
		URL:       "https://example.com",
		Message:   "write failed",
		Source:    "redis",
	}

	if err := logger.LogError(context.Background(), entry); err != nil {
		t.Fatalf("LogError() returned error: %v", err)
	}

	var count int
	if err := logger.db.QueryRow(`SELECT COUNT(*) FROM error_logs`).Scan(&count); err != nil {
		t.Fatalf("failed to count logged rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 logged row, got %d", count)
	}
}
