package pool

import (
	"sync/atomic"

	"github.com/MouadBensafir/SafeNode/internal/backend"
)

type RoundRobinStrategy struct {
	current atomic.Uint64
}

func (r *RoundRobinStrategy) NextBackend(backends []*backend.Backend) *backend.Backend {
	n := uint64(len(backends))
	if n == 0 {
		return nil
	}

	for {
		old := r.current.Load()
		next := (old + 1) % n
		i := next
		for {
			if backends[i].Alive {
				break
			}
			i = (i + 1) % n
			if i == next {
				return nil
			}
		}

		if r.current.CompareAndSwap(old, i) {
			return backends[i]
		}
	}
}
