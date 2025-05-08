package main

import (
	"context"
	"fmt"
	"net/http"
)

var _ Probe = &HTTPProbe{}

type HTTPProbe struct {
	url    string
	status int
}

func NewHTTPProbe(url string, status int) *HTTPProbe {
	return &HTTPProbe{
		url:    url,
		status: status,
	}
}

func (p *HTTPProbe) Run(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if p.status == 0 { // expecting failure
			return nil
		}

		return fmt.Errorf("%w: %w", errConnectionFailed, err)
	}

	defer res.Body.Close() //nolint:errcheck

	if p.status == 0 {
		return errConnectionSucceeded
	} else if p.status != res.StatusCode {
		return fmt.Errorf("%w: expecting response code %d, got %d", errAssertionFailed, p.status, res.StatusCode)
	}

	return nil
}
