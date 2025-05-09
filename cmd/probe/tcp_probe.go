package main

import (
	"context"
	"fmt"
	"net"
)

var _ Probe = TCPProbe{}

type TCPProbe struct {
	addr string
	fail bool
}

func NewTCPProbe(addr string, fail bool) TCPProbe {
	return TCPProbe{
		addr: addr,
		fail: fail,
	}
}

func (p TCPProbe) Run(ctx context.Context) error {
	var d net.Dialer

	cn, err := d.DialContext(ctx, "tcp", p.addr)
	if err != nil {
		if p.fail {
			return nil
		}

		return fmt.Errorf("%w: %w", errConnectionFailed, err)
	}
	defer cn.Close() //nolint:errcheck

	if p.fail {
		return fmt.Errorf("%w: %w", errAssertionFailed, errConnectionSucceeded)
	}

	return nil
}
