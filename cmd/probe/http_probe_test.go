package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPProbe(t *testing.T) {
	t.Run("expecting status code", func(t *testing.T) {
		status := http.StatusOK

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))

		p := NewHTTPProbe(ts.URL, status)

		if err := p.Run(t.Context()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("expecting connection failure", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}))

		p := NewHTTPProbe(ts.URL, 0)

		if err := p.Run(t.Context()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("connection succeeded", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		p := NewHTTPProbe(ts.URL, 0)

		if err := p.Run(t.Context()); !errors.Is(err, errConnectionSucceeded) {
			t.Fatalf("expecting error %v, got %v", errConnectionSucceeded, err)
		}
	})

	t.Run("unexpected response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		p := NewHTTPProbe(ts.URL, http.StatusOK)

		if err := p.Run(t.Context()); !errors.Is(err, errAssertionFailed) {
			t.Fatalf("expecting error %v, got %v", errAssertionFailed, err)
		}
	})

	t.Run("connection failure", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		p := NewHTTPProbe(ts.URL, http.StatusOK)

		if err := p.Run(t.Context()); !errors.Is(err, errConnectionFailed) {
			t.Fatalf("expecting error %v, got %v", errConnectionFailed, err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		p := NewHTTPProbe(ts.URL, http.StatusOK)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		if err := p.Run(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("expecting error %v, got %v", context.Canceled, err)
		}
	})

	t.Run("custom client", func(t *testing.T) {
		status := http.StatusOK

		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))

		p := NewHTTPProbeWithClient(ts.URL, status, ts.Client())

		if err := p.Run(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
}
