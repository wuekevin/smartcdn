package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// SSLDeployInput 携带待部署的证书信息。
type SSLDeployInput struct {
	Domain     string
	CertPEM    string
	PrivKeyPEM string
	ConfigJSON string // 目标特定的配置 JSON（来自 SSLDeployJob.ConfigJSON）
}

// SSLDeployProvider 将 TLS 证书部署到目标系统。
type SSLDeployProvider interface {
	DeployType() string
	Deploy(ctx context.Context, input SSLDeployInput) error
}

// SSLDeployRegistry 管理部署 Provider 工厂函数。
type SSLDeployRegistry interface {
	RegisterDeploy(deployType string, factory func(credentials []byte) (SSLDeployProvider, error))
	GetDeploy(deployType string, credentials []byte) (SSLDeployProvider, error)
	ListDeployTypes() []string
}

// SSLCredentialRepository 持久化凭据。
type SSLCredentialRepository interface {
	Create(ctx context.Context, c *entity.SSLCredential) error
	FindByID(ctx context.Context, id uint64) (*entity.SSLCredential, error)
	List(ctx context.Context, category string) ([]entity.SSLCredential, error)
	Update(ctx context.Context, c *entity.SSLCredential) error
	Delete(ctx context.Context, id uint64) error
}

// SSLDeployJobRepository 持久化部署任务。
type SSLDeployJobRepository interface {
	Create(ctx context.Context, j *entity.SSLDeployJob) error
	FindByID(ctx context.Context, id uint64) (*entity.SSLDeployJob, error)
	ListByCert(ctx context.Context, certID uint64) ([]entity.SSLDeployJob, error)
	Update(ctx context.Context, j *entity.SSLDeployJob) error
	Delete(ctx context.Context, id uint64) error
}
