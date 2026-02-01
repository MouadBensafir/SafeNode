package main

import (
	"net/url"
)

type LoadBalancer interface {
	GetNextValidPeer() *Backend
	AddBackend(backend *Backend)
	SetBackendStatus(url *url.URL, alive bool)
}


// Method to get the next valid peer concurrently
func (mainPool *ServerPool) GetNextValidPeer() *Backend {
	// If there are no backends configured
	n := uint64(len(mainPool.Backends))
	if n == 0 {
		return nil
	}

	// Loop until we can successfully update the current index and return an alive backend
	for {
		old := mainPool.Current.Load()
		// Start searching from the next index
		next := (old + 1) % n
		i := next
		for {
			if mainPool.Backends[i].Alive {
				break
			}
			i = (i + 1) % n
			if i == next {
				return nil // No backend is alive case
			}
		}

		// Attempt to set current to the chosen index i
		if mainPool.Current.CompareAndSwap(old, i) {
			return mainPool.Backends[i]
		}

		// Otherwise Retry
	}
}

// Method to add a Backend
func (mainPool *ServerPool) AddBackend(backend *Backend) {
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
	backends := append([]*Backend(nil), mainPool.Backends...)
	mainPool.mux.RUnlock()

	for _, backend := range backends {
		if backend.URL != nil && backend.URL.String() == target.String() {
			backend.SetAlive(alive)
			return
		}
	}
}
