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

	for _, StringUrl := range cfg.Backends {
		// Create valid URL
		url, _ := url.Parse(StringUrl)

		// Error Handling to mark the backend as dead immediately upon failure
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy error for %s: %v", url, err)
			mainPool.SetBackendStatus(url, false)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		}

		// Instantiate the backend
		b_temp := &Backend{
			URL:          url,
			Alive:        false,
			CurrentConns: 0,
			RevProxy:     proxy,
		}

		// Add it to the server pool
		mainPool.AddBackend(b_temp)
	}

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

			// Error Handling to mark the backend as dead immediately upon failure
			proxy := httputil.NewSingleHostReverseProxy(u)
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("proxy error for %s: %v", u, err)
				mainPool.SetBackendStatus(u, false)
				http.Error(w, "Bad Gateway", http.StatusBadGateway)
			}

			b := &Backend{
				URL:          u,
				Alive:        false,
				CurrentConns: 0,
				RevProxy:     proxy,
			}
			mainPool.AddBackend(b)
			w.WriteHeader(http.StatusCreated)
			return
		case http.MethodDelete:
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
			isFound := mainPool.RemoveBackend(u)

			if !isFound {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
			return
		default:
			http.Error(w, "not allowed", http.StatusMethodNotAllowed)
			return
	}
}

// Handles GET /status and returns JSON status of the server pool
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


