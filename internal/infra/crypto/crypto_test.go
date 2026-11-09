package crypto

import (
	"os"
	"sync"
	"testing"
)

func setupTestKey(t *testing.T) {
	t.Helper()
	// Reset singleton for each test
	initOnce = sync.Once{}
	globalKey = nil
	initErr = nil
	os.Setenv("SMARTCDN_ENCRYPT_KEY", "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	t.Cleanup(func() {
		initOnce = sync.Once{}
		globalKey = nil
		initErr = nil
		os.Unsetenv("SMARTCDN_ENCRYPT_KEY")
	})
}

func TestInit_MissingKey(t *testing.T) {
	initOnce = sync.Once{}
	globalKey = nil
	initErr = nil
	os.Unsetenv("SMARTCDN_ENCRYPT_KEY")
	defer func() {
		initOnce = sync.Once{}
		globalKey = nil
		initErr = nil
	}()

	if err := Init(); err == nil {
		t.Fatal("expected error for missing SMARTCDN_ENCRYPT_KEY")
	}
}

func TestInit_InvalidKey(t *testing.T) {
	initOnce = sync.Once{}
	globalKey = nil
	initErr = nil
	os.Setenv("SMARTCDN_ENCRYPT_KEY", "tooshort")
	defer func() {
		initOnce = sync.Once{}
		globalKey = nil
		initErr = nil
		os.Unsetenv("SMARTCDN_ENCRYPT_KEY")
	}()

	if err := Init(); err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	setupTestKey(t)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	plaintext := "super-secret-api-key-12345"
	encrypted, err := EncryptString(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("got %q, want %q", decrypted, plaintext)
	}
}

func TestEncrypt_DifferentCiphertextEachCall(t *testing.T) {
	setupTestKey(t)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	plaintext := "same-plaintext"
	c1, _ := EncryptString(plaintext)
	c2, _ := EncryptString(plaintext)

	if c1 == c2 {
		t.Error("two encryptions of the same plaintext must produce different ciphertexts (random nonce)")
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	setupTestKey(t)
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	_, err := DecryptString("not-valid-base64!!!")
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed, got %v", err)
	}

	_, err = DecryptString("dmFsaWRiYXNlNjQ=") // valid base64 but wrong content
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed for tampered ciphertext, got %v", err)
	}
}
