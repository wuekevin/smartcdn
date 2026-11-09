package entity

import "time"

// PublishJobStatus tracks progress through the 13-step DNS publish pipeline.
type PublishJobStatus string

const (
	PublishStatusPending    PublishJobStatus = "pending"
	PublishStatusRunning    PublishJobStatus = "running"
	PublishStatusVerifying  PublishJobStatus = "verifying"
	PublishStatusDone       PublishJobStatus = "done"
	PublishStatusFailed     PublishJobStatus = "failed"
	PublishStatusRolledBack PublishJobStatus = "rolled_back"
)

// PublishJob represents a single DNS publish request going through the 13-step engine.
type PublishJob struct {
	ID         uint64           `gorm:"primarykey;autoIncrement"`
	DomainID   uint64           `gorm:"not null;index"`
	Status     PublishJobStatus `gorm:"size:32;not null;default:'pending'"`
	TargetHost string           `gorm:"size:253;not null"`  // subdomain or "@"
	TargetType string           `gorm:"size:16;not null"`   // A, AAAA, CNAME, ...
	TargetLine string           `gorm:"size:64"`             // ISP line, empty = default
	PayloadJSON string          `gorm:"type:json;not null"` // []PublishRecord
	SnapshotID  uint64          `gorm:"index"`              // pre-publish snapshot ID
	DryRun      bool            `gorm:"not null;default:false"`
	StepLogs    string          `gorm:"type:json"`  // []StepLog
	ErrorMsg    string          `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	FinishedAt  *time.Time
}

// PublishRecord is the per-record payload stored in PayloadJSON.
type PublishRecord struct {
	Host     string `json:"host"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	Line     string `json:"line"`
	Priority int    `json:"priority"`
}

// StepLog records the outcome of one engine step.
type StepLog struct {
	Step    string    `json:"step"`
	Status  string    `json:"status"` // ok | skipped | failed
	Message string    `json:"message,omitempty"`
	At      time.Time `json:"at"`
}
