package cfcdn

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/cache"
	"github.com/cloudflare/cloudflare-go/v4/custom_hostnames"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// credentials JSON shape expected from the CDNAccount.CredentialsEnc.
type cfCredentials struct {
	APIToken string `json:"api_token"`
}

// CloudflareCDNProvider implements port.CDNProvider using the cloudflare-go/v4 SDK.
// Cache rules use the native Cloudflare Cache Purge API; full page-rule management
// requires Cloudflare Pro plan and is handled by the PurgeURLs/PurgeAll methods.
type CloudflareCDNProvider struct {
	client *cloudflare.Client
}

// NewCloudflareCDNProvider is the factory function registered with CDNRegistry.
func NewCloudflareCDNProvider(credentials []byte) (port.CDNProvider, error) {
	var creds cfCredentials
	if credentials != nil {
		if err := json.Unmarshal(credentials, &creds); err != nil {
			return nil, fmt.Errorf("cloudflare cdn: parse credentials: %w", err)
		}
	}
	client := cloudflare.NewClient(option.WithAPIToken(creds.APIToken))
	return &CloudflareCDNProvider{client: client}, nil
}

func (p *CloudflareCDNProvider) ProviderType() port.CDNProviderType {
	return port.CDNProviderCloudflare
}

func (p *CloudflareCDNProvider) ValidateCredentials(ctx context.Context) error {
	iter := p.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{})
	iter.Next() // just attempt a page; error surfaces here
	return iter.Err()
}

// --- Domain / Zone lifecycle ---

func (p *CloudflareCDNProvider) AddDomain(ctx context.Context, input port.CDNDomainInput) (*port.CDNZone, error) {
	resp, err := p.client.Zones.New(ctx, zones.ZoneNewParams{
		Account: cloudflare.F(zones.ZoneNewParamsAccount{}),
		Name:    cloudflare.F(input.Domain),
		Type:    cloudflare.F(zones.TypeFull),
	})
	if err != nil {
		return nil, fmt.Errorf("cloudflare cdn: add zone %q: %w", input.Domain, err)
	}
	return &port.CDNZone{
		ID:     resp.ID,
		Domain: resp.Name,
		Status: string(resp.Status),
	}, nil
}

func (p *CloudflareCDNProvider) GetDomain(ctx context.Context, domainName string) (*port.CDNZone, error) {
	zoneID, err := p.findZoneID(ctx, domainName)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Zones.Get(ctx, zones.ZoneGetParams{ZoneID: cloudflare.F(zoneID)})
	if err != nil {
		return nil, fmt.Errorf("cloudflare cdn: get zone %q: %w", domainName, err)
	}
	return &port.CDNZone{
		ID:     resp.ID,
		Domain: resp.Name,
		Status: string(resp.Status),
	}, nil
}

func (p *CloudflareCDNProvider) DeleteDomain(ctx context.Context, domainName string) error {
	zoneID, err := p.findZoneID(ctx, domainName)
	if err != nil {
		return err
	}
	_, err = p.client.Zones.Delete(ctx, zones.ZoneDeleteParams{ZoneID: cloudflare.F(zoneID)})
	return err
}

func (p *CloudflareCDNProvider) ListDomains(ctx context.Context) ([]port.CDNZone, error) {
	var result []port.CDNZone
	iter := p.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{})
	for iter.Next() {
		z := iter.Current()
		result = append(result, port.CDNZone{
			ID:     z.ID,
			Domain: z.Name,
			Status: string(z.Status),
		})
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("cloudflare cdn: list zones: %w", err)
	}
	return result, nil
}

// --- Custom Hostname / SaaS ---

