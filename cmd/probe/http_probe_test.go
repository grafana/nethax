package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPProbe(t *testing.T) {
	t.Run("expecting status code", func(t *testing.T) {
		status := http.StatusOK

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))

		p := NewHTTPProbe(ts.URL, time.Second, status)

		if err := p.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("expecting connection failure", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}))

		p := NewHTTPProbe(ts.URL, time.Second, 0)

		if err := p.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("connection succeeded", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		p := NewHTTPProbe(ts.URL, time.Second, 0)

		if err := p.Run(); !errors.Is(err, errConnectionSucceeded) {
			t.Fatalf("expecting error %v, got %v", errConnectionSucceeded, err)
		}
	})

	t.Run("unexpected response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		p := NewHTTPProbe(ts.URL, time.Second, http.StatusOK)

		if err := p.Run(); !errors.Is(err, errAssertionFailed) {
			t.Fatalf("expecting error %v, got %v", errAssertionFailed, err)
		}
	})

	t.Run("connection failure", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		p := NewHTTPProbe(ts.URL, time.Second, http.StatusOK)

		if err := p.Run(); !errors.Is(err, errConnectionFailed) {
			t.Fatalf("expecting error %v, got %v", errConnectionFailed, err)
		}
	})
}
