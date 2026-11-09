package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

// circuitBreakerWindow is the rolling window for switch counting.
const circuitBreakerWindow = 30 * time.Minute

// circuitBreakerThreshold is the number of line-switches that trips a breaker.
const circuitBreakerThreshold = 3

// scoreWeightLatency, TTFB, and success are the scoring formula weights.
const (
	scoreWeightLatency  = 0.40
	scoreWeightTTFB     = 0.40
	scoreWeightSuccess  = 0.20
	maxLatencyMs        = float64(500) // latency beyond this scores 0
	maxTTFBMs           = float64(500)
)

// Service computes scores for IP candidates, enforces circuit breakers, and
// publishes the best IP per domain/ISP to the DNS layer.
type Service struct {
	candidates     port.IPCandidateRepository
	circuitBreakers port.CircuitBreakerRepository
	probeResults   port.ProbeResultRepository
	// dnsPublisher is called when a new IP should be published to DNS.
	// It is provided by the caller (wired in main) to avoid circular imports.
	dnsPublisher DNSPublisher
}

// DNSPublisher abstracts the DNS layer for the scheduler.
type DNSPublisher interface {
	// PublishIPForISP updates the DNS record for domainID/ISP to the given IP.
	PublishIPForISP(ctx context.Context, domainID uint64, isp, ip string) error
}

func NewService(
	candidates port.IPCandidateRepository,
	breakers port.CircuitBreakerRepository,
	probeResults port.ProbeResultRepository,
	publisher DNSPublisher,
) *Service {
	return &Service{
		candidates:      candidates,
		circuitBreakers: breakers,
		probeResults:    probeResults,
		dnsPublisher:    publisher,
	}
}

// --- Scoring ---

// ScoreResult holds the computed score for one candidate.
type ScoreResult struct {
	CandidateID uint64  `json:"candidate_id"`
	IP          string  `json:"ip"`
	ISP         string  `json:"isp"`
	Score       float64 `json:"score"`
	LatencyMs   float64 `json:"latency_ms"`
	TTFBMs      float64 `json:"ttfb_ms"`
	SuccessRate float64 `json:"success_rate"`
}

// ScoreDomain rescores all eligible candidates for a given domain and returns
// the best IP per ISP. It does not publish to DNS; call PublishBestIPs for that.
func (s *Service) ScoreDomain(ctx context.Context, domainID uint64) ([]ScoreResult, error) {
	candidates, err := s.candidates.ListByDomain(ctx, domainID, 0, 200)
	if err != nil {
		return nil, fmt.Errorf("scheduler: list candidates: %w", err)
	}

	var results []ScoreResult
	now := time.Now()
	window := now.Add(-15 * time.Minute)

	for i := range candidates {
		c := &candidates[i]
		if c.Status != entity.IPCandidateEligible && c.Status != entity.IPCandidatePublished {
			continue
		}
		probes, err := s.probeResults.ListRecentByIPAndISP(ctx, c.IP, c.ISP, window, 20)
		if err != nil || len(probes) == 0 {
			continue
		}
		score, latMs, ttfbMs, successRate := computeScore(probes)
		c.Score = score
		now2 := time.Now()
		c.LastScoreAt = &now2
		_ = s.candidates.Update(ctx, c)

		results = append(results, ScoreResult{
			CandidateID: c.ID,
			IP:          c.IP,
			ISP:         c.ISP,
			Score:       score,
			LatencyMs:   latMs,
			TTFBMs:      ttfbMs,
			SuccessRate: successRate,
		})
	}
	return results, nil
}

// PublishBestIPs picks the highest-scoring eligible candidate per ISP for a
// domain and publishes it to DNS, respecting circuit breakers.
func (s *Service) PublishBestIPs(ctx context.Context, domainID uint64) error {
	scores, err := s.ScoreDomain(ctx, domainID)
	if err != nil {
		return err
	}
	// Group by ISP, pick best.
	best := map[string]ScoreResult{}
	for _, sc := range scores {
		if prev, ok := best[sc.ISP]; !ok || sc.Score > prev.Score {
			best[sc.ISP] = sc
		}
	}
	for isp, sc := range best {
		cb, err := s.circuitBreakers.FindByDomainAndISP(ctx, domainID, isp)
		if err == nil && cb.Tripped {
			// Circuit breaker is open; skip auto-publish for this ISP.
			continue
		}
		// Find the currently published candidate.
		current, _ := s.currentPublished(ctx, domainID, isp)
		if current != nil && current.IP == sc.IP {
			continue // already published
		}
		// Switch to new IP.
		if err := s.dnsPublisher.PublishIPForISP(ctx, domainID, isp, sc.IP); err != nil {
			continue
		}
		// Mark old published → eligible.
		if current != nil {
			current.Status = entity.IPCandidateEligible
			_ = s.candidates.Update(ctx, current)
		}
		// Mark new → published.
		newC, err2 := s.candidates.FindByID(ctx, sc.CandidateID)
		if err2 == nil {
			newC.Status = entity.IPCandidatePublished
			_ = s.candidates.Update(ctx, newC)
		}
		// Record switch in circuit breaker.
		_ = s.recordSwitch(ctx, domainID, isp)
	}
	return nil
}

