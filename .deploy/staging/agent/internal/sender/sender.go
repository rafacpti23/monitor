package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/p-mon/agent/internal/collector"
	"github.com/p-mon/agent/internal/config"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// Send marshals the payload to gzip-compressed JSON and POSTs it to the backend.
func Send(cfg *config.Config, payload collector.Payload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("gzip close: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/agent/%s", cfg.BackendURL, cfg.ServerKey)

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server responded %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendWithRetry sends the payload with exponential backoff retries.
// It tries up to maxRetries additional attempts after the first failure.
func SendWithRetry(cfg *config.Config, payload collector.Payload, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = Send(cfg, payload)
		if lastErr == nil {
			return nil
		}
		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second // 1s, 2s, 4s
			time.Sleep(backoff)
		}
	}
	return lastErr
}
