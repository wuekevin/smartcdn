package entity

import "time"

// CertificateStatus tracks the lifecycle of an SSL certificate.
type CertificateStatus string

const (
	CertStatusPending  CertificateStatus = "pending"
	CertStatusIssuing  CertificateStatus = "issuing"
	CertStatusActive   CertificateStatus = "active"
	CertStatusExpired  CertificateStatus = "expired"
	CertStatusRevoked  CertificateStatus = "revoked"
	CertStatusFailed   CertificateStatus = "failed"
)

// ACMEChallenge is the type of ACME challenge used.
type ACMEChallenge string

const (
	ACMEDNS01  ACMEChallenge = "dns-01"
	ACMEHTTP01 ACMEChallenge = "http-01"
)

// Certificate is an SSL/TLS certificate managed by SmartCDN.
type Certificate struct {
	ID           uint64            `gorm:"primarykey;autoIncrement"`
	DomainID     uint64            `gorm:"not null;index"`
	DomainName   string            `gorm:"size:253;not null"`
	SANs         string            `gorm:"type:text"` // JSON array of additional SANs
	Status       CertificateStatus `gorm:"size:32;not null;default:'pending'"`
	ACMEProvider string            `gorm:"size:64;not null;default:'letsencrypt'"` // letsencrypt / zerossl
	Challenge    ACMEChallenge     `gorm:"size:32;not null;default:'dns-01'"`
	AccountEmail string            `gorm:"size:255;not null"`

	// Encrypted certificate private key (AES-256-GCM)
	PrivateKeyEnc string `gorm:"type:text"`
	// PEM certificate chain
	CertChainPEM  string `gorm:"type:longtext"`

	// Validity window
	NotBefore     *time.Time
	NotAfter      *time.Time
	AutoRenew     bool `gorm:"not null;default:true"`

	// Last issuance attempt details
	LastIssuedAt  *time.Time
	LastError     string `gorm:"type:text"`

	// DNS account to use for DNS-01 challenge
	DNSAccountID *uint64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

// CertDeployTarget records where a certificate has been deployed.
type CertDeployTarget struct {
	ID            uint64     `gorm:"primarykey;autoIncrement"`
	CertificateID uint64     `gorm:"not null;index"`
	TargetType    string     `gorm:"size:64;not null"` // cdn_node / cdn_provider / nginx_local
	TargetID      string     `gorm:"size:255;not null"` // depends on TargetType
	DeployedAt    *time.Time
	Status        string `gorm:"size:32;not null;default:'pending'"` // pending / deployed / failed
	LastError     string `gorm:"type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ACMEAccount stores the ACME registration for an email address.
type ACMEAccount struct {
	ID           uint64 `gorm:"primarykey;autoIncrement"`
	Email        string `gorm:"size:255;not null;uniqueIndex"`
	Provider     string `gorm:"size:64;not null"` // letsencrypt / zerossl
	PrivKeyEnc   string `gorm:"type:text"`        // encrypted ECDSA private key
	RegistrationURL string `gorm:"size:512"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
