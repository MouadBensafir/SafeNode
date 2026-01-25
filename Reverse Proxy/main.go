package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync/atomic"
)

var mainPool ServerPool

func main() {
	cfg := SetupConfigurations()

	targetStr := "http://localhost:8081"
	targetURL, _ := url.Parse(targetStr)

	// Generating the instance
	b1 := &Backend{
		URL:          targetURL,
		Alive:        true,
		CurrentConns: 0,
		RevProxy:     httputil.NewSingleHostReverseProxy(targetURL),
	}

	mainPool.Backends = append(mainPool.Backends, b1)

	// Serve the proxy
	http.HandleFunc("/", handleRequests)
	addr := ":8080"
	if cfg.Port != 0 {
		addr = ":" + fmt.Sprint(cfg.Port)
	}
	http.ListenAndServe(addr, nil)
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

func SetupConfigurations() ProxyConfig {
	// Read the flag when running, if provided ofc
	configPath := flag.String("config", "", "Path to the configuration file")
	flag.Parse()
	if *configPath == "" {
		*configPath = "config.json" // Defaut configuration
	}

	// Parse the configurations passed
	var proxyConfig ProxyConfig
	fileData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Printf("could not read config file %s: %v, using defaults", *configPath, err)
		return proxyConfig
	}
	
	if err := json.Unmarshal(fileData, &proxyConfig); err != nil {
		log.Printf("invalid config file %s: %v, using defaults", *configPath, err)
		return proxyConfig
	}

	return proxyConfig
}
