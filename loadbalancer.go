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
	// Repeat operation until the current loaded value is unchanged through the whole process
	for {
		old := mainPool.Current.Load()
		new := old + 1

		// Skip until we can reach an active backend
		n := uint64(len(mainPool.Backends))
		for !mainPool.Backends[new].Alive {
			new = (new+1)%n

			// If all backends are DOWN, return an arror
			if old == new { 
				return nil
			}
		}
		// Ensure that we return the unchanged value of current and update it
		if mainPool.Current.CompareAndSwap(old, new) {
			return mainPool.Backends[old]
		}
	}
}


func (mainPool *ServerPool) AddBackend(backend *Backend) {
	mainPool.Backends = append(mainPool.Backends, backend)
}



func (mainPool *ServerPool) SetBackendStatus(url *url.URL, alive bool) {
	for _, backend := range mainPool.Backends {
		if backend.URL == url {
			backend.Alive = alive
			break
		}
	}
}