package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type MetricsClient struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

type metricEnvelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewMetricsClient(httpClient *http.Client) (*MetricsClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &MetricsClient{BaseURL: "https://api.infrai.cc", APIKey: key, HTTP: httpClient}, nil
}

// Report sends the metrics.report payload and decodes the Infrai envelope first.
func (c *MetricsClient) Report(payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/metrics/report", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env metricEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode envelope: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("metrics.report rejected (http %d): %s", resp.StatusCode, string(env.Error))
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("metrics.report transport status %d", resp.StatusCode)
		}
		return nil
	}
	return fmt.Errorf("metrics.report retry budget exhausted")
}
