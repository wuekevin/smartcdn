package entity

import "time"

// Domain represents a managed DNS zone (apex domain).
type Domain struct {
	ID           uint64     `gorm:"primarykey;autoIncrement"`
	Name         string     `gorm:"uniqueIndex;size:253;not null"` // apex domain, e.g. "example.com"
	AccountID    uint64     `gorm:"not null;index"`                 // FK → dns_accounts.id
	ICPFiled     bool       `gorm:"not null;default:false"`         // 大陆备案标志
	MaxPublishIPs int       `gorm:"not null;default:1"`            // 最大发布 IP 数
	IsActive     bool       `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

// DNSRecord is a local copy of a DNS record managed by SmartCDN.
type DNSRecord struct {
	ID         uint64     `gorm:"primarykey;autoIncrement"`
	DomainID   uint64     `gorm:"not null;index"`
	ProviderID string     `gorm:"size:255"` // provider-side record ID
	Host       string     `gorm:"size:253;not null"`
	Type       string     `gorm:"size:16;not null"`
	Value      string     `gorm:"size:4096;not null"`
	TTL        int        `gorm:"not null;default:600"`
	Line       string     `gorm:"size:64"`
	Priority   int
	IsExternal bool       `gorm:"not null;default:false"` // true = not managed by SmartCDN
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `gorm:"index"`
}

// DNSSnapshot is a full snapshot of all records for a domain at a point in time.
type DNSSnapshot struct {
	ID           uint64    `gorm:"primarykey;autoIncrement"`
	DomainID     uint64    `gorm:"not null;index"`
	SnapshotJSON string    `gorm:"type:json;not null"` // JSON array of DNSRecord
	TriggerEvent string    `gorm:"size:128"`           // e.g. "pre_publish", "manual"
	CreatedAt    time.Time
}
