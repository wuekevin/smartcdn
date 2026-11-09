package entity

import "time"

// CDNAccount stores credentials for a CDN provider.
type CDNAccount struct {
	ID             uint64     `gorm:"primarykey;autoIncrement"`
	Name           string     `gorm:"size:128;not null"`
	ProviderType   string     `gorm:"size:64;not null"`
	CredentialsEnc string     `gorm:"type:text;not null"` // AES-256-GCM ciphertext
	IsActive       bool       `gorm:"not null;default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time `gorm:"index"`
}

// CDNDomain tracks a CDN-accelerated domain.
type CDNDomain struct {
	ID           uint64     `gorm:"primarykey;autoIncrement"`
	AccountID    uint64     `gorm:"not null;index"`
	DomainName   string     `gorm:"uniqueIndex;size:253;not null"`
	ProviderZoneID string   `gorm:"size:255"`  // provider-side zone/domain ID
	CNAME        string     `gorm:"size:512"`   // CDN CNAME target
	Status       string     `gorm:"size:32;not null;default:'pending'"`
	OriginAddr   string     `gorm:"size:512"`
	OriginPort   int        `gorm:"default:80"`
	Protocol     string     `gorm:"size:16;default:'follow'"`
	IsActive     bool       `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

// CDNCustomHostname stores a Cloudflare SaaS custom hostname record.
type CDNCustomHostname struct {
	ID             uint64    `gorm:"primarykey;autoIncrement"`
	CDNDomainID    uint64    `gorm:"not null;index"`    // FK → cdn_domains.id (SaaS zone)
	Hostname       string    `gorm:"size:253;not null"` // customer hostname
	ProviderID     string    `gorm:"size:255"`          // Cloudflare custom hostname ID
	Status         string    `gorm:"size:64"`
	DCVTXTName     string    `gorm:"size:512"`
	DCVTXTValue    string    `gorm:"size:512"`
	DCVHTTPPath    string    `gorm:"size:512"`
	DCVHTTPValue   string    `gorm:"size:512"`
	SSLStatus      string    `gorm:"size:64"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
