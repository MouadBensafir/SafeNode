package main

import (
	"flag"
	"os"
	"log"
	"encoding/json"
	"net/http"
	"sync/atomic"	
)

func handleRequests(w http.ResponseWriter, r *http.Request) {
	backnd := mainPool.GetNextValidPeer()
	if backnd == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	atomic.AddInt64(&backnd.CurrentConns, 1)
	defer atomic.AddInt64(&backnd.CurrentConns, -1)

    backnd.RevProxy.ServeHTTP(w, r)
}

var mainPool ServerPool

func main() {
	// Read the flag when running, if provided ofc
	configPath := flag.String("config", "", "Path to the configuration file")
	flag.Parse()
	if *configPath == "" {
		*configPath = "config.json" // Defaut configuration
	}

	// Parse the configurations passed
	var proxyConfig ProxyConfig
	file, _ := os.ReadFile(*configPath)
    if err := json.Unmarshal(file, &proxyConfig); err != nil {
        log.Fatal(err)
    }

	// Serve the proxy
	http.HandleFunc("/", handleRequests)
	http.ListenAndServe(":8081", nil)
}