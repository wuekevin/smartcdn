package provider

import (
	"fmt"
	"sync"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// Registry implements port.DNSProviderRegistry.
// Factory functions are registered at startup by infra/wire or main; the application layer
// never imports this package directly — it receives the interface.
type Registry struct {
	mu        sync.RWMutex
	factories map[port.ProviderType]func(credentials []byte) (port.DNSProvider, error)
}

func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[port.ProviderType]func(credentials []byte) (port.DNSProvider, error)),
	}
}

// Register adds a factory function for the given provider type.
// Calling Register twice for the same type overwrites the previous factory.
func (r *Registry) Register(pt port.ProviderType, factory func(credentials []byte) (port.DNSProvider, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[pt] = factory
}

// GetDNS creates a provider instance using the factory registered for pt.
// credentials is the raw (decrypted) credential JSON passed to the factory.
func (r *Registry) GetDNS(pt port.ProviderType) (port.DNSProvider, error) {
	r.mu.RLock()
	factory, ok := r.factories[pt]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", port.ErrProviderNotFound, pt)
	}
	// Credentials are injected by the caller (ProviderService), not stored in the registry.
	return factory(nil)
}

// GetDNSWithCredentials creates a provider instance and passes credentials to its factory.
func (r *Registry) GetDNSWithCredentials(pt port.ProviderType, credentials []byte) (port.DNSProvider, error) {
	r.mu.RLock()
	factory, ok := r.factories[pt]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", port.ErrProviderNotFound, pt)
	}
	return factory(credentials)
}

// Registered returns a snapshot of the registered provider type IDs.
func (r *Registry) Registered() []port.ProviderType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]port.ProviderType, 0, len(r.factories))
	for k := range r.factories {
		out = append(out, k)
	}
	return out
}
