package aliyunesa

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	esa "github.com/alibabacloud-go/esa-20240910/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// credentials JSON shape expected from the CDNAccount.CredentialsEnc.
type esaCredentials struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
}

// AliyunESAProvider implements port.CDNProvider using the Alibaba Cloud ESA SDK.
// ESA (Edge Security Acceleration) is Alibaba's next-gen CDN with DDoS/WAF/Bot
// protection — functionally similar to Cloudflare.
type AliyunESAProvider struct {
	client *esa.Client
}

// NewAliyunESAProvider is the factory function registered with CDNRegistry.
func NewAliyunESAProvider(credentials []byte) (port.CDNProvider, error) {
	var creds esaCredentials
	if credentials != nil {
		if err := json.Unmarshal(credentials, &creds); err != nil {
			return nil, fmt.Errorf("aliyun esa: parse credentials: %w", err)
		}
	}
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		Endpoint:        tea.String("esa.cn-hangzhou.aliyuncs.com"),
	}
	c, err := esa.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("aliyun esa: create client: %w", err)
	}
	return &AliyunESAProvider{client: c}, nil
}

func (p *AliyunESAProvider) ProviderType() port.CDNProviderType {
	return port.CDNProviderAliyunESA
}

func (p *AliyunESAProvider) ValidateCredentials(ctx context.Context) error {
	req := &esa.ListSitesRequest{
		PageSize:   tea.Int32(1),
		PageNumber: tea.Int32(1),
	}
	_, err := p.client.ListSites(req)
	if err != nil {
		return fmt.Errorf("aliyun esa: validate credentials: %w", err)
	}
	return nil
}

// --- Domain / Site lifecycle ---

func (p *AliyunESAProvider) AddDomain(_ context.Context, input port.CDNDomainInput) (*port.CDNZone, error) {
	req := &esa.CreateSiteRequest{
		SiteName:   tea.String(input.Domain),
		AccessType: tea.String("NS"), // NS delegation mode; "CNAME" for partial zone
		Coverage:   tea.String("global"),
	}
	resp, err := p.client.CreateSite(req)
	if err != nil {
		return nil, fmt.Errorf("aliyun esa: create site %q: %w", input.Domain, err)
	}
	siteID := int64(0)
	if resp.Body != nil && resp.Body.SiteId != nil {
		siteID = *resp.Body.SiteId
	}
	return &port.CDNZone{
		ID:     strconv.FormatInt(siteID, 10),
		Domain: input.Domain,
		Status: "pending",
	}, nil
}

func (p *AliyunESAProvider) GetDomain(_ context.Context, domainName string) (*port.CDNZone, error) {
	site, err := p.findSite(domainName)
	if err != nil {
		return nil, err
	}
	return site, nil
}

func (p *AliyunESAProvider) DeleteDomain(_ context.Context, domainName string) error {
	site, err := p.findSite(domainName)
	if err != nil {
		return err
	}
	siteID, _ := strconv.ParseInt(site.ID, 10, 64)
	_, err = p.client.DeleteSite(&esa.DeleteSiteRequest{
		SiteId: tea.Int64(siteID),
	})
	if err != nil {
		return fmt.Errorf("aliyun esa: delete site %q: %w", domainName, err)
	}
	return nil
}

func (p *AliyunESAProvider) ListDomains(_ context.Context) ([]port.CDNZone, error) {
	var result []port.CDNZone
	pageNum := int32(1)
	for {
		resp, err := p.client.ListSites(&esa.ListSitesRequest{
			PageNumber: tea.Int32(pageNum),
			PageSize:   tea.Int32(50),
		})
		if err != nil {
			return nil, fmt.Errorf("aliyun esa: list sites: %w", err)
		}
		if resp.Body == nil {
			break
		}
		for _, s := range resp.Body.Sites {
			zone := port.CDNZone{Domain: tea.StringValue(s.SiteName)}
			if s.SiteId != nil {
				zone.ID = strconv.FormatInt(*s.SiteId, 10)
			}
			if s.CnameZone != nil {
				zone.CNAME = *s.CnameZone
			}
			if s.Status != nil {
				zone.Status = *s.Status
			}
			result = append(result, zone)
		}
		// pagination: stop when we got fewer than a full page
		if len(resp.Body.Sites) < 50 {
			break
		}
		pageNum++
	}
	return result, nil
}

// --- Custom Hostnames (ESA doesn't support Cloudflare SaaS model) ---

func (p *AliyunESAProvider) CreateCustomHostname(_ context.Context, _ string, _ port.CustomHostnameInput) (*port.CustomHostname, error) {
	return nil, &port.UnsupportedCapabilityError{Operation: "CreateCustomHostname", Reason: "ESA does not support Cloudflare SaaS custom hostnames"}
}

func (p *AliyunESAProvider) GetCustomHostname(_ context.Context, _, _ string) (*port.CustomHostname, error) {
	return nil, &port.UnsupportedCapabilityError{Operation: "GetCustomHostname", Reason: "ESA does not support Cloudflare SaaS custom hostnames"}
}

func (p *AliyunESAProvider) DeleteCustomHostname(_ context.Context, _, _ string) error {
	return &port.UnsupportedCapabilityError{Operation: "DeleteCustomHostname", Reason: "ESA does not support Cloudflare SaaS custom hostnames"}
}

func (p *AliyunESAProvider) ListCustomHostnames(_ context.Context, _ string) ([]port.CustomHostname, error) {
	return nil, &port.UnsupportedCapabilityError{Operation: "ListCustomHostnames", Reason: "ESA does not support Cloudflare SaaS custom hostnames"}
}

