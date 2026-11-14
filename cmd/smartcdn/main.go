package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"go.uber.org/zap"

	"github.com/wuekevin/smartcdn/internal/application/auth"
	appconfig "github.com/wuekevin/smartcdn/internal/infra/config"
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
	"github.com/wuekevin/smartcdn/migrations"
)

func main() {
	portFlag := flag.String("port", "", "HTTP listen port (overrides PORT env and config.yaml)")
	flag.Parse()

	// Setup mode: logger at info level, gin in debug for easier troubleshooting.
	if err := logger.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "FATAL: logger init:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.Named("main")
	log.Info("SmartCDN starting")

	setupSvc := setup.NewService()

	if !setupSvc.IsInstalled() {
		log.Info("install.lock not found — running in setup mode")
		gin.SetMode(gin.DebugMode)
		r := buildRouter()
		setupHandler := api.NewSetupHandler(setupSvc)
		setupHandler.RegisterSetup(r.Group("/api/v1"))
		api.RegisterSPA(r, frontendFS())
		startServer(r, resolvePort(*portFlag, setupPort()), log)
		return
	}

	// --- Full application mode ---
	cfg := mustLoadConfig(log)

	// Re-init logger at the level from config.yaml.
	if err := logger.InitWithLevel(cfg.LogLevel); err != nil {
		log.Fatal("logger re-init", zap.Error(err))
	}
	log = logger.Named("main")

	// Apply gin mode from config.yaml (default: release).
	ginMode := cfg.Mode
	if ginMode == "" {
		ginMode = gin.ReleaseMode
	}
	gin.SetMode(ginMode)

	if err := crypto.InitWithKey(cfg.EncryptKey); err != nil {
		log.Fatal("encryption key error", zap.Error(err))
	}

	// Use encrypt key as JWT signing secret (both are derived from the same configured key).
	auth.SetJWTSecret([]byte(cfg.EncryptKey))

	gdb := mustConnectDB(cfg, log)
	rdb := mustConnectRedis(cfg, log)

	r := buildRouter()
	applyTrustedProxies(r, cfg.TrustedProxies, log)
	setupHandler := api.NewSetupHandler(setupSvc)
	v1 := r.Group("/api/v1")
	setupHandler.RegisterSetup(v1)

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

	startServer(r, resolvePort(*portFlag, cfg.Port), log)
}

func mustLoadConfig(log *zap.Logger) *appconfig.AppConfig {
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatal("load config", zap.Error(err))
	}
	return cfg
}

func mustConnectDB(cfg *appconfig.AppConfig, log *zap.Logger) *gorm.DB {
	gdb, err := infradb.New(infradb.Config{DSN: cfg.Database.DSN()})
	if err != nil {
		log.Fatal("DB connect failed", zap.Error(err))
	}
	return gdb
}

func mustConnectRedis(cfg *appconfig.AppConfig, log *zap.Logger) *redis.Client {
	rdb, err := infraredis.New(infraredis.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatal("Redis connect failed", zap.Error(err))
	}
	return rdb
}

func runMigrations(gdb *gorm.DB, log *zap.Logger) {
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Fatal("get sql.DB for migrations", zap.Error(err))
	}
	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatal("goose set dialect", zap.Error(err))
	}
	// Use embedded SQL files so the binary is self-contained.
	goose.SetBaseFS(migrations.FS)
	defer goose.SetBaseFS(nil)
	if err := goose.Up(sqlDB, "."); err != nil {
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

func buildRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(api.RequestIDMiddleware())
	r.Use(api.SecurityHeadersMiddleware())
	return r
}

// applyTrustedProxies configures gin's trusted proxy list.
// Empty slice → trust only loopback (safe default).
// ["*"] → trust all (not recommended).
func applyTrustedProxies(r *gin.Engine, proxies []string, log *zap.Logger) {
	if len(proxies) == 0 {
		// Trust only loopback by default — silences the "trust all proxies" warning.
		if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
			log.Warn("set trusted proxies", zap.Error(err))
		}
		return
	}
	if err := r.SetTrustedProxies(proxies); err != nil {
		log.Warn("set trusted proxies", zap.Error(err))
	}
}

func startServer(r *gin.Engine, port string, log *zap.Logger) {
	log.Info("listening", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}

// setupPort returns the port for setup mode (no config.yaml yet).
// Reads PORT env var; defaults to 8080.
func setupPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}

// resolvePort returns flagVal when non-empty, otherwise fallback.
// This lets -port override both env var and config.yaml.
func resolvePort(flagVal, fallback string) string {
	if flagVal != "" {
		return flagVal
	}
	return fallback
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
