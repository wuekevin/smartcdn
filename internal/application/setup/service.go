package setup

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
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

// Check runs all environment checks and returns a report.
func (s *Service) Check(ctx context.Context, dbDSN, redisAddr string) entity.SetupCheckResult {
	result := entity.SetupCheckResult{
		GoVer: checkGoVersion(),
		Perms: checkFilePermissions(),
		Clock: checkClock(),
		Ports: checkPorts([]string{"8080"}),
	}
	result.MySQL = checkMySQL(dbDSN)
	result.Redis = checkRedis(ctx, redisAddr)
	return result
}

// Complete finalizes the installation: runs migrations, creates admin user, writes install.lock.
func (s *Service) Complete(ctx context.Context, input entity.SetupCompleteInput) error {
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

	if err := writeLock(input.SiteURL); err != nil {
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

func checkPorts(ports []string) entity.CheckItem {
	for _, port := range ports {
		ln, err := net.Listen("tcp", ":"+port)
		if err != nil {
			return entity.CheckItem{Name: "ports", Status: "fail", Message: fmt.Sprintf("port %s is in use", port)}
		}
		ln.Close()
	}
	return entity.CheckItem{Name: "ports", Status: "ok"}
}

func checkMySQL(dsn string) entity.CheckItem {
	if dsn == "" {
		return entity.CheckItem{Name: "mysql", Status: "fail", Message: "no DSN provided"}
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
		return entity.CheckItem{Name: "redis", Status: "fail", Message: "no address provided"}
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return entity.CheckItem{Name: "redis", Status: "fail", Message: "connection refused"}
	}
	return entity.CheckItem{Name: "redis", Status: "ok"}
}

func buildDSN(input entity.SetupCompleteInput) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		input.DBUser, input.DBPassword, input.DBHost, input.DBPort, input.DBName)
}

func runMigrations(db *gorm.DB) error {
	// goose migrations are run via the CLI or migration runner at startup.
	// Here we do a minimal auto-migrate of the users table so the setup can create the admin.
	type User struct {
		ID        uint64 `gorm:"primarykey;autoIncrement"`
		Username  string `gorm:"uniqueIndex;size:64;not null"`
		Password  string `gorm:"not null"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	return db.AutoMigrate(&User{})
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

func writeLock(siteURL string) error {
	content := fmt.Sprintf("installed=%s\nsite_url=%s\n", time.Now().Format(time.RFC3339), siteURL)
	return os.WriteFile(installLockPath, []byte(content), 0600)
}
