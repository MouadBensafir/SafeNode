package pool

import 	"github.com/MouadBensafir/SafeNode/internal/backend"


type LeastConnStrategy struct {}

func (l *LeastConnStrategy) NextBackend(backends []*backend.Backend) *backend.Backend {
	// TODO : Implement least conn logic
	return nil
}