package auth

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

func init() {
	os.Setenv("SMARTCDN_ENCRYPT_KEY", "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
}

// mockUserRepo 实现 port.UserRepository 接口，用于隔离测试。
type mockUserRepo struct {
	user        *entity.User
	findErr     error
	updateCalls int
}

func (m *mockUserRepo) FindByUsername(_ context.Context, _ string) (*entity.User, error) {
	return m.user, m.findErr
}
func (m *mockUserRepo) FindByID(_ context.Context, _ uint64) (*entity.User, error) { return nil, nil }
func (m *mockUserRepo) UpdateLoginFailed(_ context.Context, _ uint64, _ int, _ interface{}) error {
	m.updateCalls++
	return nil
}
func (m *mockUserRepo) UpdateLastLogin(_ context.Context, _ uint64) error {
	m.updateCalls++
	return nil
}
func (m *mockUserRepo) GetPermissions(_ context.Context, _ uint64) ([]entity.Permission, error) {
	return nil, nil
}

func hashPw(t *testing.T, pw string) string {
	t.Helper()
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func newTestService(r *mockUserRepo) *Service {
	return &Service{users: r, perms: nil}
}

func TestLogin_Success(t *testing.T) {
	r := &mockUserRepo{user: &entity.User{ID: 1, Username: "alice", Password: hashPw(t, "Password1"), IsActive: true}}
	svc := newTestService(r)

	result, err := svc.Login(context.Background(), "alice", "Password1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if result.ExpiresAt.Before(time.Now()) {
		t.Fatal("token should expire in the future")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	r := &mockUserRepo{user: &entity.User{ID: 1, Username: "alice", Password: hashPw(t, "Password1")}}
	svc := newTestService(r)

	_, err := svc.Login(context.Background(), "alice", "WrongPass1")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
	if r.updateCalls == 0 {
		t.Error("expected UpdateLoginFailed to be called")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	r := &mockUserRepo{findErr: repo.ErrNotFound}
	svc := newTestService(r)

	_, err := svc.Login(context.Background(), "ghost", "pw")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for missing user, got %v", err)
	}
}

func TestLogin_AccountLocked(t *testing.T) {
	lockedUntil := time.Now().Add(10 * time.Minute)
	r := &mockUserRepo{user: &entity.User{
		ID:          1,
		Username:    "locked",
		Password:    hashPw(t, "Password1"),
		LockedUntil: &lockedUntil,
	}}
	svc := newTestService(r)

	_, err := svc.Login(context.Background(), "locked", "Password1")
	if !errors.Is(err, ErrAccountLocked) {
		t.Errorf("expected ErrAccountLocked, got %v", err)
	}
}
