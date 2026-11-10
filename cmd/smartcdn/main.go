package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"go.uber.org/zap"

	"github.com/wuekevin/smartcdn/internal/application/auth"
	certapp "github.com/wuekevin/smartcdn/internal/application/certificate"
	cdnsvc "github.com/wuekevin/smartcdn/internal/application/cdn"
	domainsvc "github.com/wuekevin/smartcdn/internal/application/domain"
	migsvc "github.com/wuekevin/smartcdn/internal/application/migration"
	monapp "github.com/wuekevin/smartcdn/internal/application/monitor"
	nodeapp "github.com/wuekevin/smartcdn/internal/application/node"
	probeapp "github.com/wuekevin/smartcdn/internal/application/probe"
	providersvc "github.com/wuekevin/smartcdn/internal/application/provider"
	publishsvc "github.com/wuekevin/smartcdn/internal/application/publish"
	"github.com/wuekevin/smartcdn/internal/application/setup"
	wfapp "github.com/wuekevin/smartcdn/internal/application/workflow"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/handler/api"
	"github.com/wuekevin/smartcdn/internal/infra/crypto"
	infradb "github.com/wuekevin/smartcdn/internal/infra/db"
	"github.com/wuekevin/smartcdn/internal/infra/logger"
	infraprovider "github.com/wuekevin/smartcdn/internal/infra/provider"
	infraredis "github.com/wuekevin/smartcdn/internal/infra/redis"
	"github.com/wuekevin/smartcdn/internal/infra/pubsub"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

