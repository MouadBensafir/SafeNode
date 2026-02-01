package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/MouadBensafir/SafeNode/internal/admin"
	"github.com/MouadBensafir/SafeNode/internal/config"
	"github.com/MouadBensafir/SafeNode/internal/healthcheck"
	"github.com/MouadBensafir/SafeNode/internal/pool"
	"github.com/MouadBensafir/SafeNode/internal/proxy"
)

func main() {
	// Read and Parse config.json
	cfg := config.SetupConfigurations()
	mainPool := &pool.ServerPool{}

	// Automatically Add backends from config.json
	for _, strURL := range cfg.Backends {
		u, err := url.Parse(strURL)
		if err != nil {
			log.Printf("Skipping invalid URL %s: %v", strURL, err)
			continue
		}
		mainPool.AddBackend(proxy.NewBackend(u, mainPool))
	}

	// Admin endpoint
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/backends", admin.BackendsHandler(mainPool))
	adminMux.HandleFunc("/status", admin.StatusHandler(mainPool))

	// Proxy endpoint
	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("/", proxy.Handler(mainPool))

	// Start health checker
	go healthcheck.StartHealthChecker(cfg.HealthCheckFreq, cfg.HealthCheckPath, mainPool)

	//Serve the Admin API
	go func() {
		adminAddr := ":" + fmt.Sprint(cfg.AdminPort)
		log.Printf("Starting Admin API on %s", adminAddr)
		http.ListenAndServeTLS(adminAddr, "cert.pem", "key.pem", adminMux)
	}()

	// Serve the proxy
	proxyAddr := ":" + fmt.Sprint(cfg.Port)
	log.Printf("Starting proxy on %s", proxyAddr)
	err := http.ListenAndServeTLS(proxyAddr, "cert.pem", "key.pem" , proxyMux)
	if err != nil {
		log.Fatal(err)
	}
}
