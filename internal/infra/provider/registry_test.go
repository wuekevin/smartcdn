package provider_test

import (
	"context"
	"errors"
	"testing"

	domainport "github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/provider"
)

// stubProvider implements port.DNSProvider for testing.
type stubProvider struct{ pt domainport.ProviderType }

func (s *stubProvider) ProviderType() domainport.ProviderType { return s.pt }
func (s *stubProvider) ValidateCredentials(_ context.Context) error { return nil }
func (s *stubProvider) DetectCapabilities(_ context.Context) (*domainport.Capabilities, error) {
	return &domainport.Capabilities{Lines: []string{"default"}}, nil
}
func (s *stubProvider) ListZones(_ context.Context) ([]domainport.DNSZone, error) { return nil, nil }
func (s *stubProvider) ListRecords(_ context.Context, _ string) ([]domainport.DNSRecord, error) {
	return nil, nil
}
func (s *stubProvider) CreateRecord(_ context.Context, _ string, _ domainport.DNSRecordInput) (string, error) {
	return "rec-1", nil
}
func (s *stubProvider) UpdateRecord(_ context.Context, _, _ string, _ domainport.DNSRecordInput) error {
	return nil
}
func (s *stubProvider) DeleteRecord(_ context.Context, _, _ string) error { return nil }
func (s *stubProvider) CreateBatch(_ context.Context, _ string, inputs []domainport.DNSRecordInput) ([]domainport.BatchResult, error) {
	return nil, nil
}
func (s *stubProvider) GetRecord(_ context.Context, _, _ string) (*domainport.DNSRecord, error) {
	return nil, nil
}
func (s *stubProvider) FindRecords(_ context.Context, _, _, _, _ string) ([]domainport.DNSRecord, error) {
	return nil, nil
}
func (s *stubProvider) PauseRecord(_ context.Context, _, _ string) error  { return nil }
func (s *stubProvider) ResumeRecord(_ context.Context, _, _ string) error { return nil }
func (s *stubProvider) GetNameservers(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestRegistry_GetDNS_Found(t *testing.T) {
	reg := provider.NewRegistry()
	reg.Register(domainport.ProviderAliyun, func(_ []byte) (domainport.DNSProvider, error) {
		return &stubProvider{pt: domainport.ProviderAliyun}, nil
	})

	p, err := reg.GetDNS(domainport.ProviderAliyun)
	if err != nil {
		t.Fatalf("expected provider, got error: %v", err)
	}
	if p.ProviderType() != domainport.ProviderAliyun {
		t.Errorf("unexpected provider type: %s", p.ProviderType())
	}
}

func TestRegistry_GetDNS_NotFound(t *testing.T) {
	reg := provider.NewRegistry()
	_, err := reg.GetDNS("nonexistent")
	if !errors.Is(err, domainport.ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestRegistry_GetDNSWithCredentials(t *testing.T) {
	reg := provider.NewRegistry()
	var capturedCreds []byte
	reg.Register(domainport.ProviderDNSPod, func(creds []byte) (domainport.DNSProvider, error) {
		capturedCreds = creds
		return &stubProvider{pt: domainport.ProviderDNSPod}, nil
	})

	creds := []byte(`{"token":"abc123"}`)
	p, err := reg.GetDNSWithCredentials(domainport.ProviderDNSPod, creds)
	if err != nil {
		t.Fatal(err)
	}
	if p.ProviderType() != domainport.ProviderDNSPod {
		t.Errorf("unexpected type: %s", p.ProviderType())
	}
	if string(capturedCreds) != string(creds) {
		t.Errorf("credentials not passed: got %s", capturedCreds)
	}
}
