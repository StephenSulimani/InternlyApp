package ats

import (
	"context"
	"net/http"
	"time"
)

const defaultUserAgent = "InternlyApp/1.0 (job-aggregator)"

func NewClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}

func newJSONRequest(ctx context.Context, method, rawURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/json")
	return req, nil
}
