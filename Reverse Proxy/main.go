package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var mainPool ServerPool
var cfg ProxyConfig

func main() {
	// Read and Parse config.json
	cfg = SetupConfigurations()

	// Automatically Add backends from config.json
	for _, strURL := range cfg.Backends {
		u, err := url.Parse(strURL)
		if err != nil {
			log.Printf("Skipping invalid URL %s: %v", strURL, err)
			continue
		}
		mainPool.AddBackend(initBackend(u))
	}

	// Admin endpoint
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/backends", backendsHandler)
	adminMux.HandleFunc("/status", statusHandler)

	// Proxy endpoint 
	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("/", handleRequests)

	// Start health checker
	go StartHealthChecker(cfg.HealthCheckFreq)

	//Serve the Admin API
	go func() {
		adminAddr := ":" + fmt.Sprint(cfg.AdminPort)
		log.Printf("Starting Admin API on %s", adminAddr)
		http.ListenAndServe(adminAddr, adminMux)
	} ()

	// Serve the proxy
	proxyAddr := ":" + fmt.Sprint(cfg.Port)
	log.Printf("Starting proxy on %s", proxyAddr)
	http.ListenAndServe(proxyAddr, proxyMux)
}

// Function to initialize a backend with a proxy 
func initBackend(u *url.URL) *Backend {
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error for %s: %v", u, err)
		mainPool.SetBackendStatus(u, false)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	return &Backend{
		URL:          u,
		Alive:        false,
		CurrentConns: 0,
		RevProxy:     proxy,
	}
}

// Function to handle the repeated JSON decoding and URL parsing logic
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

func backendsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse URL from Request
	u, err := parseURLFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		mainPool.AddBackend(initBackend(u))
		w.WriteHeader(http.StatusCreated)
	case http.MethodDelete:
		if !mainPool.RemoveBackend(u) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	backnd := mainPool.GetNextValidPeer()
	if backnd == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	atomic.AddInt64(&backnd.CurrentConns, 1)
	defer atomic.AddInt64(&backnd.CurrentConns, -1)

	backnd.RevProxy.ServeHTTP(w, r)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "not allowed", http.StatusMethodNotAllowed)
        return
    }

    var infos []backendInfo
    active := 0
    

    mainPool.mux.RLock()
    defer mainPool.mux.RUnlock()

    for _, b := range mainPool.Backends {
        conns := atomic.LoadInt64(&b.CurrentConns)
        
        b.mux.RLock()
        alive := b.Alive
        b.mux.RUnlock()

        if alive {
            active++
        }
        infos = append(infos, backendInfo{
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