package entity

import "time"

// SSLCredential 是可复用的凭据，支持 DNS 验证、证书部署、通知等多种用途。
type SSLCredential struct {
	ID             uint64     `gorm:"primarykey;autoIncrement"`
	Name           string     `gorm:"size:255;not null"`
	ProviderType   string     `gorm:"size:64;not null"` // aliyun_dns / cloudflare / nginx_ssh 等
	Category       string     `gorm:"size:32;not null"` // dns | deploy | notify | acme_ca
	CredentialsEnc string     `gorm:"type:text;not null"` // AES-256-GCM 加密的 JSON
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time `gorm:"index"`
}

// SSLDeployJob 记录证书的一个部署目标。
type SSLDeployJob struct {
	ID             uint64     `gorm:"primarykey;autoIncrement"`
	CertID         uint64     `gorm:"not null;index"`
	CredentialID   *uint64
	Name           string     `gorm:"size:255;not null"`
	DeployType     string     `gorm:"size:64;not null"` // nginx_ssh | aliyun_esa_deploy | cloudflare_deploy | local_file
	ConfigJSON     string     `gorm:"type:text;not null"`
	Status         string     `gorm:"size:32;not null;default:'pending'"` // pending|success|failed
	LastDeployedAt *time.Time
	LastError      string `gorm:"type:text"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
