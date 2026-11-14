package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/crypto"
	"github.com/wuekevin/smartcdn/internal/infra/httpclient"
	infraprovider "github.com/wuekevin/smartcdn/internal/infra/provider"
)

var (
	ErrDomainNotFound  = errors.New("domain not found")
	ErrDomainExists    = errors.New("domain already exists")
	ErrRecordNotFound  = errors.New("record not found")
	ErrSnapshotNotFound = errors.New("snapshot not found")
)

// Service handles all domain and DNS record business logic.
type Service struct {
	domains   port.DomainRepository
	records   port.DNSRecordRepository
	snapshots port.DNSSnapshotRepository
	accounts  port.DNSAccountRepository
	registry  *infraprovider.Registry
}

func NewService(
	domains port.DomainRepository,
	records port.DNSRecordRepository,
	snapshots port.DNSSnapshotRepository,
	accounts port.DNSAccountRepository,
	registry *infraprovider.Registry,
) *Service {
	return &Service{
		domains:   domains,
		records:   records,
		snapshots: snapshots,
		accounts:  accounts,
		registry:  registry,
	}
}

// --- Domain management ---

type CreateDomainInput struct {
	Name         string `json:"name"`
	AccountID    uint64 `json:"account_id"`
	ICPFiled     bool   `json:"icp_filed"`
	MaxPublishIPs int   `json:"max_publish_ips"`
}

type DomainView struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	AccountID    uint64    `json:"account_id"`
	ICPFiled     bool      `json:"icp_filed"`
	MaxPublishIPs int      `json:"max_publish_ips"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Service) CreateDomain(ctx context.Context, input CreateDomainInput) (*DomainView, error) {
	if input.MaxPublishIPs <= 0 {
		input.MaxPublishIPs = 1
	}

	_, err := s.domains.FindByName(ctx, input.Name)
	if err == nil {
		return nil, ErrDomainExists
	}

	d := &entity.Domain{
		Name:         input.Name,
		AccountID:    input.AccountID,
		ICPFiled:     input.ICPFiled,
		MaxPublishIPs: input.MaxPublishIPs,
		IsActive:     true,
	}
	if err := s.domains.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("create domain: %w", err)
	}
	return domainToView(d), nil
}

func (s *Service) GetDomain(ctx context.Context, id uint64) (*DomainView, error) {
	d, err := s.domains.FindByID(ctx, id)
	if err != nil {
		return nil, ErrDomainNotFound
	}
	return domainToView(d), nil
}

func (s *Service) ListDomains(ctx context.Context, afterID uint64, limit int) ([]DomainView, error) {
	ds, err := s.domains.List(ctx, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	views := make([]DomainView, len(ds))
	for i, d := range ds {
		views[i] = *domainToView(&d)
	}
	return views, nil
}

type UpdateDomainInput struct {
	ICPFiled      bool   `json:"icp_filed"`
	IsActive      bool   `json:"is_active"`
	MaxPublishIPs int    `json:"max_publish_ips"`
}

func (s *Service) UpdateDomain(ctx context.Context, id uint64, input UpdateDomainInput) (*DomainView, error) {
	d, err := s.domains.FindByID(ctx, id)
	if err != nil {
		return nil, ErrDomainNotFound
	}
	d.ICPFiled = input.ICPFiled
	d.IsActive = input.IsActive
	if input.MaxPublishIPs > 0 {
		d.MaxPublishIPs = input.MaxPublishIPs
	}
	if err := s.domains.Update(ctx, d); err != nil {
		return nil, fmt.Errorf("update domain: %w", err)
	}
	return domainToView(d), nil
}

func (s *Service) DeleteDomain(ctx context.Context, id uint64) error {
	if _, err := s.domains.FindByID(ctx, id); err != nil {
		return ErrDomainNotFound
	}
	if err := s.records.DeleteByDomain(ctx, id); err != nil {
		return fmt.Errorf("delete domain records: %w", err)
	}
	return s.domains.Delete(ctx, id)
}

// --- DNS Record management ---

type CreateRecordInput struct {
	Host     string `json:"host"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	Line     string `json:"line"`
	Priority int    `json:"priority"`
}

