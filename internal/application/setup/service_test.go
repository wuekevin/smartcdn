package setup

import (
	"context"
	"os"
	"testing"
)

func TestIsInstalled_False(t *testing.T) {
	os.Remove(installLockPath)
	svc := NewService()
	if svc.IsInstalled() {
		t.Fatal("expected IsInstalled=false before lock exists")
	}
}

func TestIsInstalled_True(t *testing.T) {
	if err := os.WriteFile(installLockPath, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(installLockPath)

	svc := NewService()
	if !svc.IsInstalled() {
		t.Fatal("expected IsInstalled=true when lock file exists")
	}
}

func TestCheck_NoConnections(t *testing.T) {
	os.Remove(installLockPath)
	svc := NewService()
	result := svc.Check(context.Background(), "", "")

	// Without DB/Redis config, those checks fail — but Go/perms/clock pass
	if result.GoVer.Status != "ok" {
		t.Errorf("expected go_version=ok, got %s: %s", result.GoVer.Status, result.GoVer.Message)
	}
	if result.Perms.Status != "ok" {
		t.Errorf("expected permissions=ok, got %s", result.Perms.Status)
	}
	if result.Clock.Status != "ok" {
		t.Errorf("expected clock=ok, got %s", result.Clock.Status)
	}
	if result.MySQL.Status != "fail" {
		t.Errorf("expected mysql=fail when no DSN, got %s", result.MySQL.Status)
	}
	if result.Redis.Status != "fail" {
		t.Errorf("expected redis=fail when no addr, got %s", result.Redis.Status)
	}
}

func TestWriteLock_CreatesFile(t *testing.T) {
	defer os.Remove(installLockPath)

	if err := writeLock("http://localhost:8080"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(installLockPath)
	if err != nil {
		t.Fatal("install.lock not created:", err)
	}
	if len(data) == 0 {
		t.Fatal("install.lock is empty")
	}
}