// --- Circuit breaker ---

type CircuitBreakerView struct {
	DomainID    uint64     `json:"domain_id"`
	ISP         string     `json:"isp"`
	Tripped     bool       `json:"tripped"`
	SwitchCount int        `json:"switch_count"`
	TrippedAt   *time.Time `json:"tripped_at,omitempty"`
}

func (s *Service) GetCircuitBreaker(ctx context.Context, domainID uint64, isp string) (*CircuitBreakerView, error) {
	cb, err := s.circuitBreakers.FindByDomainAndISP(ctx, domainID, isp)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return &CircuitBreakerView{DomainID: domainID, ISP: isp}, nil
		}
		return nil, err
	}
	return &CircuitBreakerView{
		DomainID:    cb.DomainID,
		ISP:         cb.ISP,
		Tripped:     cb.Tripped,
		SwitchCount: cb.SwitchCount,
		TrippedAt:   cb.TrippedAt,
	}, nil
}

// ResetCircuitBreaker manually clears a tripped breaker (operator action).
func (s *Service) ResetCircuitBreaker(ctx context.Context, domainID uint64, isp string) error {
	cb, err := s.circuitBreakers.FindByDomainAndISP(ctx, domainID, isp)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil // nothing to reset
		}
		return err
	}
	cb.Tripped = false
	cb.SwitchCount = 0
	cb.WindowStart = time.Now()
	cb.TrippedAt = nil
	return s.circuitBreakers.Upsert(ctx, cb)
}

// --- internal helpers ---

func (s *Service) recordSwitch(ctx context.Context, domainID uint64, isp string) error {
	cb, err := s.circuitBreakers.FindByDomainAndISP(ctx, domainID, isp)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			cb = &entity.SchedulerCircuitBreaker{
				DomainID:    domainID,
				ISP:         isp,
				WindowStart: time.Now(),
			}
		} else {
			return err
		}
	}
	// Reset window if expired.
	if time.Since(cb.WindowStart) > circuitBreakerWindow {
		cb.SwitchCount = 0
		cb.WindowStart = time.Now()
	}
	cb.SwitchCount++
	if cb.SwitchCount >= circuitBreakerThreshold && !cb.Tripped {
		cb.Tripped = true
		now := time.Now()
		cb.TrippedAt = &now
	}
	return s.circuitBreakers.Upsert(ctx, cb)
}

func (s *Service) currentPublished(ctx context.Context, domainID uint64, isp string) (*entity.IPCandidate, error) {
	eligible, err := s.candidates.ListEligibleByDomainAndISP(ctx, domainID, isp)
	if err != nil {
		return nil, err
	}
	for i := range eligible {
		if eligible[i].Status == entity.IPCandidatePublished {
			return &eligible[i], nil
		}
	}
	return nil, fmt.Errorf("no published candidate")
}

// computeScore applies: latency 40% + TTFB 40% + success rate 20%.
// Returns score in [0,1] plus the average latency ms, TTFB ms, and success rate.
func computeScore(probes []entity.ProbeResult) (score, avgLatMs, avgTTFBMs, successRate float64) {
	if len(probes) == 0 {
		return 0, 0, 0, 0
	}
	var totalLat, totalTTFB float64
	var successes int
	for _, p := range probes {
		totalLat += float64(p.LatencyMs)
		totalTTFB += float64(p.TTFBMs)
		if p.Success {
			successes++
		}
	}
	n := float64(len(probes))
	avgLatMs = totalLat / n
	avgTTFBMs = totalTTFB / n
	successRate = float64(successes) / n

	latScore := 1.0 - clamp(avgLatMs/maxLatencyMs, 0, 1)
	ttfbScore := 1.0 - clamp(avgTTFBMs/maxTTFBMs, 0, 1)

	score = latScore*scoreWeightLatency + ttfbScore*scoreWeightTTFB + successRate*scoreWeightSuccess
	return score, avgLatMs, avgTTFBMs, successRate
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
