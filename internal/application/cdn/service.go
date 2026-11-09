package cdn

import (
	"context"
	"errors"
	"fmt"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/crypto"
	infraprovider "github.com/wuekevin/smartcdn/internal/infra/provider"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

var (
	ErrAccountNotFound  = errors.New("cdn account not found")
	ErrDomainNotFound   = errors.New("cdn domain not found")
	ErrHostnameNotFound = errors.New("cdn hostname not found")
)

// Service handles CDN account, domain, and custom hostname management.
type Service struct {
	accounts  port.CDNAccountRepository
	domains   port.CDNDomainRepository
	hostnames port.CDNCustomHostnameRepository
	registry  *infraprovider.CDNRegistry
}

func NewService(
	accounts port.CDNAccountRepository,
	domains port.CDNDomainRepository,
	hostnames port.CDNCustomHostnameRepository,
	registry *infraprovider.CDNRegistry,
) *Service {
	return &Service{
		accounts:  accounts,
		domains:   domains,
		hostnames: hostnames,
		registry:  registry,
	}
}

// --- Account management ---

type CreateAccountInput struct {
	Name         string
	ProviderType string
	Credentials  []byte // raw credentials to encrypt
}

type AccountView struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	IsActive     bool   `json:"is_active"`
}

func (s *Service) CreateAccount(ctx context.Context, input CreateAccountInput) (*AccountView, error) {
	enc, err := crypto.Encrypt(input.Credentials)
	if err != nil {
		return nil, fmt.Errorf("cdn: encrypt credentials: %w", err)
	}
	acc := &entity.CDNAccount{
		Name:           input.Name,
		ProviderType:   input.ProviderType,
		CredentialsEnc: enc,
		IsActive:       true,
	}
	if err := s.accounts.Create(ctx, acc); err != nil {
		return nil, fmt.Errorf("cdn: create account: %w", err)
	}
	return toAccountView(acc), nil
}

func (s *Service) GetAccount(ctx context.Context, id uint64) (*AccountView, error) {
	acc, err := s.accounts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return toAccountView(acc), nil
}

func (s *Service) ListAccounts(ctx context.Context, afterID uint64, limit int) ([]AccountView, error) {
	accs, err := s.accounts.List(ctx, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]AccountView, len(accs))
	for i, a := range accs {
		views[i] = *toAccountView(&a)
	}
	return views, nil
}

func (s *Service) ValidateAccount(ctx context.Context, accountID uint64) error {
	provider, err := s.loadProvider(ctx, accountID)
	if err != nil {
		return err
	}
	return provider.ValidateCredentials(ctx)
}

// --- Domain management ---

type AddDomainInput struct {
	AccountID  uint64
	DomainName string
	OriginAddr string
	OriginPort int
	Protocol   string
}

type DomainView struct {
	ID             uint64 `json:"id"`
	AccountID      uint64 `json:"account_id"`
	DomainName     string `json:"domain_name"`
	ProviderZoneID string `json:"provider_zone_id,omitempty"`
	CNAME          string `json:"cname,omitempty"`
	Status         string `json:"status"`
	OriginAddr     string `json:"origin_addr"`
	OriginPort     int    `json:"origin_port"`
	Protocol       string `json:"protocol"`
	IsActive       bool   `json:"is_active"`
}

func (s *Service) AddDomain(ctx context.Context, input AddDomainInput) (*DomainView, error) {
	provider, err := s.loadProvider(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}
	zone, err := provider.AddDomain(ctx, port.CDNDomainInput{
		Domain:     input.DomainName,
		OriginAddr: input.OriginAddr,
		OriginPort: input.OriginPort,
		Protocol:   input.Protocol,
	})
	if err != nil {
		return nil, fmt.Errorf("cdn: add domain %q via provider: %w", input.DomainName, err)
	}
	d := &entity.CDNDomain{
		AccountID:      input.AccountID,
		DomainName:     input.DomainName,
		ProviderZoneID: zone.ID,
		CNAME:          zone.CNAME,
		Status:         zone.Status,
		OriginAddr:     input.OriginAddr,
		OriginPort:     input.OriginPort,
		Protocol:       input.Protocol,
		IsActive:       true,
	}
	if err := s.domains.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("cdn: persist domain: %w", err)
	}
	return toDomainView(d), nil
}

