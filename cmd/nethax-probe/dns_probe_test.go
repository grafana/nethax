package main

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestDNSProbe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Run("host resolves", func(t *testing.T) {
			p := NewDNSProbe("grafana.com", false)

			if err := p.Run(t.Context()); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("host not found", func(t *testing.T) {
			p := NewDNSProbe("example.com", true)

			if err := p.Run(t.Context()); err != nil {
				t.Fatal(err)
			}
		})
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
		p := NewDNSProbe("example.com", true)

		if err := p.Run(t.Context()); err != nil {
			t.Fatalf("unexoected error %v", err)
		}
	})

	t.Run("context aware", func(t *testing.T) {
		p := NewDNSProbe("grafana.com", false)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		if err := p.Run(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("expecting error %v, got %v", context.Canceled, err)
		}
	})
}
