package httpclient

import (
	"testing"
)

func TestCheckIP(t *testing.T) {
	blocked := []string{
		"10.0.0.1",
		"10.255.255.255",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.1.1",
		"127.0.0.1",
		"127.0.0.100",
		"169.254.1.1",
		"100.64.0.1",
		"0.0.0.0",
		"::1",
		"fc00::1",
		"fe80::1",
	}
	for _, ip := range blocked {
		if err := CheckIP(ip); err == nil {
			t.Errorf("expected block for %q, got nil", ip)
		}
	}

	allowed := []string{
		"1.1.1.1",
		"8.8.8.8",
		"203.0.113.5",
		"2001:db8::1",
	}
	for _, ip := range allowed {
		if err := CheckIP(ip); err != nil {
			t.Errorf("expected allow for %q, got %v", ip, err)
		}
	}
}

func TestCheckIP_NonIP(t *testing.T) {
	// Non-IP strings should return nil (DNS validation is handled elsewhere)
	if err := CheckIP("not-an-ip"); err != nil {
		t.Errorf("expected nil for non-IP, got %v", err)
	}
	if err := CheckIP(""); err != nil {
		t.Errorf("expected nil for empty string, got %v", err)
	}
}
