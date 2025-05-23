package main

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestTCPProbe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("unexpected error creating listener: %v", err)
		}

		t.Run("should connect", func(t *testing.T) {
			p := NewTCPProbe(l.Addr().String(), false)

			if err := p.Run(t.Context()); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("should not connect", func(t *testing.T) {
			if err := l.Close(); err != nil {
				t.Fatalf("unexpected error closing listener: %v", err)
			}
			p := NewTCPProbe(l.Addr().String(), true)

			if err := p.Run(t.Context()); err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("conn should fail", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("unexpected error creating listener: %v", err)
		}

		p := NewTCPProbe(l.Addr().String(), true)

		if err := p.Run(t.Context()); !errors.Is(err, errConnectionSucceeded) {
			t.Fatalf("expecting error %v, got %v", errConnectionSucceeded, err)
		}
	})

	t.Run("conn should not fail", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("unexpected error creating listener: %v", err)
		}

		p := NewTCPProbe(l.Addr().String(), false)

		l.Close()

		if err := p.Run(t.Context()); !errors.Is(err, errConnectionFailed) {
			t.Fatalf("expecting error %v, got %v", errConnectionFailed, err)
		}
	})

	t.Run("context aware", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("unexpected error creating listener: %v", err)
		}

		p := NewTCPProbe(l.Addr().String(), false)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		if err := p.Run(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("expecting error %v, got %v", context.Canceled, err)
		}
	})
}
