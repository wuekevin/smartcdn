package migration

import (
	"context"
	"fmt"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/crypto"
	infraprovider "github.com/wuekevin/smartcdn/internal/infra/provider"
)

// MultiDNSService manages dual-authority DNS configuration.
// When SyncEnabled=true, every record write is mirrored to the secondary provider.
type MultiDNSService struct {
	configs  port.MultiDNSConfigRepository
	domains  port.DomainRepository
	accounts port.DNSAccountRepository
	registry *infraprovider.Registry
}

func NewMultiDNSService(
	configs port.MultiDNSConfigRepository,
	domains port.DomainRepository,
	accounts port.DNSAccountRepository,
	registry *infraprovider.Registry,
) *MultiDNSService {
	return &MultiDNSService{
		configs:  configs,
		domains:  domains,
		accounts: accounts,
		registry: registry,
	}
}

type MultiDNSConfigView struct {
	ID                 uint64 `json:"id"`
	DomainID           uint64 `json:"domain_id"`
	PrimaryAccountID   uint64 `json:"primary_account_id"`
	SecondaryAccountID uint64 `json:"secondary_account_id"`
	SyncEnabled        bool   `json:"sync_enabled"`
}

// Enable activates dual-authority for a domain.
func (s *MultiDNSService) Enable(ctx context.Context, domainID, primaryID, secondaryID uint64) (*MultiDNSConfigView, error) {
	if primaryID == secondaryID {
		return nil, fmt.Errorf("primary and secondary accounts must differ")
	}

	// Check if already configured; update if so
	existing, err := s.configs.FindByDomainID(ctx, domainID)
	if err == nil {
		existing.PrimaryAccountID = primaryID
		existing.SecondaryAccountID = secondaryID
		existing.SyncEnabled = true
		if err := s.configs.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("update multidns config: %w", err)
		}
		return toConfigView(existing), nil
	}

	cfg := &entity.MultiDNSConfig{
		DomainID:           domainID,
		PrimaryAccountID:   primaryID,
		SecondaryAccountID: secondaryID,
		SyncEnabled:        true,
	}
	if err := s.configs.Create(ctx, cfg); err != nil {
		return nil, fmt.Errorf("create multidns config: %w", err)
	}
	return toConfigView(cfg), nil
}

// Disable removes dual-authority for a domain.
func (s *MultiDNSService) Disable(ctx context.Context, domainID uint64) error {
	return s.configs.Delete(ctx, domainID)
}

// GetConfig returns the current dual-authority config for a domain.
func (s *MultiDNSService) GetConfig(ctx context.Context, domainID uint64) (*MultiDNSConfigView, error) {
	cfg, err := s.configs.FindByDomainID(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("multidns config not found for domain %d", domainID)
	}
	return toConfigView(cfg), nil
}

// SyncRecord mirrors a record creation to the secondary provider (best-effort).
// Returns nil on success; logs the error without failing the primary write.
func (s *MultiDNSService) SyncRecord(ctx context.Context, domainID uint64, input port.DNSRecordInput) error {
	cfg, err := s.configs.FindByDomainID(ctx, domainID)
	if err != nil || !cfg.SyncEnabled {
		return nil // no config or sync disabled — silent no-op
	}

	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		return nil
	}

	secProv, zoneID, err := s.loadProviderAndZone(ctx, cfg.SecondaryAccountID, d.Name)
	if err != nil {
		return fmt.Errorf("multidns sync: load secondary provider: %w", err)
	}
	_, err = secProv.CreateRecord(ctx, zoneID, input)
	return err
}

// SyncDelete mirrors a record deletion to the secondary provider (best-effort).
func (s *MultiDNSService) SyncDelete(ctx context.Context, domainID uint64, host, recordType string) error {
	cfg, err := s.configs.FindByDomainID(ctx, domainID)
	if err != nil || !cfg.SyncEnabled {
		return nil
	}

	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		return nil
	}

	secProv, zoneID, err := s.loadProviderAndZone(ctx, cfg.SecondaryAccountID, d.Name)
	if err != nil {
		return fmt.Errorf("multidns sync delete: %w", err)
	}

	recs, err := secProv.FindRecords(ctx, zoneID, host, recordType, "")
	if err != nil {
		return fmt.Errorf("multidns find records: %w", err)
	}
	for _, r := range recs {
		_ = secProv.DeleteRecord(ctx, zoneID, r.ID) // best-effort
	}
	return nil
}

func (s *MultiDNSService) loadProviderAndZone(ctx context.Context, accountID uint64, domainName string) (port.DNSProvider, string, error) {
	acct, err := s.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, "", fmt.Errorf("load account: %w", err)
	}
	credBytes, err := crypto.Decrypt(acct.CredentialsEnc)
	if err != nil {
		return nil, "", fmt.Errorf("decrypt credentials: %w", err)
	}
	prov, err := s.registry.GetDNSWithCredentials(port.ProviderType(acct.ProviderType), credBytes)
	if err != nil {
		return nil, "", err
	}
	zones, err := prov.ListZones(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("list zones: %w", err)
	}
	for _, z := range zones {
		if z.Domain == domainName {
			return prov, z.ID, nil
		}
	}
	return nil, "", fmt.Errorf("zone not found for %q", domainName)
}

func toConfigView(cfg *entity.MultiDNSConfig) *MultiDNSConfigView {
	return &MultiDNSConfigView{
		ID:                 cfg.ID,
		DomainID:           cfg.DomainID,
		PrimaryAccountID:   cfg.PrimaryAccountID,
		SecondaryAccountID: cfg.SecondaryAccountID,
		SyncEnabled:        cfg.SyncEnabled,
	}
}
