package logging

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubLogger struct {
	err   error
	calls int
}

func (l *stubLogger) LogError(ctx context.Context, entry Entry) error {
	l.calls++
	return l.err
}

func TestMultiLogger_LogError(t *testing.T) {
	first := &stubLogger{err: errors.New("first")}
	second := &stubLogger{}

	logger := NewMultiLogger(first, second)
	err := logger.LogError(context.Background(), Entry{
		Timestamp: time.Now().UTC(),
		Operation: "test",
		Message:   "failure",
	})

	if err == nil || err.Error() != "first" {
		t.Fatalf("expected first error to be returned, got %v", err)
	}
	if first.calls != 1 || second.calls != 1 {
		t.Fatalf("expected both loggers to be called once, got %d and %d", first.calls, second.calls)
	}
}

func TestNoopLogger_LogError(t *testing.T) {
	logger := &NoopLogger{}
	if err := logger.LogError(context.Background(), Entry{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
