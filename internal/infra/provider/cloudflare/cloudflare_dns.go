package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cf "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"

	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/ratelimit"
)

const (
	// Cloudflare API: 1200 requests per 5 minutes = 240/min
	cloudflareRateLimit = 240
)

// Credentials holds the JSON structure for Cloudflare accounts.
type Credentials struct {
	APIToken string `json:"api_token"`
}

// CloudflareDNSProvider implements port.DNSProvider for Cloudflare DNS.
type CloudflareDNSProvider struct {
	client  *cf.Client
	limiter *ratelimit.TokenBucket
}

// NewCloudflareDNSProvider creates a provider from raw credential JSON.
func NewCloudflareDNSProvider(credentialsJSON []byte) (port.DNSProvider, error) {
	var creds Credentials
	if err := json.Unmarshal(credentialsJSON, &creds); err != nil {
		return nil, fmt.Errorf("cloudflare: invalid credentials JSON: %w", err)
	}
	if creds.APIToken == "" {
		return nil, fmt.Errorf("cloudflare: api_token is required")
	}

	client := cf.NewClient(option.WithAPIToken(creds.APIToken))
	return &CloudflareDNSProvider{
		client:  client,
		limiter: ratelimit.NewTokenBucket("cloudflare-dns", cloudflareRateLimit),
	}, nil
}

func (p *CloudflareDNSProvider) ProviderType() port.ProviderType {
	return port.ProviderCloudflare
}

func (p *CloudflareDNSProvider) ValidateCredentials(ctx context.Context) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	iter := p.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{
		PerPage: cf.F(float64(1)),
	})
	if err := iter.Err(); err != nil {
		if isAuthError(err) {
			return fmt.Errorf("invalid credentials: %w", err)
		}
	}
	return nil
}

func (p *CloudflareDNSProvider) DetectCapabilities(_ context.Context) (*port.Capabilities, error) {
	return &port.Capabilities{
		ISPRouting:       false,
		Lines:            []string{"default"},
		MinTTL:           1,
		MaxRecordsPerSet: 1000,
		RawPlanName:      "cloudflare",
	}, nil
}

func (p *CloudflareDNSProvider) ListZones(ctx context.Context) ([]port.DNSZone, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	var result []port.DNSZone
	iter := p.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{})
	for iter.Next() {
		z := iter.Current()
		result = append(result, port.DNSZone{
			ID:     z.ID,
			Domain: z.Name,
		})
	}
	return result, iter.Err()
}

func (p *CloudflareDNSProvider) ListRecords(ctx context.Context, zoneID string) ([]port.DNSRecord, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	var result []port.DNSRecord
	iter := p.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID: cf.F(zoneID),
	})
	for iter.Next() {
		r := iter.Current()
		result = append(result, recordFromCF(r))
	}
	return result, iter.Err()
}

func (p *CloudflareDNSProvider) CreateRecord(ctx context.Context, zoneID string, input port.DNSRecordInput) (string, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return "", err
	}
	r, err := p.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cf.F(zoneID),
		Body: dns.ARecordParam{
			Name:    cf.F(input.Host),
			Type:    cf.F(dns.ARecordType("A")),
			Content: cf.F(input.Value),
			TTL:     cf.F(dns.TTL(input.TTL)),
			Comment: cf.F("managed by SmartCDN"),
		},
	})
	if err != nil {
		return "", fmt.Errorf("cloudflare CreateRecord: %w", err)
	}
	return r.ID, nil
}

func (p *CloudflareDNSProvider) UpdateRecord(ctx context.Context, zoneID, recordID string, input port.DNSRecordInput) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	_, err := p.client.DNS.Records.Update(ctx, recordID, dns.RecordUpdateParams{
		ZoneID: cf.F(zoneID),
		Body: dns.ARecordParam{
			Name:    cf.F(input.Host),
			Type:    cf.F(dns.ARecordType("A")),
			Content: cf.F(input.Value),
			TTL:     cf.F(dns.TTL(input.TTL)),
		},
	})
	return err
}

func (p *CloudflareDNSProvider) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	_, err := p.client.DNS.Records.Delete(ctx, recordID, dns.RecordDeleteParams{
		ZoneID: cf.F(zoneID),
	})
	return err
}

func (p *CloudflareDNSProvider) CreateBatch(ctx context.Context, zoneID string, inputs []port.DNSRecordInput) ([]port.BatchResult, error) {
	results := make([]port.BatchResult, len(inputs))
	for i, input := range inputs {
		id, err := p.CreateRecord(ctx, zoneID, input)
		results[i] = port.BatchResult{RecordID: id, Err: err}
	}
	return results, nil
}

func (p *CloudflareDNSProvider) GetRecord(ctx context.Context, zoneID, recordID string) (*port.DNSRecord, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	r, err := p.client.DNS.Records.Get(ctx, recordID, dns.RecordGetParams{
		ZoneID: cf.F(zoneID),
	})
	if err != nil {
		return nil, err
	}
	rec := recordFromCF(*r)
	return &rec, nil
}

func (p *CloudflareDNSProvider) FindRecords(ctx context.Context, zoneID, host, recordType, _ string) ([]port.DNSRecord, error) {
	all, err := p.ListRecords(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	var out []port.DNSRecord
	for _, r := range all {
		if host != "" && r.Host != host {
			continue
		}
		if recordType != "" && !strings.EqualFold(r.Type, recordType) {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (p *CloudflareDNSProvider) PauseRecord(_ context.Context, _, _ string) error {
	return &port.UnsupportedCapabilityError{
		Operation: "PauseRecord",
		Reason:    "Cloudflare does not support pausing individual DNS records",
	}
}

func (p *CloudflareDNSProvider) ResumeRecord(_ context.Context, _, _ string) error {
	return &port.UnsupportedCapabilityError{
		Operation: "ResumeRecord",
		Reason:    "Cloudflare does not support pausing individual DNS records",
	}
}

func (p *CloudflareDNSProvider) GetNameservers(ctx context.Context, zoneID string) ([]string, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	zone, err := p.client.Zones.Get(ctx, zones.ZoneGetParams{ZoneID: cf.F(zoneID)})
	if err != nil {
		return nil, err
	}
	ns := make([]string, len(zone.NameServers))
	copy(ns, zone.NameServers)
	return ns, nil
}

// --- helpers ---

func recordFromCF(r dns.RecordResponse) port.DNSRecord {
	return port.DNSRecord{
		ID:    r.ID,
		Host:  r.Name,
		Type:  string(r.Type),
		Value: r.Content,
		TTL:   int(r.TTL),
	}
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "9103") ||
		strings.Contains(msg, "10000") ||
		strings.Contains(msg, "Unauthorized")
}
