package auth

import (
	"os"
	"testing"
	"time"
)

func init() {
	os.Setenv("SMARTCDN_ENCRYPT_KEY", "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
}

func TestIssueAndParseToken(t *testing.T) {
	token, expiresAt, err := IssueToken(42, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if expiresAt.Before(time.Now()) {
		t.Error("expiresAt should be in the future")
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID: got %d, want 42", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("Username: got %s, want alice", claims.Username)
	}
}

func TestParseToken_Invalid(t *testing.T) {
	_, err := ParseToken("not.a.valid.jwt")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestParseToken_Tampered(t *testing.T) {
	token, _, _ := IssueToken(1, "bob")
	tampered := token[:len(token)-4] + "XXXX"
	_, err := ParseToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}
