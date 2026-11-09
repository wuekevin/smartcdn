package provider

import (
	"fmt"
	"sync"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// CDNRegistry holds registered CDN provider factories (same pattern as DNS Registry).
type CDNRegistry struct {
	mu        sync.RWMutex
	factories map[port.CDNProviderType]func(credentials []byte) (port.CDNProvider, error)
}

func NewCDNRegistry() *CDNRegistry {
	return &CDNRegistry{factories: make(map[port.CDNProviderType]func([]byte) (port.CDNProvider, error))}
}

// RegisterCDN registers a CDN provider factory function.
func (r *CDNRegistry) RegisterCDN(pt port.CDNProviderType, factory func([]byte) (port.CDNProvider, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[pt] = factory
}

// GetCDN returns a CDN provider instance with nil credentials (for capability checks).
func (r *CDNRegistry) GetCDN(pt port.CDNProviderType) (port.CDNProvider, error) {
	return r.GetCDNWithCredentials(pt, nil)
}

// GetCDNWithCredentials returns a CDN provider instance initialised with credentials.
func (r *CDNRegistry) GetCDNWithCredentials(pt port.CDNProviderType, credentials []byte) (port.CDNProvider, error) {
	r.mu.RLock()
	factory, ok := r.factories[pt]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", port.ErrProviderNotFound, pt)
	}
	return factory(credentials)
}
