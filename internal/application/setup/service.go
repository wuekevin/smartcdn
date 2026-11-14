package setup

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	appconfig "github.com/wuekevin/smartcdn/internal/infra/config"
	"github.com/wuekevin/smartcdn/internal/infra/logger"
)

const installLockPath = "install.lock"

// Service handles the installation wizard logic.
type Service struct{}

func NewService() *Service { return &Service{} }

// IsInstalled reports whether install.lock exists.
func (s *Service) IsInstalled() bool {
	_, err := os.Stat(installLockPath)
	return err == nil
}

// Check runs environment prerequisite checks.
// MySQL and Redis are only probed when the caller supplies DSN/addr;
// if omitted they are marked pending (not a failure) so step-1 of the
// wizard can pass before the user has entered connection details.
// Port availability is not checked here — the server is already running
// on its configured port, so the check would always fail.
func (s *Service) Check(ctx context.Context, dbDSN, redisAddr string) entity.SetupCheckResult {
	return entity.SetupCheckResult{
		GoVer: checkGoVersion(),
		Perms: checkFilePermissions(),
		Clock: checkClock(),
		Ports: entity.CheckItem{Name: "ports", Status: "ok", Message: "server is running"},
		MySQL: checkMySQL(dbDSN),
		Redis: checkRedis(ctx, redisAddr),
	}
}

// Complete finalizes the installation: validates connections, runs migrations,
// creates admin user, and writes install.lock.
func (s *Service) Complete(ctx context.Context, input entity.SetupCompleteInput) error {
	// Validate Redis connectivity before touching the database.
	redisCheck := checkRedisWithAuth(ctx, input.RedisAddr, input.RedisPassword, input.RedisDB)
	if redisCheck.Status == "fail" {
		return fmt.Errorf("redis: %s", redisCheck.Message)
	}

	dsn := buildDSN(input)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if err := runMigrations(db); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := createAdminUser(db, input.AdminUser, input.AdminPass); err != nil {
		return fmt.Errorf("creating admin user: %w", err)
	}

	if err := writeConfig(input); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := writeLock(); err != nil {
		return fmt.Errorf("writing install.lock: %w", err)
	}

	logger.Named("setup").Info("installation completed")
	return nil
}

// --- internal helpers ---

func checkGoVersion() entity.CheckItem {
	version := runtime.Version()
	if strings.HasPrefix(version, "go1.") {
		return entity.CheckItem{Name: "go_version", Status: "ok", Message: version}
	}
	return entity.CheckItem{Name: "go_version", Status: "fail", Message: "unexpected Go version: " + version}
}

func checkFilePermissions() entity.CheckItem {
	f, err := os.CreateTemp(".", ".smartcdn-perm-check-*")
	if err != nil {
		return entity.CheckItem{Name: "permissions", Status: "fail", Message: "cannot write to working directory"}
	}
	f.Close()
	os.Remove(f.Name())
	return entity.CheckItem{Name: "permissions", Status: "ok"}
}

func checkClock() entity.CheckItem {
	now := time.Now()
	if now.Year() < 2024 {
		return entity.CheckItem{Name: "clock", Status: "fail", Message: fmt.Sprintf("system clock looks wrong: %s", now.Format(time.RFC3339))}
	}
	return entity.CheckItem{Name: "clock", Status: "ok", Message: now.Format(time.RFC3339)}
}


func checkMySQL(dsn string) entity.CheckItem {
	if dsn == "" {
		return entity.CheckItem{Name: "mysql", Status: "ok", Message: "not yet configured"}
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return entity.CheckItem{Name: "mysql", Status: "fail", Message: "invalid DSN"}
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return entity.CheckItem{Name: "mysql", Status: "fail", Message: "connection refused"}
	}
	return entity.CheckItem{Name: "mysql", Status: "ok"}
}

func checkRedis(ctx context.Context, addr string) entity.CheckItem {
	if addr == "" {
		return entity.CheckItem{Name: "redis", Status: "ok", Message: "not yet configured"}
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return entity.CheckItem{Name: "redis", Status: "fail", Message: "connection refused"}
	}
	return entity.CheckItem{Name: "redis", Status: "ok"}
}

func checkRedisWithAuth(ctx context.Context, addr, password string, db int) entity.CheckItem {
	if addr == "" {
		return entity.CheckItem{Name: "redis", Status: "ok", Message: "not yet configured"}
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return entity.CheckItem{Name: "redis", Status: "fail", Message: "connection failed: " + err.Error()}
	}
	return entity.CheckItem{Name: "redis", Status: "ok"}
}

func buildDSN(input entity.SetupCompleteInput) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		input.DBUser, input.DBPassword, input.DBHost, input.DBPort, input.DBName)
}

func runMigrations(db *gorm.DB) error {
	// AutoMigrate the full User schema so the main app's queries (which reference
	// deleted_at, login_failed_count, etc.) work correctly after setup.
	// Goose runs CREATE TABLE IF NOT EXISTS on startup and will skip the already-created table.
	return db.AutoMigrate(&entity.User{})
}

func createAdminUser(db *gorm.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.Exec(
		`INSERT INTO users (username, password, created_at, updated_at) VALUES (?, ?, NOW(), NOW())
		 ON DUPLICATE KEY UPDATE updated_at=NOW()`,
		username, string(hash),
	).Error
}

func writeConfig(input entity.SetupCompleteInput) error {
	cfg := &appconfig.AppConfig{
		Database: appconfig.DatabaseConfig{
			Host:     input.DBHost,
			Port:     input.DBPort,
			Name:     input.DBName,
			User:     input.DBUser,
			Password: input.DBPassword,
		},
		Redis: appconfig.RedisConfig{
			Addr:     input.RedisAddr,
			Password: input.RedisPassword,
			DB:       input.RedisDB,
		},
		EncryptKey:     input.EncryptKey,
		SiteURL:        input.SiteURL,
		Port:           input.Port,
		Mode:           "release",
		LogLevel:       "info",
		TrustedProxies: []string{},
	}
	return appconfig.Write(cfg)
}

func writeLock() error {
	content := fmt.Sprintf("installed=%s\n", time.Now().Format(time.RFC3339))
	return os.WriteFile(installLockPath, []byte(content), 0600)
}
