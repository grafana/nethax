package main

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestDNSProbe(t *testing.T) {
	t.Run("host resolves", func(t *testing.T) {
		p := NewDNSProbe("grafana.com", false)

		if err := p.Run(t.Context()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("conn should fail", func(t *testing.T) {
		p := NewDNSProbe("grafana.com", false)

		// TODO(inkel) this is a hack and should be done differently
		p.r = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return nil, errors.New("host not found")
			},
		}

		if err := p.Run(t.Context()); !errors.Is(err, errConnectionFailed) {
			t.Fatal(err)
		}
	})

	t.Run("probe should fail", func(t *testing.T) {
		p := NewDNSProbe("nethax.grafana.com", true)

		if err := p.Run(t.Context()); err != nil {
			t.Fatalf("unexpected error %v", err)
		}
	})

	t.Run("probe succeeds when it should fail", func(t *testing.T) {
		p := NewDNSProbe("grafana.com", true)

		if err := p.Run(t.Context()); err == nil {
			t.Fatal("expecting error %w", errConnectionSucceeded)
		}
	})

	t.Run("context aware", func(t *testing.T) {
		p := NewDNSProbe("grafana.com", false)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		// This should be !errors.Is(err, context.Canceled) but there
		// seems to be a bug in Go stdlib, check
		// https://github.com/golang/go/issues/71939 for additional
		// information.
		if err := p.Run(ctx); !strings.Contains(err.Error(), "operation was canceled") {
			t.Fatalf("expecting error %v, got %v", context.Canceled, err)
		}
	})
}
