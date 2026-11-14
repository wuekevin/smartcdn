package deploy

import (
	"fmt"
	"sync"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// Registry 实现 port.SSLDeployRegistry。
type Registry struct {
	mu        sync.RWMutex
	factories map[string]func(credentials []byte) (port.SSLDeployProvider, error)
}

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]func([]byte) (port.SSLDeployProvider, error))}
}

func (r *Registry) RegisterDeploy(deployType string, factory func([]byte) (port.SSLDeployProvider, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[deployType] = factory
}

func (r *Registry) GetDeploy(deployType string, credentials []byte) (port.SSLDeployProvider, error) {
	r.mu.RLock()
	factory, ok := r.factories[deployType]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("deploy provider not found: %s", deployType)
	}
	return factory(credentials)
}

func (r *Registry) ListDeployTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

var _ port.SSLDeployRegistry = (*Registry)(nil)