func (s *Service) GetDomain(ctx context.Context, id uint64) (*DomainView, error) {
	d, err := s.domains.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrDomainNotFound
		}
		return nil, err
	}
	return toDomainView(d), nil
}

func (s *Service) ListDomains(ctx context.Context, afterID uint64, limit int) ([]DomainView, error) {
	ds, err := s.domains.List(ctx, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]DomainView, len(ds))
	for i, d := range ds {
		views[i] = *toDomainView(&d)
	}
	return views, nil
}

func (s *Service) DeleteDomain(ctx context.Context, id uint64) error {
	d, err := s.domains.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrDomainNotFound
		}
		return err
	}
	provider, err := s.loadProvider(ctx, d.AccountID)
	if err != nil {
		return err
	}
	if err := provider.DeleteDomain(ctx, d.DomainName); err != nil {
		return fmt.Errorf("cdn: delete domain from provider: %w", err)
	}
	return s.domains.Delete(ctx, id)
}

func (s *Service) SyncDomainStatus(ctx context.Context, id uint64) (*DomainView, error) {
	d, err := s.domains.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrDomainNotFound
		}
		return nil, err
	}
	provider, err := s.loadProvider(ctx, d.AccountID)
	if err != nil {
		return nil, err
	}
	zone, err := provider.GetDomain(ctx, d.DomainName)
	if err != nil {
		return nil, fmt.Errorf("cdn: sync domain status: %w", err)
	}
	d.Status = zone.Status
	if zone.CNAME != "" {
		d.CNAME = zone.CNAME
	}
	if err := s.domains.Update(ctx, d); err != nil {
		return nil, err
	}
	return toDomainView(d), nil
}

// --- Cache management ---

func (s *Service) PurgeURLs(ctx context.Context, domainID uint64, urls []string) error {
	d, provider, err := s.loadDomainAndProvider(ctx, domainID)
	if err != nil {
		return err
	}
	return provider.PurgeURLs(ctx, d.DomainName, urls)
}

func (s *Service) PurgeAll(ctx context.Context, domainID uint64) error {
	d, provider, err := s.loadDomainAndProvider(ctx, domainID)
	if err != nil {
		return err
	}
	return provider.PurgeAll(ctx, d.DomainName)
}

// --- Custom Hostname / SaaS ---

type CreateHostnameInput struct {
	CDNDomainID uint64
	Hostname    string
}

type HostnameView struct {
	ID           uint64 `json:"id"`
	CDNDomainID  uint64 `json:"cdn_domain_id"`
	Hostname     string `json:"hostname"`
	ProviderID   string `json:"provider_id,omitempty"`
	Status       string `json:"status"`
	DCVTXTName   string `json:"dcv_txt_name,omitempty"`
	DCVTXTValue  string `json:"dcv_txt_value,omitempty"`
	DCVHTTPPath  string `json:"dcv_http_path,omitempty"`
	DCVHTTPValue string `json:"dcv_http_value,omitempty"`
	SSLStatus    string `json:"ssl_status,omitempty"`
}

func (s *Service) CreateCustomHostname(ctx context.Context, input CreateHostnameInput) (*HostnameView, error) {
	d, err := s.domains.FindByID(ctx, input.CDNDomainID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrDomainNotFound
		}
		return nil, err
	}
	provider, err := s.loadProvider(ctx, d.AccountID)
	if err != nil {
		return nil, err
	}
	ch, err := provider.CreateCustomHostname(ctx, d.ProviderZoneID, port.CustomHostnameInput{
		Hostname: input.Hostname,
		ZoneID:   d.ProviderZoneID,
	})
	if err != nil {
		return nil, fmt.Errorf("cdn: create custom hostname %q: %w", input.Hostname, err)
	}
	h := &entity.CDNCustomHostname{
		CDNDomainID:  input.CDNDomainID,
		Hostname:     input.Hostname,
		ProviderID:   ch.ID,
		Status:       ch.Status,
		DCVTXTName:   ch.DCVTXTName,
		DCVTXTValue:  ch.DCVTXTValue,
		DCVHTTPPath:  ch.HTTPTokenPath,
		DCVHTTPValue: ch.HTTPTokenValue,
		SSLStatus:    ch.SSLStatus,
	}
	if err := s.hostnames.Create(ctx, h); err != nil {
		return nil, fmt.Errorf("cdn: persist hostname: %w", err)
	}
	return toHostnameView(h), nil
}

