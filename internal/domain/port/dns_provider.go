package port

import (
	"context"
	"errors"
	"time"
)

// ErrProviderNotFound is returned by Registry when the requested provider ID is not registered.
var ErrProviderNotFound = errors.New("provider not found")

// UnsupportedCapabilityError is returned when an operation requires a plan the account does not have.
type UnsupportedCapabilityError struct {
	Operation string
	Reason    string
}

func (e *UnsupportedCapabilityError) Error() string {
	return "unsupported capability: " + e.Operation + ": " + e.Reason
}

// ProviderType names the DNS provider implementation.
type ProviderType string

const (
	ProviderAliyun    ProviderType = "aliyun"
	ProviderDNSPod    ProviderType = "dnspod"
	ProviderCloudflare ProviderType = "cloudflare"
)

// DNSRecord is a provider-agnostic DNS record value object.
type DNSRecord struct {
	ID       string        // provider-side record ID
	Host     string        // subdomain or "@" for apex
	Type     string        // A, AAAA, CNAME, MX, TXT, ...
	Value    string        // record value
	TTL      int           // seconds
	Line     string        // ISP routing line; empty = default
	Priority int           // MX / SRV priority
	UpdatedAt time.Time
}

// DNSRecordInput is used to create or update a record.
type DNSRecordInput struct {
	Host     string
	Type     string
	Value    string
	TTL      int
	Line     string
	Priority int
}

// DNSZone represents a domain zone managed by the provider.
type DNSZone struct {
	ID     string
	Domain string // apex domain, e.g. "example.com"
}

// Capabilities describes what an account's subscription plan supports.
type Capabilities struct {
	// ISPRouting indicates whether the account supports split-horizon (三网) routing.
	ISPRouting bool
	// Lines contains supported routing line codes, e.g. ["default", "telecom", "unicom", "mobile"].
	Lines []string
	// MinTTL is the minimum allowed TTL in seconds.
	MinTTL int
	// MaxRecordsPerSet is the maximum number of records per host+type combination.
	MaxRecordsPerSet int
	// RawPlanName is the provider's plan name string, for display only.
	RawPlanName string
}

// BatchRecordOp is a single operation inside a batch.
type BatchRecordOp struct {
	Op    string // "create" | "update" | "delete"
	ID    string // provider record ID; required for update/delete
	Input *DNSRecordInput
}

// BatchResult is the per-operation result of a batch call.
type BatchResult struct {
	RecordID string
	Err      error
}

// DNSProvider is the Port interface that every DNS provider adapter must implement (AD-2).
// Application layer code may only depend on this interface, never on concrete adapter packages.
type DNSProvider interface {
	// ProviderType returns the well-known ID of this provider, e.g. "aliyun".
	ProviderType() ProviderType

	// ValidateCredentials verifies that the supplied credentials are accepted by the provider.
	ValidateCredentials(ctx context.Context) error

	// DetectCapabilities queries the provider for plan-level feature support.
	DetectCapabilities(ctx context.Context) (*Capabilities, error)

	// ListZones returns all zones (domains) accessible with the current credentials.
	ListZones(ctx context.Context) ([]DNSZone, error)

	// ListRecords returns all records for the given zone (by provider zone ID or apex domain).
	ListRecords(ctx context.Context, zoneID string) ([]DNSRecord, error)

	// CreateRecord creates a new DNS record and returns the provider-assigned ID.
	CreateRecord(ctx context.Context, zoneID string, input DNSRecordInput) (string, error)

	// UpdateRecord updates the record identified by recordID.
	UpdateRecord(ctx context.Context, zoneID, recordID string, input DNSRecordInput) error

	// DeleteRecord deletes the record identified by recordID.
	DeleteRecord(ctx context.Context, zoneID, recordID string) error

	// CreateBatch atomically (best-effort) creates multiple records.
	// Partial success is returned as individual errors inside BatchResult.
	CreateBatch(ctx context.Context, zoneID string, inputs []DNSRecordInput) ([]BatchResult, error)

	// GetRecord fetches a single record by its provider-side record ID.
	GetRecord(ctx context.Context, zoneID, recordID string) (*DNSRecord, error)

	// FindRecords returns records matching host, type, and optional line filter.
	FindRecords(ctx context.Context, zoneID, host, recordType, line string) ([]DNSRecord, error)

	// PauseRecord pauses (disables) a record without deleting it, if supported.
	PauseRecord(ctx context.Context, zoneID, recordID string) error

	// ResumeRecord re-enables a paused record.
	ResumeRecord(ctx context.Context, zoneID, recordID string) error

	// GetNameservers returns the authoritative nameservers for the given zone.
	GetNameservers(ctx context.Context, zoneID string) ([]string, error)
}

// DNSProviderRegistry allows the application layer to obtain a provider by ID without
// importing concrete adapter packages directly (AD-1).
type DNSProviderRegistry interface {
	GetDNS(providerType ProviderType) (DNSProvider, error)
	Register(providerType ProviderType, factory func(credentials []byte) (DNSProvider, error))
}
