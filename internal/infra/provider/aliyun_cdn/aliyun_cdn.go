package aliyuncdn

import (
	"context"
	"encoding/json"
	"fmt"

	cdn "github.com/alibabacloud-go/cdn-20180510/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// credentials JSON shape expected from the CDNAccount.CredentialsEnc.
type aliyunCredentials struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Region          string `json:"region,omitempty"` // default: cn-hangzhou
}

// cdnOriginSource is the JSON structure for the Sources parameter in AddCdnDomain.
type cdnOriginSource struct {
	Content  string `json:"content"`  // IP or hostname
	Type     string `json:"type"`     // "ipaddr" | "domain" | "oss"
	Priority string `json:"priority"` // "20" = primary, "30" = backup
	Port     int    `json:"port"`
	Weight   string `json:"weight"`
}

// AliyunCDNProvider implements port.CDNProvider using the Alibaba Cloud CDN SDK.
type AliyunCDNProvider struct {
	client *cdn.Client
}

// NewAliyunCDNProvider is the factory function registered with CDNRegistry.
func NewAliyunCDNProvider(credentials []byte) (port.CDNProvider, error) {
	var creds aliyunCredentials
	if credentials != nil {
		if err := json.Unmarshal(credentials, &creds); err != nil {
			return nil, fmt.Errorf("aliyun cdn: parse credentials: %w", err)
		}
	}
	region := creds.Region
	if region == "" {
		region = "cn-hangzhou"
	}
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		Endpoint:        tea.String("cdn.aliyuncs.com"),
	}
	c, err := cdn.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("aliyun cdn: init client: %w", err)
	}
	return &AliyunCDNProvider{client: c}, nil
}

func (p *AliyunCDNProvider) ProviderType() port.CDNProviderType {
	return port.CDNProviderAliyun
}

func (p *AliyunCDNProvider) ValidateCredentials(_ context.Context) error {
	req := &cdn.DescribeUserDomainsRequest{PageSize: tea.Int32(1)}
	_, err := p.client.DescribeUserDomainsWithOptions(req, &util.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("aliyun cdn: validate credentials: %w", err)
	}
	return nil
}

// --- Domain lifecycle ---

func (p *AliyunCDNProvider) AddDomain(_ context.Context, input port.CDNDomainInput) (*port.CDNZone, error) {
	sourceType := "ipaddr"
	content := input.OriginAddr
	// If origin looks like a hostname (contains letters), use domain type.
	for _, c := range input.OriginAddr {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			sourceType = "domain"
			break
		}
	}
	port_ := input.OriginPort
	if port_ == 0 {
		port_ = 80
	}
	sourcesJSON, err := json.Marshal([]cdnOriginSource{{
		Content:  content,
		Type:     sourceType,
		Priority: "20",
		Port:     port_,
		Weight:   "10",
	}})
	if err != nil {
		return nil, fmt.Errorf("aliyun cdn: marshal sources: %w", err)
	}
	req := &cdn.AddCdnDomainRequest{
		DomainName: tea.String(input.Domain),
		CdnType:    tea.String("web"),
		Sources:    tea.String(string(sourcesJSON)),
		Scope:      tea.String("domestic"),
	}
	_, err = p.client.AddCdnDomainWithOptions(req, &util.RuntimeOptions{})
	if err != nil {
		return nil, fmt.Errorf("aliyun cdn: add domain %q: %w", input.Domain, err)
	}
	// Aliyun CDN AddDomain returns no zone ID; use domain name as ID.
	return &port.CDNZone{
		ID:     input.Domain,
		Domain: input.Domain,
		Status: "configuring",
	}, nil
}

func (p *AliyunCDNProvider) GetDomain(_ context.Context, domainName string) (*port.CDNZone, error) {
	req := &cdn.DescribeCdnDomainDetailRequest{DomainName: tea.String(domainName)}
	resp, err := p.client.DescribeCdnDomainDetailWithOptions(req, &util.RuntimeOptions{})
	if err != nil {
		return nil, fmt.Errorf("aliyun cdn: get domain %q: %w", domainName, err)
	}
	if resp.Body == nil || resp.Body.GetDomainDetailModel == nil {
		return nil, fmt.Errorf("aliyun cdn: empty response for %q", domainName)
	}
	detail := resp.Body.GetDomainDetailModel
	cname := ""
	if detail.Cname != nil {
		cname = *detail.Cname
	}
	status := ""
	if detail.DomainStatus != nil {
		status = *detail.DomainStatus
	}
	return &port.CDNZone{
		ID:     domainName,
		Domain: domainName,
		CNAME:  cname,
		Status: status,
	}, nil
}

func (p *AliyunCDNProvider) DeleteDomain(_ context.Context, domainName string) error {
	req := &cdn.DeleteCdnDomainRequest{DomainName: tea.String(domainName)}
	_, err := p.client.DeleteCdnDomainWithOptions(req, &util.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("aliyun cdn: delete domain %q: %w", domainName, err)
	}
	return nil
}

