package entity

import "time"

// ─── Metric ───────────────────────────────────────────────────────────────────

// MetricType classifies the kind of data point.
type MetricType string

const (
	MetricTypeLatency    MetricType = "latency_ms"
	MetricTypeTTFB       MetricType = "ttfb_ms"
	MetricTypeSuccessRate MetricType = "success_rate"
	MetricTypeUptime     MetricType = "uptime_pct"
	MetricTypeRPS        MetricType = "rps"
	MetricTypeCPU        MetricType = "cpu_pct"
	MetricTypeMemory     MetricType = "mem_pct"
	MetricTypeDiskIO     MetricType = "disk_io_kbps"
)

// Metric is a single time-series data point collected from a node or probe.
type Metric struct {
	ID         uint64     `gorm:"primarykey;autoIncrement"`
	SourceType string     `gorm:"size:32;not null"` // node | probe | domain
	SourceID   uint64     `gorm:"not null;index"`
	Type       MetricType `gorm:"size:64;not null"`
	Value      float64    `gorm:"not null"`
	Tags       string     `gorm:"size:512"` // JSON key=value pairs, e.g. {"isp":"ChinaTelecom"}
	CollectedAt time.Time `gorm:"not null;index"`
	CreatedAt  time.Time
}

// ─── AlertRule ────────────────────────────────────────────────────────────────

// AlertCondition is the comparison operator used in an alert rule.
type AlertCondition string

const (
	AlertCondGT  AlertCondition = "gt"  // value > threshold
	AlertCondGTE AlertCondition = "gte" // value >= threshold
	AlertCondLT  AlertCondition = "lt"  // value < threshold
	AlertCondLTE AlertCondition = "lte" // value <= threshold
	AlertCondEQ  AlertCondition = "eq"  // value == threshold
)

// AlertSeverity classifies how critical an alert is.
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// AlertRule defines a threshold condition evaluated over metric time windows.
type AlertRule struct {
	ID           uint64         `gorm:"primarykey;autoIncrement"`
	Name         string         `gorm:"size:128;not null"`
	Description  string         `gorm:"size:512"`
	SourceType   string         `gorm:"size:32;not null"` // node | probe | domain | *
	SourceID     *uint64        `gorm:"index"`             // nil = match all sources of SourceType
	MetricType   MetricType     `gorm:"size:64;not null"`
	Condition    AlertCondition `gorm:"size:16;not null"`
	Threshold    float64        `gorm:"not null"`
	WindowMinutes int           `gorm:"not null;default:5"`  // evaluation window in minutes
	ForMinutes    int           `gorm:"not null;default:1"`  // must remain true for N minutes
	Severity     AlertSeverity  `gorm:"size:32;not null;default:'warning'"`
	Enabled      bool           `gorm:"not null;default:true"`
	// ChannelIDs is a JSON array of AlertChannel IDs to notify.
	ChannelIDs   string    `gorm:"type:text;not null;default:'[]'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

// ─── Alert ────────────────────────────────────────────────────────────────────

// AlertStatus tracks whether an alert is firing or resolved.
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

// Alert is an instance of an AlertRule firing.
type Alert struct {
	ID           uint64        `gorm:"primarykey;autoIncrement"`
	RuleID       uint64        `gorm:"not null;index"`
	RuleName     string        `gorm:"size:128;not null"` // snapshot
	SourceType   string        `gorm:"size:32;not null"`
	SourceID     uint64        `gorm:"not null"`
	MetricType   MetricType    `gorm:"size:64;not null"`
	TriggerValue float64       `gorm:"not null"`
	Threshold    float64       `gorm:"not null"`
	Severity     AlertSeverity `gorm:"size:32;not null"`
	Status       AlertStatus   `gorm:"size:32;not null;default:'firing'"`
	FiredAt      time.Time     `gorm:"not null"`
	ResolvedAt   *time.Time
	Message      string    `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ─── AlertChannel ─────────────────────────────────────────────────────────────

// AlertChannelType is the delivery mechanism for notifications.
type AlertChannelType string

const (
	AlertChannelWebhook  AlertChannelType = "webhook"
	AlertChannelEmail    AlertChannelType = "email"
	AlertChannelDingTalk AlertChannelType = "dingtalk"
	AlertChannelSlack    AlertChannelType = "slack"
)

// AlertChannel stores the configuration for a notification endpoint.
type AlertChannel struct {
	ID         uint64           `gorm:"primarykey;autoIncrement"`
	Name       string           `gorm:"size:128;not null"`
	Type       AlertChannelType `gorm:"size:32;not null"`
	// ConfigJSON holds type-specific settings (URL, token, emails list) encrypted at rest.
	ConfigJSON string    `gorm:"type:text;not null;default:'{}'"`
	Enabled    bool      `gorm:"not null;default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `gorm:"index"`
}
