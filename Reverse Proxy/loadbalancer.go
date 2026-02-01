package main

import (
	"net/url"
)

type LoadBalancer interface {
	GetNextValidPeer() *Backend
	AddBackend(backend *Backend)
	SetBackendStatus(url *url.URL, alive bool)
}

func (mainPool *ServerPool) GetNextValidPeer() *Backend {
	// If there are no backends configured, return nil
	n := uint64(len(mainPool.Backends))
	if n == 0 {
		return nil
	}

	// Loop until we can successfully update the current index and return an alive backend
	for {
		old := mainPool.Current.Load()
		// start searching from the next index
		next := (old + 1) % n

		// search for an alive backend starting at `next`
		i := next
		for {
			if mainPool.Backends[i].Alive {
				break
			}
			i = (i + 1) % n
			if i == next {
				// completed a full loop, no backend is alive
				return nil
			}
		}

		// attempt to set current to the chosen index `i`
		if mainPool.Current.CompareAndSwap(old, i) {
			return mainPool.Backends[i]
		}
		// otherwise retry
	}
}

func (mainPool *ServerPool) AddBackend(backend *Backend) {
	mainPool.mux.Lock()
	mainPool.Backends = append(mainPool.Backends, backend)
	mainPool.mux.Unlock()
}

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