func (s *Service) GetCustomHostname(ctx context.Context, id uint64) (*HostnameView, error) {
	h, err := s.hostnames.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrHostnameNotFound
		}
		return nil, err
	}
	return toHostnameView(h), nil
}

func (s *Service) ListCustomHostnames(ctx context.Context, cdnDomainID uint64, afterID uint64, limit int) ([]HostnameView, error) {
	hs, err := s.hostnames.ListByCDNDomain(ctx, cdnDomainID, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]HostnameView, len(hs))
	for i, h := range hs {
		views[i] = *toHostnameView(&h)
	}
	return views, nil
}

func (s *Service) SyncCustomHostname(ctx context.Context, id uint64) (*HostnameView, error) {
	h, err := s.hostnames.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrHostnameNotFound
		}
		return nil, err
	}
	d, err := s.domains.FindByID(ctx, h.CDNDomainID)
	if err != nil {
		return nil, err
	}
	provider, err := s.loadProvider(ctx, d.AccountID)
	if err != nil {
		return nil, err
	}
	ch, err := provider.GetCustomHostname(ctx, d.ProviderZoneID, h.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("cdn: sync hostname status: %w", err)
	}
	h.Status = ch.Status
	h.SSLStatus = ch.SSLStatus
	if ch.DCVTXTName != "" {
		h.DCVTXTName = ch.DCVTXTName
		h.DCVTXTValue = ch.DCVTXTValue
	}
	if err := s.hostnames.Update(ctx, h); err != nil {
		return nil, err
	}
	return toHostnameView(h), nil
}

func (s *Service) DeleteCustomHostname(ctx context.Context, id uint64) error {
	h, err := s.hostnames.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrHostnameNotFound
		}
		return err
	}
	d, err := s.domains.FindByID(ctx, h.CDNDomainID)
	if err != nil {
		return err
	}
	provider, err := s.loadProvider(ctx, d.AccountID)
	if err != nil {
		return err
	}
	if err := provider.DeleteCustomHostname(ctx, d.ProviderZoneID, h.ProviderID); err != nil {
		return fmt.Errorf("cdn: delete hostname from provider: %w", err)
	}
	return s.hostnames.Delete(ctx, id)
}

// --- helpers ---

func (s *Service) loadProvider(ctx context.Context, accountID uint64) (port.CDNProvider, error) {
	acc, err := s.accounts.FindByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	credBytes, err := crypto.Decrypt(acc.CredentialsEnc)
	if err != nil {
		return nil, fmt.Errorf("cdn: decrypt credentials: %w", err)
	}
	return s.registry.GetCDNWithCredentials(port.CDNProviderType(acc.ProviderType), credBytes)
}

func (s *Service) loadDomainAndProvider(ctx context.Context, domainID uint64) (*entity.CDNDomain, port.CDNProvider, error) {
	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil, ErrDomainNotFound
		}
		return nil, nil, err
	}
	provider, err := s.loadProvider(ctx, d.AccountID)
	if err != nil {
		return nil, nil, err
	}
	return d, provider, nil
}

func toAccountView(a *entity.CDNAccount) *AccountView {
	return &AccountView{
		ID:           a.ID,
		Name:         a.Name,
		ProviderType: a.ProviderType,
		IsActive:     a.IsActive,
	}
}

func toDomainView(d *entity.CDNDomain) *DomainView {
	return &DomainView{
		ID:             d.ID,
		AccountID:      d.AccountID,
		DomainName:     d.DomainName,
		ProviderZoneID: d.ProviderZoneID,
		CNAME:          d.CNAME,
		Status:         d.Status,
		OriginAddr:     d.OriginAddr,
		OriginPort:     d.OriginPort,
		Protocol:       d.Protocol,
		IsActive:       d.IsActive,
	}
}

func toHostnameView(h *entity.CDNCustomHostname) *HostnameView {
	return &HostnameView{
		ID:           h.ID,
		CDNDomainID:  h.CDNDomainID,
		Hostname:     h.Hostname,
		ProviderID:   h.ProviderID,
		Status:       h.Status,
		DCVTXTName:   h.DCVTXTName,
		DCVTXTValue:  h.DCVTXTValue,
		DCVHTTPPath:  h.DCVHTTPPath,
		DCVHTTPValue: h.DCVHTTPValue,
		SSLStatus:    h.SSLStatus,
	}
}
