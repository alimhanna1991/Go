package logging

import (
	"context"
	"time"
)

// Entry represents a single error log entry.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation"`
	URL       string    `json:"url,omitempty"`
	Message   string    `json:"message"`
	Source    string    `json:"source,omitempty"`
}

// Logger abstracts error logging sinks.
type Logger interface {
	LogError(ctx context.Context, entry Entry) error
}

type MultiLogger struct {
	loggers []Logger
}

func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{loggers: loggers}
}

func (l *MultiLogger) LogError(ctx context.Context, entry Entry) error {
	var firstErr error
	for _, logger := range l.loggers {
		if logger == nil {
			continue
		}
		if err := logger.LogError(ctx, entry); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type NoopLogger struct{}

func (l *NoopLogger) LogError(context.Context, Entry) error {
	return nil
}
