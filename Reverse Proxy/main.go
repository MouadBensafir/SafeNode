package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var mainPool ServerPool

func main() {
	cfg := SetupConfigurations()

	// If no backends configured yet, add a default one for quick testing
	targetStr := "http://localhost:8081"
	targetURL, _ := url.Parse(targetStr)

	b1 := &Backend{
		URL:          targetURL,
		Alive:        false,
		CurrentConns: 0,
		RevProxy:     httputil.NewSingleHostReverseProxy(targetURL),
	}
	mainPool.Backends = append(mainPool.Backends, b1)

	// Admin endpoints
	http.HandleFunc("/backends", backendsHandler)
	http.HandleFunc("/status", statusHandler)

	// Proxy endpoint (catch-all)
	http.HandleFunc("/", handleRequests)

	// Start health checker
	go StartHealthChecker(cfg.HealthCheckFreq)

	addr := ":8080"
	if cfg.Port != 0 {
		addr = ":" + fmt.Sprint(cfg.Port)
	}
	log.Printf("starting proxy on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
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

// backendsHandler handles POST /backends to add a backend via admin API
// JSON body: { "url": "http://localhost:8082" }
func backendsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var payload struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		u, err := url.Parse(payload.URL)
		if err != nil {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}
		b := &Backend{
			URL:          u,
			Alive:        false,
			CurrentConns: 0,
			RevProxy:     httputil.NewSingleHostReverseProxy(u),
		}
		mainPool.AddBackend(b)
		w.WriteHeader(http.StatusCreated)
		return
	default:
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// statusHandler handles GET /status and returns JSON status of the server pool
func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
		return
	}
	type backendInfo struct {
		URL                string `json:"url"`
		Alive              bool   `json:"alive"`
		CurrentConnections int64  `json:"current_connections"`
	}
	var infos []backendInfo
	total := 0
	active := 0
	for _, b := range mainPool.Backends {
		total++
		b.mux.RLock()
		alive := b.Alive
		conns := b.CurrentConns
		b.mux.RUnlock()
		if alive {
			active++
		}
		infos = append(infos, backendInfo{URL: b.URL.String(), Alive: alive, CurrentConnections: conns})
	}
	resp := map[string]interface{}{
		"total_backends":  total,
		"active_backends": active,
		"backends":        infos,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}


