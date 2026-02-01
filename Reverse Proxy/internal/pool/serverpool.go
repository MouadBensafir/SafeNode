package pool

import (
	"sync"
	"sync/atomic"

	"github.com/MouadBensafir/proxyApp/internal/backend"
)

type ServerPool struct { 
	Backends []*backend.Backend `json:"backends"`
	Current  atomic.Uint64
	mux      sync.RWMutex
}

func (mainPool *ServerPool) BackendsSnapshot() []*backend.Backend {
	mainPool.mux.RLock()
	defer mainPool.mux.RUnlock()
	backends := make([]*backend.Backend, len(mainPool.Backends))
	copy(backends, mainPool.Backends)
	return backends
}
