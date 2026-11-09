package port

import "context"

// CDNProviderType names a CDN provider implementation.
type CDNProviderType string

const (
	CDNProviderCloudflare CDNProviderType = "cloudflare"
	CDNProviderAliyun     CDNProviderType = "aliyun_cdn"
)

// CDNZone represents a CDN-accelerated domain.
type CDNZone struct {
	ID     string // provider-side zone/domain ID
	Domain string // apex domain, e.g. "example.com"
	CNAME  string // CDN CNAME to point to, e.g. "example.com.cdn.cloudflare.net"
	Status string // "active" | "pending" | "disabled"
}

// CDNDomainInput is used to add a CDN domain.
type CDNDomainInput struct {
	Domain     string // domain to accelerate
	OriginAddr string // origin server IP or hostname
	OriginPort int    // 80 or 443
	Protocol   string // "http" | "https" | "follow"
}

// CustomHostname represents a Cloudflare SaaS custom hostname.
type CustomHostname struct {
	ID             string
	Hostname       string
	Status         string   // "active" | "pending" | "pending_validation" | "blocked"
	DCVTXTName     string   // TXT record name for DCV
	DCVTXTValue    string   // TXT record value for DCV
	HTTPTokenPath  string   // HTTP DCV token path (alternative to TXT)
	HTTPTokenValue string
	SSLStatus      string
}

// CustomHostnameInput is used to create a Cloudflare SaaS custom hostname.
type CustomHostnameInput struct {
	Hostname string
	ZoneID   string // SaaS zone ID on Cloudflare
}

// CacheRuleInput describes a cache rule.
type CacheRuleInput struct {
	PathPattern string // URL path pattern, e.g. "/static/*"
	TTL         int    // seconds; 0 = no-cache
	Priority    int
}

// CacheRule is a stored cache configuration rule.
type CacheRule struct {
	ID          string
	PathPattern string
	TTL         int
	Priority    int
}

// CDNProvider is the plugin interface for CDN operations.
// Providers that don't support an operation return UnsupportedCapabilityError.
type CDNProvider interface {
	ProviderType() CDNProviderType
	ValidateCredentials(ctx context.Context) error

	// Domain lifecycle
	AddDomain(ctx context.Context, input CDNDomainInput) (*CDNZone, error)
	GetDomain(ctx context.Context, domainName string) (*CDNZone, error)
	DeleteDomain(ctx context.Context, domainName string) error
	ListDomains(ctx context.Context) ([]CDNZone, error)

	// Custom Hostname / SaaS (Cloudflare-specific; others return UnsupportedCapabilityError)
	CreateCustomHostname(ctx context.Context, zoneID string, input CustomHostnameInput) (*CustomHostname, error)
	GetCustomHostname(ctx context.Context, zoneID, hostnameID string) (*CustomHostname, error)
	DeleteCustomHostname(ctx context.Context, zoneID, hostnameID string) error
	ListCustomHostnames(ctx context.Context, zoneID string) ([]CustomHostname, error)

	// Cache
	PurgeURLs(ctx context.Context, domainName string, urls []string) error
	PurgeAll(ctx context.Context, domainName string) error

	// Cache rules
	SetCacheRule(ctx context.Context, domainName string, rule CacheRuleInput) (*CacheRule, error)
	ListCacheRules(ctx context.Context, domainName string) ([]CacheRule, error)
	DeleteCacheRule(ctx context.Context, domainName, ruleID string) error
}

// CDNProviderRegistry registers and instantiates CDN provider adapters.
type CDNProviderRegistry interface {
	RegisterCDN(pt CDNProviderType, factory func(credentials []byte) (CDNProvider, error))
	GetCDN(pt CDNProviderType) (CDNProvider, error)
	GetCDNWithCredentials(pt CDNProviderType, credentials []byte) (CDNProvider, error)
}