type UpdateRecordInput struct {
	Host     string `json:"host"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	Line     string `json:"line"`
	Priority int    `json:"priority"`
}

type RecordView struct {
	ID         uint64    `json:"id"`
	DomainID   uint64    `json:"domain_id"`
	ProviderID string    `json:"provider_id,omitempty"`
	Host       string    `json:"host"`
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	TTL        int       `json:"ttl"`
	Line       string    `json:"line,omitempty"`
	Priority   int       `json:"priority,omitempty"`
	IsExternal bool      `json:"is_external"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateRecord creates a DNS record both locally and on the provider.
// SSRF protection: rejects RFC1918/loopback values for A records.
func (s *Service) CreateRecord(ctx context.Context, domainID uint64, input CreateRecordInput) (*RecordView, error) {
	// SSRF guard: reject private/loopback IPs in A/AAAA records
	if input.Type == "A" || input.Type == "AAAA" {
		if err := httpclient.CheckIP(input.Value); err != nil {
			return nil, fmt.Errorf("SSRF protection: %w", err)
		}
	}

	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		return nil, ErrDomainNotFound
	}

	providerRecordID, err := s.createOnProvider(ctx, d, input)
	if err != nil {
		return nil, fmt.Errorf("provider create record: %w", err)
	}

	rec := &entity.DNSRecord{
		DomainID:   domainID,
		ProviderID: providerRecordID,
		Host:       input.Host,
		Type:       input.Type,
		Value:      input.Value,
		TTL:        input.TTL,
		Line:       input.Line,
		Priority:   input.Priority,
	}
	if rec.TTL <= 0 {
		rec.TTL = 600
	}
	if err := s.records.Create(ctx, rec); err != nil {
		return nil, fmt.Errorf("store record: %w", err)
	}
	return recordToView(rec), nil
}

func (s *Service) UpdateRecord(ctx context.Context, domainID, recordID uint64, input UpdateRecordInput) (*RecordView, error) {
	if input.Type == "A" || input.Type == "AAAA" {
		if err := httpclient.CheckIP(input.Value); err != nil {
			return nil, fmt.Errorf("SSRF protection: %w", err)
		}
	}

	rec, err := s.records.FindByID(ctx, recordID)
	if err != nil || rec.DomainID != domainID {
		return nil, ErrRecordNotFound
	}

	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		return nil, ErrDomainNotFound
	}

	if err := s.updateOnProvider(ctx, d, rec.ProviderID, input); err != nil {
		return nil, fmt.Errorf("provider update record: %w", err)
	}

	rec.Host = input.Host
	rec.Type = input.Type
	rec.Value = input.Value
	rec.TTL = input.TTL
	rec.Line = input.Line
	rec.Priority = input.Priority
	if err := s.records.Update(ctx, rec); err != nil {
		return nil, fmt.Errorf("update record: %w", err)
	}
	return recordToView(rec), nil
}

func (s *Service) DeleteRecord(ctx context.Context, domainID, recordID uint64) error {
	rec, err := s.records.FindByID(ctx, recordID)
	if err != nil || rec.DomainID != domainID {
		return ErrRecordNotFound
	}

	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		return ErrDomainNotFound
	}

	if err := s.deleteOnProvider(ctx, d, rec.ProviderID); err != nil {
		return fmt.Errorf("provider delete record: %w", err)
	}
	return s.records.Delete(ctx, recordID)
}

func (s *Service) ListRecords(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]RecordView, error) {
	recs, err := s.records.ListByDomain(ctx, domainID, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}
	views := make([]RecordView, len(recs))
	for i, r := range recs {
		views[i] = *recordToView(&r)
	}
	return views, nil
}

// --- Snapshot management ---

