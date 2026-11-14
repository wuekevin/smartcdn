package entity

import "time"

// MigrationJobStatus tracks Mode-B DNS migration progress.
type MigrationJobStatus string

const (
	MigrationPending    MigrationJobStatus = "pending"
	MigrationRunning    MigrationJobStatus = "running"
	MigrationVerifying  MigrationJobStatus = "verifying"
	MigrationCutover    MigrationJobStatus = "cutover"
	MigrationDone       MigrationJobStatus = "done"
	MigrationFailed     MigrationJobStatus = "failed"
	MigrationRolledBack MigrationJobStatus = "rolled_back"
)

// MigrationJob represents a non-disruptive DNS provider-to-provider migration (Mode B).
// Mode B sequence: export_source → import_target → verify_target → cutover → cleanup_source
type MigrationJob struct {
	ID              uint64             `gorm:"primarykey;autoIncrement"`
	DomainID        uint64             `gorm:"not null;index"`
	SourceAccountID uint64             `gorm:"not null"` // origin DNS provider account
	TargetAccountID uint64             `gorm:"not null"` // destination DNS provider account
	Status          MigrationJobStatus `gorm:"size:32;not null;default:'pending'"`
	SnapshotID      uint64             `gorm:"index"`    // pre-migration snapshot
	StepLogs        string             `gorm:"type:text"` // TEXT avoids MySQL JSON validation on empty string
	ErrorMsg        string             `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FinishedAt      *time.Time
}

// MultiDNSConfig enables dual-authority mode for a domain:
// all record writes are mirrored to both primary and secondary providers.
type MultiDNSConfig struct {
	ID                 uint64    `gorm:"primarykey;autoIncrement"`
	DomainID           uint64    `gorm:"uniqueIndex;not null"`
	PrimaryAccountID   uint64    `gorm:"not null"`
	SecondaryAccountID uint64    `gorm:"not null"`
	SyncEnabled        bool      `gorm:"not null;default:true"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
