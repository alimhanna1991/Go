package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ElasticsearchLogger struct {
	client  *http.Client
	baseURL string
	index   string
}

func NewElasticsearchLogger(client *http.Client, baseURL, index string) *ElasticsearchLogger {
	return &ElasticsearchLogger{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		index:   index,
	}
}

func (l *ElasticsearchLogger) LogError(ctx context.Context, entry Entry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%s/_doc", l.baseURL, l.index),
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("elasticsearch returned status %d", resp.StatusCode)
	}

	return nil
}