func main() {
	if err := crypto.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "FATAL:", err)
		os.Exit(1)
	}
	if err := logger.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "FATAL: logger init:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.Named("main")
	log.Info("SmartCDN starting")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(api.RequestIDMiddleware())
	r.Use(api.SecurityHeadersMiddleware())

	v1 := r.Group("/api/v1")

	setupSvc := setup.NewService()
	setupHandler := api.NewSetupHandler(setupSvc)
	setupHandler.RegisterSetup(v1)

	if !setupSvc.IsInstalled() {
		log.Info("install.lock not found — running in setup mode")
		api.RegisterSPA(r, frontendFS())
		startServer(r, log)
		return
	}

	// --- Full application mode ---
	gdb := mustConnectDB(log)
	rdb := mustConnectRedis(log)

	runMigrations(gdb, log)

	// --- Repositories ---
	dnsAccountRepo := repo.NewDNSAccountRepo(gdb)
	userRepo := repo.NewUserRepo(gdb)
	domainRepo := repo.NewDomainRepo(gdb)
	dnsRecordRepo := repo.NewDNSRecordRepo(gdb)
	dnsSnapshotRepo := repo.NewDNSSnapshotRepo(gdb)
	publishJobRepo := repo.NewPublishJobRepo(gdb)
	cdnAccountRepo := repo.NewCDNAccountRepo(gdb)
	cdnDomainRepo := repo.NewCDNDomainRepo(gdb)
	cdnHostnameRepo := repo.NewCDNCustomHostnameRepo(gdb)
	cdnNodeRepo := repo.NewCDNNodeRepo(gdb)
	nodeAgentRepo := repo.NewNodeAgentRepo(gdb)
	nodeTaskRepo := repo.NewNodeTaskRepo(gdb)
	nodePoolRepo := repo.NewNodePoolRepo(gdb)
	probeAgentRepo := repo.NewProbeAgentRepo(gdb)
	probeResultRepo := repo.NewProbeResultRepo(gdb)
	ipCandidateRepo := repo.NewIPCandidateRepo(gdb)
	circuitBreakerRepo := repo.NewCircuitBreakerRepo(gdb)
	certRepo := repo.NewCertificateRepo(gdb)
	certDeployTargetRepo := repo.NewCertDeployTargetRepo(gdb)
	acmeAccountRepo := repo.NewACMEAccountRepo(gdb)
	wfTemplateRepo := repo.NewWorkflowTemplateRepo(gdb)
	wfRunRepo := repo.NewWorkflowRunRepo(gdb)
	wfNodeRunRepo := repo.NewWorkflowNodeRunRepo(gdb)
	metricRepo := repo.NewMetricRepo(gdb)
	alertRuleRepo := repo.NewAlertRuleRepo(gdb)
	alertRepo := repo.NewAlertRepo(gdb)
	alertChannelRepo := repo.NewAlertChannelRepo(gdb)
	migJobRepo := repo.NewMigrationJobRepo(gdb)
	multiDNSConfigRepo := repo.NewMultiDNSConfigRepo(gdb)
	auditRepo := repo.NewAuditRepo(gdb)

	// --- Infra ---
	dnsRegistry := infraprovider.NewRegistry()
	cdnRegistry := infraprovider.NewCDNRegistry()
	hub := pubsub.NewHub()

	// --- Application services ---
	permSvc := auth.NewPermissionService(userRepo, rdb)
	authSvc := auth.NewService(userRepo, permSvc)
	providerSvc := providersvc.NewService(dnsAccountRepo, dnsRegistry, rdb)
	domainSvc := domainsvc.NewService(domainRepo, dnsRecordRepo, dnsSnapshotRepo, dnsAccountRepo, dnsRegistry)
	cdnSvc := cdnsvc.NewService(cdnAccountRepo, cdnDomainRepo, cdnHostnameRepo, cdnRegistry)
	publishEngine := publishsvc.NewEngine(publishJobRepo, domainRepo, dnsRecordRepo, dnsSnapshotRepo, dnsAccountRepo, dnsRegistry, rdb)
	nodeSvc := nodeapp.NewService(cdnNodeRepo, nodeAgentRepo, nodeTaskRepo, nodePoolRepo)
	probeSvc := probeapp.NewService(probeAgentRepo, probeResultRepo, ipCandidateRepo)
	_ = circuitBreakerRepo
	certSvc := certapp.NewService(certRepo, certDeployTargetRepo, acmeAccountRepo, &providerDNSAdapter{svc: providerSvc})
	wfEngine := wfapp.NewEngine(wfTemplateRepo, wfRunRepo, wfNodeRunRepo)
	migEngine := migsvc.NewEngine(migJobRepo, domainRepo, dnsRecordRepo, dnsSnapshotRepo, dnsAccountRepo, dnsRegistry)
	multiDNSSvc := migsvc.NewMultiDNSService(multiDNSConfigRepo, domainRepo, dnsAccountRepo, dnsRegistry)
	monitorSvc := monapp.NewService(metricRepo, alertRuleRepo, alertRepo, alertChannelRepo)

	// --- Handlers ---
	authHandler := api.NewAuthHandler(authSvc)
	dnsAccountHandler := api.NewDNSAccountHandler(providerSvc)
	domainHandler := api.NewDomainHandler(domainSvc)
	cdnHandler := api.NewCDNHandler(cdnSvc)
	publishHandler := api.NewPublishHandler(publishEngine)
	nodeHandler := api.NewNodeHandler(nodeSvc)
	agentHandler := api.NewAgentHandler(nodeSvc, probeSvc)
	certHandler := api.NewCertificateHandler(certSvc)
	wfHandler := api.NewWorkflowHandler(wfEngine)
	migHandler := api.NewMigrationHandler(migEngine, multiDNSSvc)
	monitorHandler := api.NewMonitorHandler(monitorSvc)
	sseHandler := api.NewSSEHandler(hub)
	dashboardHandler := api.NewDashboardHandler(gdb)
	userHandler := api.NewUserHandler(userRepo)
	auditLogsHandler := api.NewAuditLogsHandler(auditRepo)

	// Public auth routes
	authHandler.RegisterAuth(v1)

	// Protected routes (JWT required + CSRF + audit logging)
	protected := v1.Group("")
	protected.Use(api.JWTMiddleware())
	protected.Use(api.CSRFMiddleware(rdb))
	protected.Use(api.AuditLoggerMiddleware(auditRepo))

	dashboardHandler.RegisterRoutes(protected)
	dnsAccountHandler.RegisterRoutes(protected)
	domainHandler.RegisterRoutes(protected)
	cdnHandler.RegisterRoutes(protected)
	publishHandler.RegisterRoutes(protected)
	nodeHandler.RegisterRoutes(protected)
	certHandler.RegisterRoutes(protected)
	wfHandler.RegisterRoutes(protected)
	migHandler.RegisterRoutes(protected)
	monitorHandler.RegisterRoutes(protected)
	sseHandler.RegisterRoutes(protected)
	userHandler.RegisterRoutes(protected)
	auditLogsHandler.RegisterRoutes(protected)

	// Agent/probe routes use token auth (not JWT)
	agentHandler.RegisterRoutes(v1)

	api.RegisterSPA(r, frontendFS())

	// Background: periodic alert evaluation
	go runAlertEvaluator(monitorSvc, log)

	startServer(r, log)
}

