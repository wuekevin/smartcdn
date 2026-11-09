package entity

import "time"

// CDNNode represents a self-hosted reverse-proxy node.
type CDNNode struct {
	ID            uint64     `gorm:"primarykey;autoIncrement"`
	Name          string     `gorm:"size:128;not null"`
	IP            string     `gorm:"size:45;not null;uniqueIndex"` // IPv4 or IPv6
	ISP           string     `gorm:"size:64;not null"`             // 电信/联通/移动/BGP
	Region        string     `gorm:"size:64;not null"`             // 上海/北京/香港/etc.
	OriginAddr    string     `gorm:"size:512"`                     // upstream origin
	OriginPort    int        `gorm:"default:80"`
	IsActive      bool       `gorm:"not null;default:true"`
	AuthToken     string     `gorm:"size:128;uniqueIndex;not null"` // mTLS bootstrap token
	AuthConfirmed bool       `gorm:"not null;default:false"`        // operator agreed to auth

	// HTTPS check result
	HTTPSStatus   string     `gorm:"size:32"`  // ok / failed / pending
	HTTPSCode     int
	TLSDomain     string     `gorm:"size:253"`
	HTTPSLatencyMs int64
	HTTPSCheckedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

// NodeAgent stores per-node Agent registration and last heartbeat.
type NodeAgent struct {
	ID         uint64     `gorm:"primarykey;autoIncrement"`
	NodeID     uint64     `gorm:"not null;uniqueIndex"`
	ClientCertSN string   `gorm:"size:128"` // TLS client cert serial number
	LastSeenAt *time.Time
	CPUPercent float64
	MemPercent float64
	HTTPSStatus string    `gorm:"size:32"` // ok / degraded / down
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NodeTaskStatus enumerates possible task states.
type NodeTaskStatus string

const (
	NodeTaskPending   NodeTaskStatus = "pending"
	NodeTaskRunning   NodeTaskStatus = "running"
	NodeTaskSucceeded NodeTaskStatus = "succeeded"
	NodeTaskFailed    NodeTaskStatus = "failed"
	NodeTaskForbidden NodeTaskStatus = "forbidden"
)

// NodeTask is a management task dispatched to a Node Agent.
type NodeTask struct {
	ID         uint64         `gorm:"primarykey;autoIncrement"`
	NodeID     uint64         `gorm:"not null;index"`
	Type       string         `gorm:"size:64;not null"` // e.g. "reload_nginx"
	Payload    string         `gorm:"type:text"`
	Status     NodeTaskStatus `gorm:"size:32;not null;default:'pending'"`
	ResultMsg  string         `gorm:"type:text"`
	DispatchedAt time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NodePool groups nodes into a named resource pool.
type NodePool struct {
	ID        uint64     `gorm:"primarykey;autoIncrement"`
	Name      string     `gorm:"size:128;not null;uniqueIndex"`
	IsActive  bool       `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

// NodePoolMember is the M2M join table between pools and nodes.
type NodePoolMember struct {
	ID       uint64 `gorm:"primarykey;autoIncrement"`
	PoolID   uint64 `gorm:"not null;uniqueIndex:ux_pool_node"`
	NodeID   uint64 `gorm:"not null;uniqueIndex:ux_pool_node;index"`
	Priority int    `gorm:"default:10"`
}
