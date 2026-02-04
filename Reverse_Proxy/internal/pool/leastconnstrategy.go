package pool

import (
    "math"

    "github.com/MouadBensafir/SafeNode/internal/backend"
)

type LeastConnStrategy struct {}

func (l *LeastConnStrategy) NextBackend(backends []*backend.Backend) *backend.Backend {
    var best *backend.Backend
    var min int64 = math.MaxInt64

    for _, backnd := range backends {
        if !backnd.Alive {
            continue
        }

        conn := backnd.CurrentConns 

        if conn < min {
            min = conn
            best = backnd
        }
    }

    return best
}