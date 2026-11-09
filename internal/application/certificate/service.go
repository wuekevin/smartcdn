package certificate

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	cryptopkg "github.com/wuekevin/smartcdn/internal/infra/crypto"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

var (
	ErrCertNotFound  = errors.New("certificate not found")
	ErrACMEAccount   = errors.New("ACME account error")
)

// DNSChallengeProvider can set/delete TXT records for a domain.
// It is implemented by each DNS provider adapter.
type DNSChallengeProvider interface {
	// Present creates the DNS-01 challenge TXT record.
	Present(domain, token, keyAuth string) error
	// CleanUp removes the DNS-01 challenge TXT record.
	CleanUp(domain, token, keyAuth string) error
}

// Service manages SSL/TLS certificates using ACME (Let's Encrypt / ZeroSSL).
type Service struct {
	certs         port.CertificateRepository
	deployTargets port.CertDeployTargetRepository
	acmeAccounts  port.ACMEAccountRepository
	dnsProviders  DNSProviderResolver
}

// DNSProviderResolver resolves a DNS challenge provider for a given DNS account ID.
type DNSProviderResolver interface {
	GetDNSChallenge(ctx context.Context, dnsAccountID uint64) (DNSChallengeProvider, error)
}

func NewService(
	certs port.CertificateRepository,
	deployTargets port.CertDeployTargetRepository,
	acmeAccounts port.ACMEAccountRepository,
	dnsProviders DNSProviderResolver,
) *Service {
	return &Service{
		certs:         certs,
		deployTargets: deployTargets,
		acmeAccounts:  acmeAccounts,
		dnsProviders:  dnsProviders,
	}
}

// --- Certificate management ---

type IssueCertInput struct {
	DomainID     uint64
	DomainName   string
	SANs         []string
	ACMEProvider string // "letsencrypt" | "zerossl"
	Challenge    entity.ACMEChallenge
	AccountEmail string
	AutoRenew    bool
	DNSAccountID *uint64
	Staging      bool // use ACME staging environment
}

