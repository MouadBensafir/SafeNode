package pool

import (
	"net/url"

	"github.com/MouadBensafir/SafeNode/internal/backend"
)

type LoadBalancer interface {
	GetNextValidPeer() *backend.Backend
	AddBackend(backend *backend.Backend)
	SetBackendStatus(url *url.URL, alive bool)
	SetStrategy(strategy BalancingStrategy)
}

// Method to get the next valid peer concurrently
func (mainPool *ServerPool) GetNextValidPeer() *backend.Backend {
	backends := mainPool.BackendsSnapshot()
	if len(backends) == 0 {
		return nil
	}

	// Incase no strategy is chosen
	mainPool.mux.RLock()
	strategy := mainPool.Strategy
	mainPool.mux.RUnlock()

	return strategy.NextBackend(backends)
}

// Method to add a Backend
func (mainPool *ServerPool) AddBackend(backend *backend.Backend) {
	mainPool.mux.Lock()
	mainPool.Backends = append(mainPool.Backends, backend)
	mainPool.mux.Unlock()
}

// Method to remove a Backend using its url
func (mainPool *ServerPool) RemoveBackend(target *url.URL) bool {
	mainPool.mux.Lock()
	defer mainPool.mux.Unlock()
	for i, backend := range mainPool.Backends {
		if backend.URL != nil && backend.URL.String() == target.String() {
			mainPool.Backends = append(mainPool.Backends[:i], mainPool.Backends[i+1:]...)
			return true
		}
	}
	return false
}

// Method to Change a Backend's status
func (mainPool *ServerPool) SetBackendStatus(target *url.URL, alive bool) {
	mainPool.mux.RLock()
	backends := append([]*backend.Backend(nil), mainPool.Backends...)
	mainPool.mux.RUnlock()

	for _, backend := range backends {
		if backend.URL != nil && backend.URL.String() == target.String() {
			backend.SetAlive(alive)
			return
		}
	}
}
