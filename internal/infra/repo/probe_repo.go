package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// ProbeAgentRepo ──────────────────────────────────────────────────────────

type ProbeAgentRepo struct{ db *gorm.DB }

func NewProbeAgentRepo(db *gorm.DB) *ProbeAgentRepo { return &ProbeAgentRepo{db: db} }

func (r *ProbeAgentRepo) Create(ctx context.Context, a *entity.ProbeAgent) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *ProbeAgentRepo) FindByID(ctx context.Context, id uint64) (*entity.ProbeAgent, error) {
	var a entity.ProbeAgent
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ProbeAgentRepo.FindByID: %w", err)
	}
	return &a, nil
}

func (r *ProbeAgentRepo) FindByCertSN(ctx context.Context, certSN string) (*entity.ProbeAgent, error) {
	var a entity.ProbeAgent
	if err := r.db.WithContext(ctx).Where("client_cert_s_n = ? AND is_active = true", certSN).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ProbeAgentRepo.FindByCertSN: %w", err)
	}
	return &a, nil
}

func (r *ProbeAgentRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.ProbeAgent, error) {
	if limit <= 0 {
		limit = 20
	}
	var agents []entity.ProbeAgent
	q := r.db.WithContext(ctx)
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return agents, q.Order("id ASC").Limit(limit).Find(&agents).Error
}

func (r *ProbeAgentRepo) Update(ctx context.Context, a *entity.ProbeAgent) error {
	return r.db.WithContext(ctx).Save(a).Error
}

// ProbeResultRepo ─────────────────────────────────────────────────────────

type ProbeResultRepo struct{ db *gorm.DB }

func NewProbeResultRepo(db *gorm.DB) *ProbeResultRepo { return &ProbeResultRepo{db: db} }

func (r *ProbeResultRepo) Create(ctx context.Context, res *entity.ProbeResult) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *ProbeResultRepo) ListRecentByIP(ctx context.Context, ip string, since time.Time, limit int) ([]entity.ProbeResult, error) {
	if limit <= 0 {
		limit = 5
	}
	var results []entity.ProbeResult
	return results, r.db.WithContext(ctx).
		Where("target_ip = ? AND probe_at >= ?", ip, since).
		Order("probe_at DESC").Limit(limit).Find(&results).Error
}

func (r *ProbeResultRepo) ListRecentByIPAndISP(ctx context.Context, ip, isp string, since time.Time, limit int) ([]entity.ProbeResult, error) {
	if limit <= 0 {
		limit = 5
	}
	var results []entity.ProbeResult
	return results, r.db.WithContext(ctx).
		Where("target_ip = ? AND isp = ? AND probe_at >= ?", ip, isp, since).
		Order("probe_at DESC").Limit(limit).Find(&results).Error
}

// IPCandidateRepo ─────────────────────────────────────────────────────────

type IPCandidateRepo struct{ db *gorm.DB }

func NewIPCandidateRepo(db *gorm.DB) *IPCandidateRepo { return &IPCandidateRepo{db: db} }

func (r *IPCandidateRepo) Create(ctx context.Context, c *entity.IPCandidate) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *IPCandidateRepo) FindByID(ctx context.Context, id uint64) (*entity.IPCandidate, error) {
	var c entity.IPCandidate
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("IPCandidateRepo.FindByID: %w", err)
	}
	return &c, nil
}

func (r *IPCandidateRepo) FindByDomainAndIP(ctx context.Context, domainID uint64, ip string) (*entity.IPCandidate, error) {
	var c entity.IPCandidate
	if err := r.db.WithContext(ctx).Where("domain_id = ? AND ip = ?", domainID, ip).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("IPCandidateRepo.FindByDomainAndIP: %w", err)
	}
	return &c, nil
}

func (r *IPCandidateRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.IPCandidate, error) {
	if limit <= 0 {
		limit = 20
	}
	var cs []entity.IPCandidate
	q := r.db.WithContext(ctx)
	if afterID > 0 {
		q = q.Where("id < ?", afterID)
	}
	return cs, q.Order("id DESC").Limit(limit).Find(&cs).Error
}

func (r *IPCandidateRepo) ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.IPCandidate, error) {
	if limit <= 0 {
		limit = 20
	}
	var cs []entity.IPCandidate
	q := r.db.WithContext(ctx).Where("domain_id = ?", domainID)
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return cs, q.Order("id ASC").Limit(limit).Find(&cs).Error
}

func (r *IPCandidateRepo) ListEligibleByDomainAndISP(ctx context.Context, domainID uint64, isp string) ([]entity.IPCandidate, error) {
	var cs []entity.IPCandidate
	return cs, r.db.WithContext(ctx).
		Where("domain_id = ? AND isp = ? AND status IN ('eligible','published','locked')", domainID, isp).
		Find(&cs).Error
}

func (r *IPCandidateRepo) Update(ctx context.Context, c *entity.IPCandidate) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *IPCandidateRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.IPCandidate{}).Error
}

// CircuitBreakerRepo ──────────────────────────────────────────────────────

type CircuitBreakerRepo struct{ db *gorm.DB }

func NewCircuitBreakerRepo(db *gorm.DB) *CircuitBreakerRepo { return &CircuitBreakerRepo{db: db} }

func (r *CircuitBreakerRepo) Upsert(ctx context.Context, cb *entity.SchedulerCircuitBreaker) error {
	return r.db.WithContext(ctx).Save(cb).Error
}

func (r *CircuitBreakerRepo) FindByDomainAndISP(ctx context.Context, domainID uint64, isp string) (*entity.SchedulerCircuitBreaker, error) {
	var cb entity.SchedulerCircuitBreaker
	if err := r.db.WithContext(ctx).
		Where("domain_id = ? AND isp = ?", domainID, isp).First(&cb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CircuitBreakerRepo.FindByDomainAndISP: %w", err)
	}
	return &cb, nil
}