// --- Cache purge ---

func (p *AliyunESAProvider) PurgeURLs(_ context.Context, domainName string, urls []string) error {
	site, err := p.findSite(domainName)
	if err != nil {
		return err
	}
	siteID, _ := strconv.ParseInt(site.ID, 10, 64)

	// Convert []string to []interface{} for Files field
	files := make([]interface{}, len(urls))
	for i, u := range urls {
		files[i] = u
	}
	_, err = p.client.PurgeCaches(&esa.PurgeCachesRequest{
		SiteId: tea.Int64(siteID),
		Type:   tea.String("File"),
		Content: &esa.PurgeCachesRequestContent{
			Files: files,
		},
	})
	if err != nil {
		return fmt.Errorf("aliyun esa: purge urls for %q: %w", domainName, err)
	}
	return nil
}

func (p *AliyunESAProvider) PurgeAll(_ context.Context, domainName string) error {
	site, err := p.findSite(domainName)
	if err != nil {
		return err
	}
	siteID, _ := strconv.ParseInt(site.ID, 10, 64)
	_, err = p.client.PurgeCaches(&esa.PurgeCachesRequest{
		SiteId: tea.Int64(siteID),
		Type:   tea.String("All"),
		Content: &esa.PurgeCachesRequestContent{
			PurgeAll: tea.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("aliyun esa: purge all for %q: %w", domainName, err)
	}
	return nil
}

// --- Cache rules ---

func (p *AliyunESAProvider) SetCacheRule(_ context.Context, domainName string, rule port.CacheRuleInput) (*port.CacheRule, error) {
	site, err := p.findSite(domainName)
	if err != nil {
		return nil, err
	}
	siteID, _ := strconv.ParseInt(site.ID, 10, 64)
	ttlStr := strconv.Itoa(rule.TTL)
	resp, err := p.client.CreateCacheRule(&esa.CreateCacheRuleRequest{
		SiteId:       tea.Int64(siteID),
		Rule:         tea.String(fmt.Sprintf(`http.request.uri.path matches "%s"`, rule.PathPattern)),
		RuleName:     tea.String(rule.PathPattern),
		EdgeCacheTtl: tea.String(ttlStr),
		RuleEnable:   tea.String("on"),
	})
	if err != nil {
		return nil, fmt.Errorf("aliyun esa: create cache rule: %w", err)
	}
	configID := int64(0)
	if resp.Body != nil && resp.Body.ConfigId != nil {
		configID = *resp.Body.ConfigId
	}
	return &port.CacheRule{
		ID:          strconv.FormatInt(configID, 10),
		PathPattern: rule.PathPattern,
		TTL:         rule.TTL,
	}, nil
}

func (p *AliyunESAProvider) ListCacheRules(_ context.Context, domainName string) ([]port.CacheRule, error) {
	site, err := p.findSite(domainName)
	if err != nil {
		return nil, err
	}
	siteID, _ := strconv.ParseInt(site.ID, 10, 64)
	resp, err := p.client.ListCacheRules(&esa.ListCacheRulesRequest{
		SiteId: tea.Int64(siteID),
	})
	if err != nil {
		return nil, fmt.Errorf("aliyun esa: list cache rules: %w", err)
	}
	var result []port.CacheRule
	if resp.Body != nil {
		for _, r := range resp.Body.Configs {
			cr := port.CacheRule{}
			if r.ConfigId != nil {
				cr.ID = strconv.FormatInt(*r.ConfigId, 10)
			}
			if r.RuleName != nil {
				cr.PathPattern = *r.RuleName
			}
			if r.EdgeCacheTtl != nil {
				cr.TTL, _ = strconv.Atoi(*r.EdgeCacheTtl)
			}
			result = append(result, cr)
		}
	}
	return result, nil
}

func (p *AliyunESAProvider) DeleteCacheRule(_ context.Context, domainName, ruleID string) error {
	site, err := p.findSite(domainName)
	if err != nil {
		return err
	}
	siteID, _ := strconv.ParseInt(site.ID, 10, 64)
	configID, _ := strconv.ParseInt(ruleID, 10, 64)
	_, err = p.client.DeleteCacheRule(&esa.DeleteCacheRuleRequest{
		SiteId:   tea.Int64(siteID),
		ConfigId: tea.Int64(configID),
	})
	if err != nil {
		return fmt.Errorf("aliyun esa: delete cache rule %q: %w", ruleID, err)
	}
	return nil
}

// --- helper ---

func (p *AliyunESAProvider) findSite(domainName string) (*port.CDNZone, error) {
	resp, err := p.client.ListSites(&esa.ListSitesRequest{
		SiteName:   tea.String(domainName),
		PageNumber: tea.Int32(1),
		PageSize:   tea.Int32(10),
	})
	if err != nil {
		return nil, fmt.Errorf("aliyun esa: list sites for %q: %w", domainName, err)
	}
	if resp.Body == nil || len(resp.Body.Sites) == 0 {
		return nil, fmt.Errorf("aliyun esa: site not found: %q", domainName)
	}
	for _, s := range resp.Body.Sites {
		if tea.StringValue(s.SiteName) == domainName {
			zone := &port.CDNZone{
				Domain: domainName,
				Status: tea.StringValue(s.Status),
			}
			if s.SiteId != nil {
				zone.ID = strconv.FormatInt(*s.SiteId, 10)
			}
			if s.CnameZone != nil {
				zone.CNAME = *s.CnameZone
			}
			return zone, nil
		}
	}
	return nil, fmt.Errorf("aliyun esa: site not found: %q", domainName)
}
