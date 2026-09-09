package extension

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

var (
	ErrInvalidModule   = errors.New("extension: invalid service module")
	ErrDuplicateModule = errors.New("extension: service module already registered")
	ErrModuleNotFound  = errors.New("extension: service module not found")
)

type Registry struct {
	mu      sync.RWMutex
	modules map[marketplace.ServiceType]ServiceModule
}

func NewRegistry(modules ...ServiceModule) (*Registry, error) {
	r := &Registry{modules: make(map[marketplace.ServiceType]ServiceModule)}
	for _, module := range modules {
		if err := r.Register(module); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func MustRegistry(modules ...ServiceModule) *Registry {
	r, err := NewRegistry(modules...)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *Registry) Register(module ServiceModule) error {
	if module == nil {
		return ErrInvalidModule
	}
	manifest := module.Manifest()
	if !manifest.Valid() {
		return ErrInvalidModule
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.modules[manifest.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateModule, manifest.ID)
	}
	r.modules[manifest.ID] = module
	return nil
}

func (r *Registry) Get(serviceType marketplace.ServiceType) (ServiceModule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	module, ok := r.modules[serviceType]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrModuleNotFound, serviceType)
	}
	return module, nil
}

func (r *Registry) Manifests() []Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Manifest, 0, len(r.modules))
	for _, module := range r.modules {
		out = append(out, module.Manifest())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
