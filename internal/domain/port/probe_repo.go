package port

import (
	"context"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// ProbeAgentRepository abstracts persistence for Probe Agent registrations.
type ProbeAgentRepository interface {
	Create(ctx context.Context, a *entity.ProbeAgent) error
	FindByID(ctx context.Context, id uint64) (*entity.ProbeAgent, error)
	FindByCertSN(ctx context.Context, certSN string) (*entity.ProbeAgent, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.ProbeAgent, error)
	Update(ctx context.Context, a *entity.ProbeAgent) error
}

// ProbeResultRepository abstracts persistence for probe measurements.
type ProbeResultRepository interface {
	Create(ctx context.Context, r *entity.ProbeResult) error
	// ListRecentByIP returns the latest N results for a given IP from a time window.
	ListRecentByIP(ctx context.Context, ip string, since time.Time, limit int) ([]entity.ProbeResult, error)
	// ListRecentByIPAndISP returns the latest N results filtered by ISP.
	ListRecentByIPAndISP(ctx context.Context, ip, isp string, since time.Time, limit int) ([]entity.ProbeResult, error)
}

// IPCandidateRepository abstracts persistence for IP candidates.
type IPCandidateRepository interface {
	Create(ctx context.Context, c *entity.IPCandidate) error
	FindByID(ctx context.Context, id uint64) (*entity.IPCandidate, error)
	FindByDomainAndIP(ctx context.Context, domainID uint64, ip string) (*entity.IPCandidate, error)
	ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.IPCandidate, error)
	ListEligibleByDomainAndISP(ctx context.Context, domainID uint64, isp string) ([]entity.IPCandidate, error)
	Update(ctx context.Context, c *entity.IPCandidate) error
	Delete(ctx context.Context, id uint64) error
}

// CircuitBreakerRepository abstracts persistence for scheduler circuit breakers.
type CircuitBreakerRepository interface {
	Upsert(ctx context.Context, cb *entity.SchedulerCircuitBreaker) error
	FindByDomainAndISP(ctx context.Context, domainID uint64, isp string) (*entity.SchedulerCircuitBreaker, error)
}
