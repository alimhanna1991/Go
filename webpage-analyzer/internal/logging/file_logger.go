package logging

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type FileLogger struct {
	path string
	mu   sync.Mutex
}

func NewFileLogger(path string) (*FileLogger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	return &FileLogger{path: path}, nil
}

func (l *FileLogger) LogError(_ context.Context, entry Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if _, err := file.Write(append(payload, '\n')); err != nil {
		return err
	}

	return nil
}
