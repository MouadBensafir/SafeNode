package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/MouadBensafir/SafeNode/internal/backend"
	"github.com/MouadBensafir/SafeNode/internal/pool"
)

// Handler for the reverse proxy
func ProxyHandler(mainPool *pool.ServerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backnd := mainPool.GetNextValidPeer()
		if backnd == nil {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		atomic.AddInt64(&backnd.CurrentConns, 1)
		defer atomic.AddInt64(&backnd.CurrentConns, -1)

		backnd.RevProxy.ServeHTTP(w, r)
	}
}

// Helper Function that initializes a backend with a reverse proxy and error handler
func NewBackend(u *url.URL, mainPool *pool.ServerPool) *backend.Backend {
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy Error for %s: %v", u, err)
		mainPool.SetBackendStatus(u, false)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	return &backend.Backend{
		URL:          u,
		Alive:        false,
		CurrentConns: 0,
		RevProxy:     proxy,
	}
}