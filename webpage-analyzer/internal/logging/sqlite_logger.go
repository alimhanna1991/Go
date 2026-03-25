package logging

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteLogger struct {
	db *sql.DB
}

func NewSQLiteLogger(dsn string) (*SQLiteLogger, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	logger := &SQLiteLogger{db: db}
	if err := logger.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return logger, nil
}

func (l *SQLiteLogger) init() error {
	_, err := l.db.Exec(`
		CREATE TABLE IF NOT EXISTS error_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			operation TEXT NOT NULL,
			url TEXT,
			message TEXT NOT NULL,
			source TEXT
		)
	`)
	return err
}

func (l *SQLiteLogger) LogError(ctx context.Context, entry Entry) error {
	_, err := l.db.ExecContext(
		ctx,
		`INSERT INTO error_logs(timestamp, operation, url, message, source) VALUES(?, ?, ?, ?, ?)`,
		entry.Timestamp.Format("2006-01-02T15:04:05.000000000Z07:00"),
		entry.Operation,
		entry.URL,
		entry.Message,
		entry.Source,
	)
	return err
}
