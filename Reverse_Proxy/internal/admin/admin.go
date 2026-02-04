package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/MouadBensafir/SafeNode/internal/backend"
	"github.com/MouadBensafir/SafeNode/internal/pool"
	"github.com/MouadBensafir/SafeNode/internal/proxy"
)

// Handler for POST/DELETE Requests in /backends
func BackendsHandler(mainPool *pool.ServerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodDelete {
			http.Error(w, "not allowed", http.StatusMethodNotAllowed)
			return
		}

		u, err := parseURLFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodPost:
			mainPool.AddBackend(proxy.NewBackend(u, mainPool))
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			if !mainPool.RemoveBackend(u) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}
}

// Handler for GET Request in /status
func StatusHandler(mainPool *pool.ServerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "not allowed", http.StatusMethodNotAllowed)
			return
		}

		var infos []backend.BackendInfo
		active := 0

		for _, b := range mainPool.BackendsSnapshot() {
			conns := atomic.LoadInt64(&b.CurrentConns)
			alive := b.IsAlive()
			if alive {
				active++
			}
			infos = append(infos, backend.BackendInfo{
				URL:                b.URL.String(),
				Alive:              alive,
				CurrentConnections: conns,
			})
		}

		resp := map[string]interface{}{
			"total_backends":  len(infos),
			"active_backends": active,
			"backends":        infos,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// Helper function for parsing url
func parseURLFromRequest(r *http.Request) (*url.URL, error) {
	var payload struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid json")
	}
	u, err := url.Parse(payload.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid url")
	}
	return u, nil
}
