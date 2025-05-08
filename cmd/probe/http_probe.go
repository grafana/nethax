package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

var _ Probe = &HTTPProbe{}

type HTTPProbe struct {
	url     string
	timeout time.Duration
	status  int
}

func NewHTTPProbe(url string, timeout time.Duration, status int) *HTTPProbe {
	return &HTTPProbe{
		url:     url,
		timeout: timeout,
		status:  status,
	}
}

func (p *HTTPProbe) Run(_ context.Context) error {
	c := &http.Client{
		Timeout: timeout,
	}

	res, err := c.Get(p.url)
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
