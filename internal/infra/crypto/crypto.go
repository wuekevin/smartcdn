package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

var ErrDecryptionFailed = errors.New("decryption failed: invalid ciphertext or key")

var (
	globalKey  []byte
	initOnce   sync.Once
	initErr    error
)

// Init loads the encryption key from SMARTCDN_ENCRYPT_KEY env var (32-byte hex).
// Must be called before Encrypt/Decrypt.
func Init() error {
	initOnce.Do(func() {
		raw := os.Getenv("SMARTCDN_ENCRYPT_KEY")
		if raw == "" {
			initErr = fmt.Errorf("SMARTCDN_ENCRYPT_KEY is required (32-byte hex)")
			return
		}
		key, err := hex.DecodeString(raw)
		if err != nil || len(key) != 32 {
			initErr = fmt.Errorf("SMARTCDN_ENCRYPT_KEY must be a 64-character hex string (32 bytes)")
			return
		}
		globalKey = key
	})
	return initErr
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// Output format: base64(nonce[12] || ciphertext).
func Encrypt(plaintext []byte) (string, error) {
	if globalKey == nil {
		return "", fmt.Errorf("crypto not initialized: call Init() first")
	}

	block, err := aes.NewCipher(globalKey)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded nonce||ciphertext produced by Encrypt.
func Decrypt(encoded string) ([]byte, error) {
	if globalKey == nil {
		return nil, fmt.Errorf("crypto not initialized: call Init() first")
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	block, err := aes.NewCipher(globalKey)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrDecryptionFailed
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptString is a convenience wrapper for string values.
func EncryptString(s string) (string, error) {
	return Encrypt([]byte(s))
}

// DecryptString is a convenience wrapper returning a string.
func DecryptString(encoded string) (string, error) {
	b, err := Decrypt(encoded)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