type CertView struct {
	ID           uint64                  `json:"id"`
	DomainID     uint64                  `json:"domain_id"`
	DomainName   string                  `json:"domain_name"`
	Status       entity.CertificateStatus `json:"status"`
	ACMEProvider string                  `json:"acme_provider"`
	Challenge    entity.ACMEChallenge    `json:"challenge"`
	AccountEmail string                  `json:"account_email"`
	NotBefore    *time.Time              `json:"not_before,omitempty"`
	NotAfter     *time.Time              `json:"not_after,omitempty"`
	AutoRenew    bool                    `json:"auto_renew"`
	LastError    string                  `json:"last_error,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
}

func (s *Service) CreateCertificate(ctx context.Context, input IssueCertInput) (*CertView, error) {
	sans, _ := json.Marshal(input.SANs)
	provider := input.ACMEProvider
	if provider == "" {
		provider = "letsencrypt"
	}
	challenge := input.Challenge
	if challenge == "" {
		challenge = entity.ACMEDNS01
	}
	c := &entity.Certificate{
		DomainID:     input.DomainID,
		DomainName:   input.DomainName,
		SANs:         string(sans),
		Status:       entity.CertStatusPending,
		ACMEProvider: provider,
		Challenge:    challenge,
		AccountEmail: input.AccountEmail,
		AutoRenew:    input.AutoRenew,
		DNSAccountID: input.DNSAccountID,
	}
	if err := s.certs.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("cert: create: %w", err)
	}
	// Issue asynchronously.
	go s.issueCertificate(context.Background(), c.ID, input.Staging)
	return toCertView(c), nil
}

func (s *Service) GetCertificate(ctx context.Context, id uint64) (*CertView, error) {
	c, err := s.certs.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrCertNotFound
		}
		return nil, err
	}
	return toCertView(c), nil
}

func (s *Service) ListCertificates(ctx context.Context, domainID uint64) ([]CertView, error) {
	cs, err := s.certs.FindByDomainID(ctx, domainID)
	if err != nil {
		return nil, err
	}
	views := make([]CertView, len(cs))
	for i, c := range cs {
		views[i] = *toCertView(&c)
	}
	return views, nil
}

func (s *Service) DeleteCertificate(ctx context.Context, id uint64) error {
	if _, err := s.certs.FindByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrCertNotFound
		}
		return err
	}
	return s.certs.Delete(ctx, id)
}

func (s *Service) RetriggerIssuance(ctx context.Context, id uint64, staging bool) error {
	c, err := s.certs.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrCertNotFound
		}
		return err
	}
	c.Status = entity.CertStatusPending
	c.LastError = ""
	if err := s.certs.Update(ctx, c); err != nil {
		return err
	}
	go s.issueCertificate(context.Background(), c.ID, staging)
	return nil
}

// RenewDueSoon finds certificates expiring within daysThreshold days and re-issues them.
func (s *Service) RenewDueSoon(ctx context.Context, daysThreshold int, staging bool) (int, error) {
	cs, err := s.certs.ListDueForRenewal(ctx, daysThreshold, 50)
	if err != nil {
		return 0, err
	}
	for _, c := range cs {
		c.Status = entity.CertStatusPending
		c.LastError = ""
		_ = s.certs.Update(ctx, &c)
		go s.issueCertificate(context.Background(), c.ID, staging)
	}
	return len(cs), nil
}

// --- Deploy targets ---

type AddDeployTargetInput struct {
	CertID     uint64
	TargetType string // cdn_node / cdn_provider / nginx_local
	TargetID   string
}

func (s *Service) AddDeployTarget(ctx context.Context, input AddDeployTargetInput) (*entity.CertDeployTarget, error) {
	t := &entity.CertDeployTarget{
		CertificateID: input.CertID,
		TargetType:    input.TargetType,
		TargetID:      input.TargetID,
		Status:        "pending",
	}
	if err := s.deployTargets.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("cert: add deploy target: %w", err)
	}
	return t, nil
}

func (s *Service) ListDeployTargets(ctx context.Context, certID uint64) ([]entity.CertDeployTarget, error) {
	return s.deployTargets.ListByCert(ctx, certID)
}

func (s *Service) DeleteDeployTarget(ctx context.Context, id uint64) error {
	return s.deployTargets.Delete(ctx, id)
}

// --- ACME issuance (internal) ---

// legoUser implements registration.User for lego.
type legoUser struct {
	email        string
	key          crypto.PrivateKey
	registration *registration.Resource
}

func (u *legoUser) GetEmail() string                        { return u.email }
func (u *legoUser) GetRegistration() *registration.Resource { return u.registration }
func (u *legoUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

func (s *Service) issueCertificate(ctx context.Context, certID uint64, staging bool) {
	c, err := s.certs.FindByID(ctx, certID)
	if err != nil {
		return
	}
	c.Status = entity.CertStatusIssuing
	_ = s.certs.Update(ctx, c)

	certPEM, keyPEM, notBefore, notAfter, issueErr := s.runACME(ctx, c, staging)
	now := time.Now()
	c.LastIssuedAt = &now

	if issueErr != nil {
		c.Status = entity.CertStatusFailed
		c.LastError = issueErr.Error()
		_ = s.certs.Update(ctx, c)
		return
	}

	// Encrypt private key before storing.
	keyEncStr, err := cryptopkg.Encrypt(keyPEM)
	if err != nil {
		c.Status = entity.CertStatusFailed
		c.LastError = "key encryption failed: " + err.Error()
		_ = s.certs.Update(ctx, c)
		return
	}
	c.PrivateKeyEnc = keyEncStr
	c.CertChainPEM = string(certPEM)
	c.Status = entity.CertStatusActive
	c.NotBefore = notBefore
	c.NotAfter = notAfter
	c.LastError = ""
	_ = s.certs.Update(ctx, c)
}

func (s *Service) runACME(ctx context.Context, c *entity.Certificate, staging bool) (certPEM, keyPEM []byte, notBefore, notAfter *time.Time, err error) {
	// 1. Load or create ACME account.
	user, err := s.loadOrCreateACMEUser(ctx, c.AccountEmail, c.ACMEProvider)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("ACME user: %w", err)
	}

	// 2. Build lego client config.
	caURL := legoCAURL(c.ACMEProvider, staging)
	config := lego.NewConfig(user)
	config.CADirURL = caURL
	config.Certificate.KeyType = certcrypto.RSA2048
	config.HTTPClient = &http.Client{Timeout: 60 * time.Second}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("lego client: %w", err)
	}

	// 3. Register (idempotent).
	if user.registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("ACME register: %w", err)
		}
		user.registration = reg
		s.saveACMERegistration(ctx, c.AccountEmail, reg.URI)
	}

	// 4. Set challenge provider.
	switch c.Challenge {
	case entity.ACMEDNS01:
		if c.DNSAccountID == nil {
			return nil, nil, nil, nil, fmt.Errorf("DNS-01 challenge requires a DNS account ID")
		}
		provider, err := s.dnsProviders.GetDNSChallenge(ctx, *c.DNSAccountID)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("DNS challenge provider: %w", err)
		}
		wrapper := &legoDNS01Provider{provider: provider}
		if err := client.Challenge.SetDNS01Provider(wrapper, dns01.AddDNSTimeout(2*time.Minute)); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("set DNS-01 provider: %w", err)
		}
	case entity.ACMEHTTP01:
		// HTTP-01 is not supported for self-hosted CDN nodes without additional infra.
		return nil, nil, nil, nil, fmt.Errorf("HTTP-01 challenge is not supported; use dns-01")
	default:
		return nil, nil, nil, nil, fmt.Errorf("unknown challenge type: %s", c.Challenge)
	}

	// 5. Build domain list.
	domains := []string{c.DomainName}
	if c.SANs != "" {
		var extra []string
		if json.Unmarshal([]byte(c.SANs), &extra) == nil {
			domains = append(domains, extra...)
		}
	}

	// 6. Obtain certificate.
	res, err := client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("ACME obtain: %w", err)
	}

	// 7. Parse validity window from cert chain.
	nb, na := parseCertValidity(res.Certificate)
	return res.Certificate, res.PrivateKey, nb, na, nil
}

func (s *Service) loadOrCreateACMEUser(ctx context.Context, email, provider string) (*legoUser, error) {
	acc, err := s.acmeAccounts.FindByEmail(ctx, email)
	if err == nil {
		// Decrypt stored key.
		keyPEM, err := cryptopkg.Decrypt(acc.PrivKeyEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt ACME key: %w", err)
		}
		privKey, err := parsePEMKey(keyPEM)
		if err != nil {
			return nil, err
		}
		user := &legoUser{email: email, key: privKey}
		if acc.RegistrationURL != "" {
			user.registration = &registration.Resource{URI: acc.RegistrationURL}
		}
		return user, nil
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, err
	}
	// Create new account key.
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	keyDER, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	keyEnc, err := cryptopkg.Encrypt(keyPEM)
	if err != nil {
		return nil, err
	}
	newAcc := &entity.ACMEAccount{
		Email:      email,
		Provider:   provider,
		PrivKeyEnc: keyEnc,
	}
	if err := s.acmeAccounts.Create(ctx, newAcc); err != nil {
		return nil, err
	}
	return &legoUser{email: email, key: privKey}, nil
}

func (s *Service) saveACMERegistration(ctx context.Context, email, regURL string) {
	acc, err := s.acmeAccounts.FindByEmail(ctx, email)
	if err != nil {
		return
	}
	acc.RegistrationURL = regURL
	_ = s.acmeAccounts.Update(ctx, acc)
}

// legoDNS01Provider wraps our DNSChallengeProvider for lego's challenge.Provider interface.
type legoDNS01Provider struct {
	provider DNSChallengeProvider
}

func (p *legoDNS01Provider) Present(domain, token, keyAuth string) error {
	return p.provider.Present(domain, token, keyAuth)
}

func (p *legoDNS01Provider) CleanUp(domain, token, keyAuth string) error {
	return p.provider.CleanUp(domain, token, keyAuth)
}

// --- helpers ---

func legoCAURL(provider string, staging bool) string {
	switch provider {
	case "zerossl":
		return "https://acme.zerossl.com/v2/DV90"
	default: // letsencrypt
		if staging {
			return lego.LEDirectoryStaging
		}
		return lego.LEDirectoryProduction
	}
}

func parsePEMKey(pemBytes []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM key")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}
	return x509.ParsePKCS8PrivateKey(block.Bytes)
}

func parseCertValidity(certPEM []byte) (notBefore, notAfter *time.Time) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil
	}
	nb := cert.NotBefore
	na := cert.NotAfter
	return &nb, &na
}

func toCertView(c *entity.Certificate) *CertView {
	return &CertView{
		ID:           c.ID,
		DomainID:     c.DomainID,
		DomainName:   c.DomainName,
		Status:       c.Status,
		ACMEProvider: c.ACMEProvider,
		Challenge:    c.Challenge,
		AccountEmail: c.AccountEmail,
		NotBefore:    c.NotBefore,
		NotAfter:     c.NotAfter,
		AutoRenew:    c.AutoRenew,
		LastError:    c.LastError,
		CreatedAt:    c.CreatedAt,
	}
}
