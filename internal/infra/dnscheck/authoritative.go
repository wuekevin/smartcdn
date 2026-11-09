package dnscheck

import (
	"context"
	"fmt"
	"net"
	"time"
)

// AuthResult holds the result of an authoritative DNS lookup.
type AuthResult struct {
	Host    string
	Type    string
	Values  []string
	Latency time.Duration
}

var resolver = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{Timeout: 5 * time.Second}
		return d.DialContext(ctx, "udp", "8.8.8.8:53")
	},
}

// LookupA resolves A records for host using an external resolver (bypasses local cache).
func LookupA(ctx context.Context, host string) (*AuthResult, error) {
	start := time.Now()
	addrs, err := resolver.LookupHost(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("LookupA %q: %w", host, err)
	}
	return &AuthResult{Host: host, Type: "A", Values: addrs, Latency: time.Since(start)}, nil
}

// LookupTXT resolves TXT records for host.
func LookupTXT(ctx context.Context, host string) (*AuthResult, error) {
	start := time.Now()
	txts, err := resolver.LookupTXT(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("LookupTXT %q: %w", host, err)
	}
	return &AuthResult{Host: host, Type: "TXT", Values: txts, Latency: time.Since(start)}, nil
}

// ContainsValue reports whether result.Values contains the expected value.
func ContainsValue(result *AuthResult, expected string) bool {
	for _, v := range result.Values {
		if v == expected {
			return true
		}
	}
	return false
}
