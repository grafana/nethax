package main

import (
	"fmt"
	"net/http"
	"time"
)

var _ Probe = &HTTPProbe{}

type HTTPProbe struct {
	url       string
	timeout   time.Duration
	resStatus int
	expStatus int
	err       error
}

func NewHTTPProbe(url string, timeout time.Duration, expectedStatus int) *HTTPProbe {
	return &HTTPProbe{
		url:       url,
		timeout:   timeout,
		expStatus: expectedStatus,
	}
}

func (p *HTTPProbe) Run() error {
	c := &http.Client{
		Timeout: timeout,
	}

	res, err := c.Get(p.url)
	if err != nil {
		p.err = err
	} else {
		defer res.Body.Close() //nolint:errcheck
		p.resStatus = res.StatusCode
	}

	if p.expStatus == 0 && p.err == nil {
		return errConnectionSucceeded
	} else if p.expStatus != 0 && p.err != nil {
		return fmt.Errorf("%w: %w", errConnectionFailed, p.err)
	} else if p.expStatus != p.resStatus {
		return fmt.Errorf("%w: expecting response code %d, got %d", errAssertionFailed, p.expStatus, p.resStatus)
	}

	return nil
}