func mustConnectDB(log *zap.Logger) *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is required")
	}
	gdb, err := infradb.New(infradb.Config{DSN: dsn})
	if err != nil {
		log.Fatal("DB connect failed", zap.Error(err))
	}
	return gdb
}

func mustConnectRedis(log *zap.Logger) *redis.Client {
	rawURL := os.Getenv("REDIS_URL")
	if rawURL == "" {
		rawURL = "redis://localhost:6379/0"
	}
	addr, pass, db := parseRedisURL(rawURL)
	rdb, err := infraredis.New(infraredis.Config{Addr: addr, Password: pass, DB: db})
	if err != nil {
		log.Fatal("Redis connect failed", zap.Error(err))
	}
	return rdb
}

func parseRedisURL(rawURL string) (addr, password string, db int) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, "", 0
	}
	addr = u.Host
	if u.User != nil {
		password, _ = u.User.Password()
	}
	if len(u.Path) > 1 {
		n, _ := strconv.Atoi(u.Path[1:])
		db = n
	}
	return
}

func runMigrations(gdb *gorm.DB, log *zap.Logger) {
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Fatal("get sql.DB for migrations", zap.Error(err))
	}
	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatal("goose set dialect", zap.Error(err))
	}
	migrationsDir := "migrations"
	if d := os.Getenv("MIGRATIONS_DIR"); d != "" {
		migrationsDir = d
	}
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		log.Fatal("goose up", zap.Error(err))
	}
	log.Info("database migrations applied")
}

func runAlertEvaluator(svc *monapp.Service, log *zap.Logger) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		svc.EvaluateAll(context.Background())
	}
}

func startServer(r *gin.Engine, log *zap.Logger) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Info("listening", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}

// providerDNSAdapter adapts provider.Service to certificate.DNSProviderResolver.
type providerDNSAdapter struct {
	svc *providersvc.Service
}

func (a *providerDNSAdapter) GetDNSChallenge(ctx context.Context, dnsAccountID uint64) (certapp.DNSChallengeProvider, error) {
	p, err := a.svc.GetProvider(ctx, dnsAccountID)
	if err != nil {
		return nil, err
	}
	return &dnsChallengeImpl{p: p}, nil
}

type dnsChallengeImpl struct {
	p port.DNSProvider
}

func (d *dnsChallengeImpl) Present(domain, token, keyAuth string) error {
	_, err := d.p.CreateRecord(context.Background(), domain, port.DNSRecordInput{
		Host:  "_acme-challenge",
		Type:  "TXT",
		Value: keyAuth,
		TTL:   120,
	})
	return err
}

func (d *dnsChallengeImpl) CleanUp(domain, token, keyAuth string) error {
	return nil
}
