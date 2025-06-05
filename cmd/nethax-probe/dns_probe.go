package main

import (
	"context"
	"fmt"
	"net"
)

type DNSProbe struct {
	host string
	fail bool
	r    *net.Resolver
}

func NewDNSProbe(host string, fail bool) DNSProbe {
	return DNSProbe{host: host, fail: fail}
}

func (p DNSProbe) Run(ctx context.Context) error {
	if _, err := p.r.LookupHost(ctx, p.host); err != nil {
		if p.fail {
			return nil
		}
		return fmt.Errorf("%w: %w", errConnectionFailed, err)
	}

	if p.fail {
		return fmt.Errorf("%w: %w", errAssertionFailed, errConnectionSucceeded)
	}

	return nil
}
