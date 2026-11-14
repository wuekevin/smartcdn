package entity

import "time"

// DNSAccount stores provider credentials (encrypted) and detected capabilities.
type DNSAccount struct {
	ID               uint64     `gorm:"primarykey;autoIncrement"`
	Name             string     `gorm:"size:128;not null"`
	ProviderType     string     `gorm:"size:64;not null"`
	CredentialsEnc   string     `gorm:"type:text;not null"`  // AES-256-GCM encrypted JSON
	CapabilitiesJSON string     `gorm:"type:text"`           // cached capabilities JSON, stored as TEXT to avoid MySQL JSON validation on empty string
	IsActive         bool       `gorm:"not null;default:true"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `gorm:"index"`
}