type SnapshotView struct {
	ID           uint64    `json:"id"`
	DomainID     uint64    `json:"domain_id"`
	TriggerEvent string    `json:"trigger_event"`
	RecordCount  int       `json:"record_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateSnapshot saves a snapshot of all current records for the domain.
func (s *Service) CreateSnapshot(ctx context.Context, domainID uint64, trigger string) (*SnapshotView, error) {
	recs, err := s.records.ListByDomain(ctx, domainID, 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("list records for snapshot: %w", err)
	}
	data, err := json.Marshal(recs)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}
	snap := &entity.DNSSnapshot{
		DomainID:     domainID,
		SnapshotJSON: string(data),
		TriggerEvent: trigger,
	}
	if err := s.snapshots.Create(ctx, snap); err != nil {
		return nil, fmt.Errorf("store snapshot: %w", err)
	}
	return &SnapshotView{
		ID:           snap.ID,
		DomainID:     snap.DomainID,
		TriggerEvent: snap.TriggerEvent,
		RecordCount:  len(recs),
		CreatedAt:    snap.CreatedAt,
	}, nil
}

func (s *Service) ListSnapshots(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]SnapshotView, error) {
	snaps, err := s.snapshots.ListByDomain(ctx, domainID, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	views := make([]SnapshotView, len(snaps))
	for i, snap := range snaps {
		var recs []entity.DNSRecord
		_ = json.Unmarshal([]byte(snap.SnapshotJSON), &recs)
		views[i] = SnapshotView{
			ID:           snap.ID,
			DomainID:     snap.DomainID,
			TriggerEvent: snap.TriggerEvent,
			RecordCount:  len(recs),
			CreatedAt:    snap.CreatedAt,
		}
	}
	return views, nil
}

// --- Provider helpers ---

func (s *Service) loadProvider(ctx context.Context, d *entity.Domain) (port.DNSProvider, error) {
	acct, err := s.accounts.FindByID(ctx, d.AccountID)
	if err != nil {
		return nil, fmt.Errorf("load account: %w", err)
	}
	credBytes, err := crypto.Decrypt(acct.CredentialsEnc)
	if err != nil {
		return nil, fmt.Errorf("decrypt credentials: %w", err)
	}
	return s.registry.GetDNSWithCredentials(port.ProviderType(acct.ProviderType), credBytes)
}

func (s *Service) getZoneID(ctx context.Context, p port.DNSProvider, domainName string) (string, error) {
	zones, err := p.ListZones(ctx)
	if err != nil {
		return "", fmt.Errorf("list zones: %w", err)
	}
	for _, z := range zones {
		if z.Domain == domainName {
			return z.ID, nil
		}
	}
	return "", fmt.Errorf("zone not found for domain %q", domainName)
}

func (s *Service) createOnProvider(ctx context.Context, d *entity.Domain, input CreateRecordInput) (string, error) {
	p, err := s.loadProvider(ctx, d)
	if err != nil {
		return "", err
	}
	zoneID, err := s.getZoneID(ctx, p, d.Name)
	if err != nil {
		return "", err
	}
	return p.CreateRecord(ctx, zoneID, port.DNSRecordInput{
		Host:     input.Host,
		Type:     input.Type,
		Value:    input.Value,
		TTL:      input.TTL,
		Line:     input.Line,
		Priority: input.Priority,
	})
}

func (s *Service) updateOnProvider(ctx context.Context, d *entity.Domain, providerRecordID string, input UpdateRecordInput) error {
	if providerRecordID == "" {
		return nil // external record; skip
	}
	p, err := s.loadProvider(ctx, d)
	if err != nil {
		return err
	}
	zoneID, err := s.getZoneID(ctx, p, d.Name)
	if err != nil {
		return err
	}
	return p.UpdateRecord(ctx, zoneID, providerRecordID, port.DNSRecordInput{
		Host:     input.Host,
		Type:     input.Type,
		Value:    input.Value,
		TTL:      input.TTL,
		Line:     input.Line,
		Priority: input.Priority,
	})
}

func (s *Service) deleteOnProvider(ctx context.Context, d *entity.Domain, providerRecordID string) error {
	if providerRecordID == "" {
		return nil
	}
	p, err := s.loadProvider(ctx, d)
	if err != nil {
		return err
	}
	zoneID, err := s.getZoneID(ctx, p, d.Name)
	if err != nil {
		return err
	}
	return p.DeleteRecord(ctx, zoneID, providerRecordID)
}

// PublishIPForISP finds the A record for the given ISP line on domainID and
// updates its value to ip. It satisfies the scheduler.DNSPublisher interface.
func (s *Service) PublishIPForISP(ctx context.Context, domainID uint64, isp, ip string) error {
	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		return err
	}
	records, err := s.records.ListByDomain(ctx, domainID, 0, 500)
	if err != nil {
		return err
	}
	for _, r := range records {
		if r.Type == "A" && r.Line == isp {
			return s.updateOnProvider(ctx, d, r.ProviderID, UpdateRecordInput{
				Host:  r.Host,
				Type:  r.Type,
				Value: ip,
				TTL:   r.TTL,
				Line:  r.Line,
			})
		}
	}
	// No existing record for this ISP line — create one.
	_, err = s.createOnProvider(ctx, d, CreateRecordInput{
		Host:  "@",
		Type:  "A",
		Value: ip,
		TTL:   60,
		Line:  isp,
	})
	return err
}

// --- View converters ---

func domainToView(d *entity.Domain) *DomainView {
	return &DomainView{
		ID:           d.ID,
		Name:         d.Name,
		AccountID:    d.AccountID,
		ICPFiled:     d.ICPFiled,
		MaxPublishIPs: d.MaxPublishIPs,
		IsActive:     d.IsActive,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func recordToView(r *entity.DNSRecord) *RecordView {
	return &RecordView{
		ID:         r.ID,
		DomainID:   r.DomainID,
		ProviderID: r.ProviderID,
		Host:       r.Host,
		Type:       r.Type,
		Value:      r.Value,
		TTL:        r.TTL,
		Line:       r.Line,
		Priority:   r.Priority,
		IsExternal: r.IsExternal,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
