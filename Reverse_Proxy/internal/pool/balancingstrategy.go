package pool

import "github.com/MouadBensafir/SafeNode/internal/backend"


type BalancingStrategy interface {
	NextBackend(backends []*backend.Backend) *backend.Backend
}