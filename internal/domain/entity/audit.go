package entity

import "time"

// AuditLog is append-only — no deleted_at column (AD-15).
type AuditLog struct {
	ID           uint64    `gorm:"primarykey;autoIncrement"`
	OperatorID   *uint64   `gorm:"column:operator_id"`
	OperatorIP   string    `gorm:"column:operator_ip;size:45"`
	Action       string    `gorm:"size:128;not null"`
	ResourceType string    `gorm:"column:resource_type;size:64"`
	ResourceID   string    `gorm:"column:resource_id;size:64"`
	BeforeState  *string   `gorm:"column:before_state;type:json"`
	AfterState   *string   `gorm:"column:after_state;type:json"`
	CreatedAt    time.Time `gorm:"not null"`
}

// TableName prevents GORM from soft-deleting this table.
func (AuditLog) TableName() string { return "audit_logs" }
