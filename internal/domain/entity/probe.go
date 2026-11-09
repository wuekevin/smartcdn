package entity

import "time"

// ProbeAgent stores registration for a distributed Probe Agent.
type ProbeAgent struct {
	ID           uint64     `gorm:"primarykey;autoIncrement"`
	Name         string     `gorm:"size:128;not null"`
	ISP          string     `gorm:"size:64;not null"`  // 电信/联通/移动
	Region       string     `gorm:"size:64;not null"`
	ClientCertSN string     `gorm:"size:128;uniqueIndex"` // mTLS cert serial
	IsActive     bool       `gorm:"not null;default:true"`
	LastSeenAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ProbeResult stores one measurement from a Probe Agent.
type ProbeResult struct {
	ID            uint64    `gorm:"primarykey;autoIncrement"`
	ProbeAgentID  uint64    `gorm:"not null;index"`
	TargetIP      string    `gorm:"size:45;not null;index"`
	ISP           string    `gorm:"size:64;not null"` // which ISP this result is for
	LatencyMs     int64     // TCP round-trip latency
	TTFBMs        int64     // time to first byte
	HTTPSStatus   int       // HTTP status code (0 = unreachable)
	Success       bool
	ProbeAt       time.Time `gorm:"index"`
	CreatedAt     time.Time
}

// IPCandidateStatus enumerates the lifecycle states of an IP candidate.
type IPCandidateStatus string

const (
	IPCandidateCandidate IPCandidateStatus = "candidate" // just added
	IPCandidateEligible  IPCandidateStatus = "eligible"  // passed HTTPS check
	IPCandidatePublished IPCandidateStatus = "published" // currently live in DNS
	IPCandidateOffline   IPCandidateStatus = "offline"   // consecutive failures
	IPCandidateLocked    IPCandidateStatus = "locked"    // operator-pinned
)

// IPCandidate is a node IP tracked per domain per line (ISP).
type IPCandidate struct {
	ID            uint64            `gorm:"primarykey;autoIncrement"`
	DomainID      uint64            `gorm:"not null;index"`
	NodeID        uint64            `gorm:"not null;index"`
	IP            string            `gorm:"size:45;not null"`
	ISP           string            `gorm:"size:64;not null"` // 电信/联通/移动/默认
	Status        IPCandidateStatus `gorm:"size:32;not null;default:'candidate'"`
	FailStreak    int               `gorm:"default:0"` // consecutive HTTPS check failures
	Score         float64           `gorm:"default:0"` // last computed score
	LastScoreAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SchedulerCircuitBreaker tracks the auto-schedule circuit-breaker per domain/ISP.
type SchedulerCircuitBreaker struct {
	ID         uint64     `gorm:"primarykey;autoIncrement"`
	DomainID   uint64     `gorm:"not null;uniqueIndex:ux_cb_domain_isp"`
	ISP        string     `gorm:"size:64;not null;uniqueIndex:ux_cb_domain_isp"`
	Tripped    bool       `gorm:"not null;default:false"`
	SwitchCount int       `gorm:"default:0"` // switches in rolling window
	WindowStart time.Time
	TrippedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