func (p *CloudflareCDNProvider) CreateCustomHostname(ctx context.Context, zoneID string, input port.CustomHostnameInput) (*port.CustomHostname, error) {
	resp, err := p.client.CustomHostnames.New(ctx, custom_hostnames.CustomHostnameNewParams{
		ZoneID:   cloudflare.F(zoneID),
		Hostname: cloudflare.F(input.Hostname),
		SSL: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSL{
			Method: cloudflare.F(custom_hostnames.DCVMethodTXT),
			Type:   cloudflare.F(custom_hostnames.DomainValidationTypeDv),
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("cloudflare cdn: create custom hostname %q: %w", input.Hostname, err)
	}
	ch := &port.CustomHostname{
		ID:       resp.ID,
		Hostname: resp.Hostname,
		Status:   string(resp.Status),
	}
	if resp.OwnershipVerification.Name != "" {
		ch.DCVTXTName = resp.OwnershipVerification.Name
		ch.DCVTXTValue = resp.OwnershipVerification.Value
	}
	if resp.OwnershipVerificationHTTP.HTTPURL != "" {
		ch.HTTPTokenPath = resp.OwnershipVerificationHTTP.HTTPURL
		ch.HTTPTokenValue = resp.OwnershipVerificationHTTP.HTTPBody
	}
	ch.SSLStatus = string(resp.SSL.Status)
	return ch, nil
}

func (p *CloudflareCDNProvider) GetCustomHostname(ctx context.Context, zoneID, hostnameID string) (*port.CustomHostname, error) {
	resp, err := p.client.CustomHostnames.Get(ctx, hostnameID, custom_hostnames.CustomHostnameGetParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("cloudflare cdn: get custom hostname %q: %w", hostnameID, err)
	}
	ch := &port.CustomHostname{
		ID:       resp.ID,
		Hostname: resp.Hostname,
		Status:   string(resp.Status),
	}
	if resp.OwnershipVerification.Name != "" {
		ch.DCVTXTName = resp.OwnershipVerification.Name
		ch.DCVTXTValue = resp.OwnershipVerification.Value
	}
	ch.SSLStatus = string(resp.SSL.Status)
	return ch, nil
}

func (p *CloudflareCDNProvider) DeleteCustomHostname(ctx context.Context, zoneID, hostnameID string) error {
	_, err := p.client.CustomHostnames.Delete(ctx, hostnameID, custom_hostnames.CustomHostnameDeleteParams{
		ZoneID: cloudflare.F(zoneID),
	})
	return err
}

func (p *CloudflareCDNProvider) ListCustomHostnames(ctx context.Context, zoneID string) ([]port.CustomHostname, error) {
	var result []port.CustomHostname
	iter := p.client.CustomHostnames.ListAutoPaging(ctx, custom_hostnames.CustomHostnameListParams{
		ZoneID: cloudflare.F(zoneID),
	})
	for iter.Next() {
		h := iter.Current()
		result = append(result, port.CustomHostname{
			ID:        h.ID,
			Hostname:  h.Hostname,
			Status:    string(h.Status),
			SSLStatus: string(h.SSL.Status),
		})
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("cloudflare cdn: list custom hostnames: %w", err)
	}
	return result, nil
}

// --- Cache ---

func (p *CloudflareCDNProvider) PurgeURLs(ctx context.Context, domainName string, urls []string) error {
	zoneID, err := p.findZoneID(ctx, domainName)
	if err != nil {
		return err
	}
	_, err = p.client.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(zoneID),
		Body:   cache.CachePurgeParamsBodyCachePurgeSingleFile{Files: cloudflare.F(urls)},
	})
	return err
}

func (p *CloudflareCDNProvider) PurgeAll(ctx context.Context, domainName string) error {
	zoneID, err := p.findZoneID(ctx, domainName)
	if err != nil {
		return err
	}
	_, err = p.client.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(zoneID),
		Body:   cache.CachePurgeParamsBodyCachePurgeEverything{PurgeEverything: cloudflare.F(true)},
	})
	return err
}

// --- Cache Rules ---
// Cloudflare cache rules use the Rulesets API (Business/Enterprise plan).
// For Free/Pro, PurgeAll and PurgeURLs are supported; rule management returns
// UnsupportedCapabilityError unless the zone has the required plan.

func (p *CloudflareCDNProvider) SetCacheRule(_ context.Context, _ string, _ port.CacheRuleInput) (*port.CacheRule, error) {
	return nil, &port.UnsupportedCapabilityError{
		Operation: "SetCacheRule",
		Reason:    "Cloudflare cache rules require Rulesets API (Business/Enterprise plan); use PurgeURLs/PurgeAll for cache management",
	}
}

func (p *CloudflareCDNProvider) ListCacheRules(_ context.Context, _ string) ([]port.CacheRule, error) {
	return nil, &port.UnsupportedCapabilityError{
		Operation: "ListCacheRules",
		Reason:    "use Cloudflare dashboard or Rulesets API directly",
	}
}

func (p *CloudflareCDNProvider) DeleteCacheRule(_ context.Context, _, _ string) error {
	return &port.UnsupportedCapabilityError{
		Operation: "DeleteCacheRule",
		Reason:    "use Cloudflare dashboard or Rulesets API directly",
	}
}

// --- helper ---

func (p *CloudflareCDNProvider) findZoneID(ctx context.Context, domainName string) (string, error) {
	iter := p.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{
		Name: cloudflare.F(domainName),
	})
	for iter.Next() {
		z := iter.Current()
		if z.Name == domainName {
			return z.ID, nil
		}
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("cloudflare cdn: list zones: %w", err)
	}
	return "", fmt.Errorf("cloudflare cdn: zone not found for %q", domainName)
}
