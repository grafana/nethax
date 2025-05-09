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
	client *http.Client
}

func NewHTTPProbe(url string, status int) *HTTPProbe {
	return NewHTTPProbeWithClient(url, status, http.DefaultClient)
}

func NewHTTPProbeWithClient(url string, status int, client *http.Client) *HTTPProbe {
	return &HTTPProbe{
		url:    url,
		status: status,
		client: client,
	}
}

func (p *HTTPProbe) Run(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return err
	}

	res, err := p.client.Do(req)
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
