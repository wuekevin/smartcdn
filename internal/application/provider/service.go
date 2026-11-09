package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/crypto"
	"github.com/wuekevin/smartcdn/internal/infra/provider"
)

const capabilitiesCacheTTL = time.Hour

// Service manages DNS provider accounts: create, credential rotation, capability detection.
type Service struct {
	accounts  port.DNSAccountRepository
	registry  *provider.Registry
	rdb       *redis.Client
}

func NewService(accounts port.DNSAccountRepository, registry *provider.Registry, rdb *redis.Client) *Service {
	return &Service{accounts: accounts, registry: registry, rdb: rdb}
}

// capCacheKey returns the Redis key for a cached capabilities JSON.
func capCacheKey(accountID uint64) string {
	return fmt.Sprintf("dns_cap:%d", accountID)
}

// CreateAccount encrypts the credentials and persists a new DNS account.
func (s *Service) CreateAccount(ctx context.Context, name string, pt port.ProviderType, creds map[string]string) (*entity.DNSAccount, error) {
	raw, err := json.Marshal(creds)
	if err != nil {
		return nil, fmt.Errorf("marshal creds: %w", err)
	}
	enc, err := crypto.Encrypt(raw)
	if err != nil {
		return nil, fmt.Errorf("encrypt creds: %w", err)
	}

	acc := &entity.DNSAccount{
		Name:           name,
		ProviderType:   string(pt),
		CredentialsEnc: enc,
		IsActive:       true,
	}
	if err := s.accounts.Create(ctx, acc); err != nil {
		return nil, err
	}
	return acc, nil
}

// GetAccount returns the account with masked credentials for display.
func (s *Service) GetAccount(ctx context.Context, id uint64) (*AccountView, error) {
	acc, err := s.accounts.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toView(acc), nil
}

// ListAccounts returns a cursor-paged list of accounts.
func (s *Service) ListAccounts(ctx context.Context, afterID uint64, limit int) ([]*AccountView, error) {
	accs, err := s.accounts.List(ctx, afterID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*AccountView, len(accs))
	for i := range accs {
		out[i] = toView(&accs[i])
	}
	return out, nil
}

// RotateCredentials validates new credentials before replacing the old ones (AD-8).
func (s *Service) RotateCredentials(ctx context.Context, id uint64, newCreds map[string]string) error {
	acc, err := s.accounts.FindByID(ctx, id)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(newCreds)
	if err != nil {
		return fmt.Errorf("marshal creds: %w", err)
	}

	p, err := s.registry.GetDNSWithCredentials(port.ProviderType(acc.ProviderType), raw)
	if err != nil {
		return fmt.Errorf("load provider: %w", err)
	}
	if err := p.ValidateCredentials(ctx); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	enc, err := crypto.Encrypt(raw)
	if err != nil {
		return fmt.Errorf("encrypt creds: %w", err)
	}
	return s.accounts.UpdateCredentials(ctx, id, enc)
}

// RefreshCapabilities clears the cache and re-detects capabilities from the provider.
func (s *Service) RefreshCapabilities(ctx context.Context, id uint64) (*port.Capabilities, error) {
	acc, err := s.accounts.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rawCreds, err := crypto.Decrypt(acc.CredentialsEnc)
	if err != nil {
		return nil, fmt.Errorf("decrypt creds: %w", err)
	}

	p, err := s.registry.GetDNSWithCredentials(port.ProviderType(acc.ProviderType), rawCreds)
	if err != nil {
		return nil, fmt.Errorf("load provider: %w", err)
	}

	caps, err := p.DetectCapabilities(ctx)
	if err != nil {
		return nil, fmt.Errorf("detect capabilities: %w", err)
	}

	capJSON, _ := json.Marshal(caps)
	_ = s.accounts.UpdateCapabilities(ctx, id, string(capJSON))
	_ = s.rdb.Set(ctx, capCacheKey(id), capJSON, capabilitiesCacheTTL).Err()
	return caps, nil
}

// GetCapabilities returns capabilities from Redis cache, falling back to DB, then live detection.
func (s *Service) GetCapabilities(ctx context.Context, id uint64) (*port.Capabilities, error) {
	cached, err := s.rdb.Get(ctx, capCacheKey(id)).Bytes()
	if err == nil {
		var caps port.Capabilities
		if json.Unmarshal(cached, &caps) == nil {
			return &caps, nil
		}
	}

	acc, err := s.accounts.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if acc.CapabilitiesJSON != "" {
		var caps port.Capabilities
		if json.Unmarshal([]byte(acc.CapabilitiesJSON), &caps) == nil {
			_ = s.rdb.Set(ctx, capCacheKey(id), []byte(acc.CapabilitiesJSON), capabilitiesCacheTTL).Err()
			return &caps, nil
		}
	}

	return s.RefreshCapabilities(ctx, id)
}

// DeleteAccount soft-deletes the given DNS account.
func (s *Service) DeleteAccount(ctx context.Context, id uint64) error {
	return s.accounts.Delete(ctx, id)
}

// GetProvider decrypts credentials and returns a live provider instance.
func (s *Service) GetProvider(ctx context.Context, accountID uint64) (port.DNSProvider, error) {
	acc, err := s.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	rawCreds, err := crypto.Decrypt(acc.CredentialsEnc)
	if err != nil {
		return nil, fmt.Errorf("decrypt creds: %w", err)
	}
	return s.registry.GetDNSWithCredentials(port.ProviderType(acc.ProviderType), rawCreds)
}

// AccountView is the read model returned to API callers (credentials masked).
type AccountView struct {
	ID           uint64 `json:"id,string"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	// Credentials are always masked; actual values are never returned over the API.
	CredentialsMasked map[string]string `json:"credentials"`
}

func toView(acc *entity.DNSAccount) *AccountView {
	var rawCreds map[string]string
	if dec, err := crypto.Decrypt(acc.CredentialsEnc); err == nil {
		_ = json.Unmarshal(dec, &rawCreds)
	}
	masked := maskCredentials(rawCreds)
	return &AccountView{
		ID:                acc.ID,
		Name:              acc.Name,
		ProviderType:      acc.ProviderType,
		IsActive:          acc.IsActive,
		CreatedAt:         acc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         acc.UpdatedAt.Format(time.RFC3339),
		CredentialsMasked: masked,
	}
}

func maskCredentials(creds map[string]string) map[string]string {
	out := make(map[string]string, len(creds))
	for k, v := range creds {
		if len(v) <= 4 {
			out[k] = "****"
		} else {
			out[k] = "****" + v[len(v)-4:]
		}
		// Keys that explicitly contain "secret" are fully masked.
		if strings.Contains(strings.ToLower(k), "secret") {
			out[k] = "****"
		}
	}
	return out
}
