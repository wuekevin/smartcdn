package httpclient

import (
	"errors"
	"net"
)

// ErrSSRFBlocked is returned when a value would route requests to internal infrastructure.
var ErrSSRFBlocked = errors.New("SSRF_BLOCKED: address resolves to a private or reserved IP range")

var privateRanges = []net.IPNet{
	parseCIDR("10.0.0.0/8"),
	parseCIDR("172.16.0.0/12"),
	parseCIDR("192.168.0.0/16"),
	parseCIDR("100.64.0.0/10"), // shared address space
	parseCIDR("127.0.0.0/8"),   // loopback
	parseCIDR("169.254.0.0/16"), // link-local
	parseCIDR("::1/128"),        // IPv6 loopback
	parseCIDR("fc00::/7"),       // IPv6 ULA
	parseCIDR("fe80::/10"),      // IPv6 link-local
}

func parseCIDR(s string) net.IPNet {
	_, n, _ := net.ParseCIDR(s)
	return *n
}

// CheckIP validates that the given IP string is not a private/reserved address.
// Returns ErrSSRFBlocked if the IP falls within any protected range.
func CheckIP(ipStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil // not a valid IP; let DNS validation handle it
	}
	// Block 0.0.0.0 and ::/0 (all-zeros)
	if ip.Equal(net.IPv4zero) || ip.Equal(net.IPv6zero) {
		return ErrSSRFBlocked
	}
	for _, r := range privateRanges {
		if r.Contains(ip) {
			return ErrSSRFBlocked
		}
	}
	return nil
}
