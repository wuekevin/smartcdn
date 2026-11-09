package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

const (
	maxFailedAttempts = 5
	lockDuration      = 15 * time.Minute
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountLocked      = errors.New("account temporarily locked")
)

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
}

type Service struct {
	users port.UserRepository
	perms *PermissionService
}

func NewService(users port.UserRepository, perms *PermissionService) *Service {
	return &Service{users: users, perms: perms}
}

// Login verifies credentials, enforces lock policy, and issues a JWT.
func (s *Service) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			// 不区分用户名错误和密码错误，防止用户名枚举
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, fmt.Errorf("%w: unlocks at %s", ErrAccountLocked, user.LockedUntil.Format(time.RFC3339))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		count := user.LoginFailedCount + 1
		var lockedUntil interface{}
		if count >= maxFailedAttempts {
			t := time.Now().Add(lockDuration)
			lockedUntil = t
		}
		_ = s.users.UpdateLoginFailed(ctx, user.ID, count, lockedUntil)
		return nil, ErrInvalidCredentials
	}

	_ = s.users.UpdateLastLogin(ctx, user.ID)

	token, expiresAt, err := IssueToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: token, ExpiresAt: expiresAt}, nil
}
