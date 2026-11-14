package ssl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	cryptopkg "github.com/wuekevin/smartcdn/internal/infra/crypto"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

// Service 管理 SSL 凭据中心和证书部署任务。
type Service struct {
	credentials port.SSLCredentialRepository
	deployJobs  port.SSLDeployJobRepository
	certs       port.CertificateRepository
	deployReg   port.SSLDeployRegistry
}

func NewService(
	credentials port.SSLCredentialRepository,
	deployJobs port.SSLDeployJobRepository,
	certs port.CertificateRepository,
	deployReg port.SSLDeployRegistry,
) *Service {
	return &Service{
		credentials: credentials,
		deployJobs:  deployJobs,
		certs:       certs,
		deployReg:   deployReg,
	}
}

// --- 凭据中心 ---

type CredentialView struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	ProviderType string    `json:"provider_type"`
	Category     string    `json:"category"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateCredentialInput struct {
	Name         string
	ProviderType string
	Category     string
	Credentials  map[string]string
}

func (s *Service) CreateCredential(ctx context.Context, input CreateCredentialInput) (*CredentialView, error) {
	jsonBytes, err := json.Marshal(input.Credentials)
	if err != nil {
		return nil, fmt.Errorf("credential: 序列化失败: %w", err)
	}
	enc, err := cryptopkg.Encrypt(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("credential: 加密失败: %w", err)
	}
	c := &entity.SSLCredential{
		Name:           input.Name,
		ProviderType:   input.ProviderType,
		Category:       input.Category,
		CredentialsEnc: enc,
	}
	if err := s.credentials.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("credential: 创建失败: %w", err)
	}
	return toCredView(c), nil
}

func (s *Service) ListCredentials(ctx context.Context, category string) ([]CredentialView, error) {
	items, err := s.credentials.List(ctx, category)
	if err != nil {
		return nil, err
	}
	views := make([]CredentialView, len(items))
	for i, c := range items {
		views[i] = *toCredView(&c)
	}
	return views, nil
}

func (s *Service) DeleteCredential(ctx context.Context, id uint64) error {
	if _, err := s.credentials.FindByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return errors.New("凭据不存在")
		}
		return err
	}
	return s.credentials.Delete(ctx, id)
}

func (s *Service) UpdateCredential(ctx context.Context, id uint64, name string, creds map[string]string) (*CredentialView, error) {
	c, err := s.credentials.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		c.Name = name
	}
	if len(creds) > 0 {
		jsonBytes, _ := json.Marshal(creds)
		enc, encErr := cryptopkg.Encrypt(jsonBytes)
		if encErr != nil {
			return nil, fmt.Errorf("credential: 加密失败: %w", encErr)
		}
		c.CredentialsEnc = enc
	}
	if err := s.credentials.Update(ctx, c); err != nil {
		return nil, err
	}
	return toCredView(c), nil
}

// --- 部署任务 ---

type DeployJobView struct {
	ID             uint64     `json:"id"`
	CertID         uint64     `json:"cert_id"`
	CredentialID   *uint64    `json:"credential_id,omitempty"`
	Name           string     `json:"name"`
	DeployType     string     `json:"deploy_type"`
	Status         string     `json:"status"`
	LastDeployedAt *time.Time `json:"last_deployed_at,omitempty"`
	LastError      string     `json:"last_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateDeployJobInput struct {
	CertID       uint64
	Name         string
	DeployType   string
	ConfigJSON   string
	CredentialID *uint64
}

func (s *Service) CreateDeployJob(ctx context.Context, input CreateDeployJobInput) (*DeployJobView, error) {
	j := &entity.SSLDeployJob{
		CertID:       input.CertID,
		Name:         input.Name,
		DeployType:   input.DeployType,
		ConfigJSON:   input.ConfigJSON,
		CredentialID: input.CredentialID,
		Status:       "pending",
	}
	if err := s.deployJobs.Create(ctx, j); err != nil {
		return nil, err
	}
	return toDeployView(j), nil
}

func (s *Service) ListDeployJobs(ctx context.Context, certID uint64) ([]DeployJobView, error) {
	jobs, err := s.deployJobs.ListByCert(ctx, certID)
	if err != nil {
		return nil, err
	}
	views := make([]DeployJobView, len(jobs))
	for i, j := range jobs {
		views[i] = *toDeployView(&j)
	}
	return views, nil
}

func (s *Service) ExecuteDeployJob(ctx context.Context, jobID uint64) error {
	j, err := s.deployJobs.FindByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("部署任务不存在: %w", err)
	}
	cert, err := s.certs.FindByID(ctx, j.CertID)
	if err != nil {
		return fmt.Errorf("证书不存在: %w", err)
	}
	if cert.CertChainPEM == "" {
		return fmt.Errorf("证书尚未颁发，无法部署")
	}

	privKey, err := cryptopkg.Decrypt(cert.PrivateKeyEnc)
	if err != nil {
		return fmt.Errorf("解密私钥失败: %w", err)
	}

	var credBytes []byte
	if j.CredentialID != nil {
		cred, credErr := s.credentials.FindByID(ctx, *j.CredentialID)
		if credErr != nil {
			return fmt.Errorf("凭据不存在: %w", credErr)
		}
		decrypted, decErr := cryptopkg.Decrypt(cred.CredentialsEnc)
		if decErr != nil {
			return fmt.Errorf("解密凭据失败: %w", decErr)
		}
		credBytes = decrypted
	}

	provider, err := s.deployReg.GetDeploy(j.DeployType, credBytes)
	if err != nil {
		return fmt.Errorf("获取部署 Provider 失败: %w", err)
	}

	deployErr := provider.Deploy(ctx, port.SSLDeployInput{
		Domain:     cert.DomainName,
		CertPEM:    cert.CertChainPEM,
		PrivKeyPEM: string(privKey),
		ConfigJSON: j.ConfigJSON,
	})

	now := time.Now()
	j.LastDeployedAt = &now
	if deployErr != nil {
		j.Status = "failed"
		j.LastError = deployErr.Error()
	} else {
		j.Status = "success"
		j.LastError = ""
	}
	_ = s.deployJobs.Update(ctx, j)
	return deployErr
}

func (s *Service) DeleteDeployJob(ctx context.Context, id uint64) error {
	return s.deployJobs.Delete(ctx, id)
}

func toCredView(c *entity.SSLCredential) *CredentialView {
	return &CredentialView{
		ID:           c.ID,
		Name:         c.Name,
		ProviderType: c.ProviderType,
		Category:     c.Category,
		CreatedAt:    c.CreatedAt,
	}
}

func toDeployView(j *entity.SSLDeployJob) *DeployJobView {
	return &DeployJobView{
		ID:             j.ID,
		CertID:         j.CertID,
		CredentialID:   j.CredentialID,
		Name:           j.Name,
		DeployType:     j.DeployType,
		Status:         j.Status,
		LastDeployedAt: j.LastDeployedAt,
		LastError:      j.LastError,
		CreatedAt:      j.CreatedAt,
	}
}