func (p *AliyunCDNProvider) ListDomains(_ context.Context) ([]port.CDNZone, error) {
	var result []port.CDNZone
	pageNum := int32(1)
	for {
		req := &cdn.DescribeUserDomainsRequest{
			PageSize:   tea.Int32(50),
			PageNumber: tea.Int32(pageNum),
		}
		resp, err := p.client.DescribeUserDomainsWithOptions(req, &util.RuntimeOptions{})
		if err != nil {
			return nil, fmt.Errorf("aliyun cdn: list domains: %w", err)
		}
		if resp.Body == nil || resp.Body.Domains == nil {
			break
		}
		for _, d := range resp.Body.Domains.PageData {
			cname := ""
			if d.Cname != nil {
				cname = *d.Cname
			}
			status := ""
			if d.DomainStatus != nil {
				status = *d.DomainStatus
			}
			name := ""
			if d.DomainName != nil {
				name = *d.DomainName
			}
			result = append(result, port.CDNZone{
				ID:     name,
				Domain: name,
				CNAME:  cname,
				Status: status,
			})
		}
		total := int64(0)
		if resp.Body.TotalCount != nil {
			total = *resp.Body.TotalCount
		}
		if int64(len(result)) >= total {
			break
		}
		pageNum++
	}
	return result, nil
}

// --- Cache ---

func (p *AliyunCDNProvider) PurgeURLs(_ context.Context, _ string, urls []string) error {
	objectPath := ""
	for i, u := range urls {
		if i > 0 {
			objectPath += "\n"
		}
		objectPath += u
	}
	req := &cdn.RefreshObjectCachesRequest{
		ObjectPath: tea.String(objectPath),
		ObjectType: tea.String("File"),
	}
	_, err := p.client.RefreshObjectCachesWithOptions(req, &util.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("aliyun cdn: purge URLs: %w", err)
	}
	return nil
}

func (p *AliyunCDNProvider) PurgeAll(_ context.Context, domainName string) error {
	// Aliyun CDN does not support purge-all by API; purge directory "/" as best effort.
	req := &cdn.RefreshObjectCachesRequest{
		ObjectPath: tea.String(fmt.Sprintf("http://%s/", domainName)),
		ObjectType: tea.String("Directory"),
	}
	_, err := p.client.RefreshObjectCachesWithOptions(req, &util.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("aliyun cdn: purge all (directory flush) %q: %w", domainName, err)
	}
	return nil
}

// --- Cache rules ---
// Aliyun CDN cache rules are managed through the BatchSetCdnDomainConfig API
// using function configurations. Full implementation requires Enterprise plan.

func (p *AliyunCDNProvider) SetCacheRule(_ context.Context, _ string, _ port.CacheRuleInput) (*port.CacheRule, error) {
	return nil, &port.UnsupportedCapabilityError{
		Operation: "SetCacheRule",
		Reason:    "Aliyun CDN cache rules require BatchSetCdnDomainConfig API; use PurgeURLs/PurgeAll for cache management",
	}
}

func (p *AliyunCDNProvider) ListCacheRules(_ context.Context, _ string) ([]port.CacheRule, error) {
	return nil, &port.UnsupportedCapabilityError{
		Operation: "ListCacheRules",
		Reason:    "use Alibaba Cloud CDN console or DescribeCdnDomainConfigs API directly",
	}
}

func (p *AliyunCDNProvider) DeleteCacheRule(_ context.Context, _, _ string) error {
	return &port.UnsupportedCapabilityError{
		Operation: "DeleteCacheRule",
		Reason:    "use Alibaba Cloud CDN console or DeleteSpecificStagingConfig API directly",
	}
}

// --- Custom Hostnames (SaaS) ---
// Aliyun CDN does not support Cloudflare-style SaaS custom hostnames.

func (p *AliyunCDNProvider) CreateCustomHostname(_ context.Context, _ string, _ port.CustomHostnameInput) (*port.CustomHostname, error) {
	return nil, &port.UnsupportedCapabilityError{
		Operation: "CreateCustomHostname",
		Reason:    "Aliyun CDN does not support SaaS custom hostnames; use Cloudflare for this feature",
	}
}

func (p *AliyunCDNProvider) GetCustomHostname(_ context.Context, _, _ string) (*port.CustomHostname, error) {
	return nil, &port.UnsupportedCapabilityError{
		Operation: "GetCustomHostname",
		Reason:    "Aliyun CDN does not support SaaS custom hostnames",
	}
}

func (p *AliyunCDNProvider) DeleteCustomHostname(_ context.Context, _, _ string) error {
	return &port.UnsupportedCapabilityError{
		Operation: "DeleteCustomHostname",
		Reason:    "Aliyun CDN does not support SaaS custom hostnames",
	}
}

func (p *AliyunCDNProvider) ListCustomHostnames(_ context.Context, _ string) ([]port.CustomHostname, error) {
	return nil, &port.UnsupportedCapabilityError{
		Operation: "ListCustomHostnames",
		Reason:    "Aliyun CDN does not support SaaS custom hostnames",
	}
}
