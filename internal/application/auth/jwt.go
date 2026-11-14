package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultTTL = 24 * time.Hour

type Claims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"sub"`
	jwt.RegisteredClaims
}

var (
	jwtKey     []byte
	jwtKeyOnce sync.Once
	jwtKeyMu   sync.RWMutex
)

// SetJWTSecret configures the JWT signing key. Called at startup from main.
func SetJWTSecret(key []byte) {
	jwtKeyMu.Lock()
	defer jwtKeyMu.Unlock()
	jwtKey = key
}

func jwtSecret() []byte {
	jwtKeyMu.RLock()
	defer jwtKeyMu.RUnlock()
	return jwtKey
}

// IssueToken creates a signed JWT for the given user.
func IssueToken(userID uint64, username string) (string, time.Time, error) {
	expiresAt := time.Now().Add(defaultTTL)
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret())
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing JWT: %w", err)
	}
	return signed, expiresAt, nil
}

// ParseToken validates and parses a JWT string.
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
