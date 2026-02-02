package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MouadBensafir/SafeNode/internal/admin"
	"github.com/MouadBensafir/SafeNode/internal/config"
	"github.com/MouadBensafir/SafeNode/internal/healthcheck"
	"github.com/MouadBensafir/SafeNode/internal/pool"
	"github.com/MouadBensafir/SafeNode/internal/proxy"
)

func main() {
	// Global main server pool
	mainPool := &pool.ServerPool{}

	// Read and Parse config.json
	cfg := config.SetupConfigurations(mainPool)
	
	// Admin endpoint
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/backends", admin.BackendsHandler(mainPool))
	adminMux.HandleFunc("/status", admin.StatusHandler(mainPool))

	// Proxy endpoint
	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("/", proxy.ProxyHandler(mainPool))

	// Start Health Checker
	go healthcheck.StartHealthChecker(cfg.HealthCheckFreq, cfg.HealthCheckPath, mainPool)

	//Serve the Admin API
	go func() {
		adminAddr := ":" + fmt.Sprint(cfg.AdminPort)
		log.Printf("Starting Admin API on %s", adminAddr)
		err := http.ListenAndServeTLS(adminAddr, "cert.pem", "key.pem", adminMux)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Serve the Proxy
	proxyAddr := ":" + fmt.Sprint(cfg.Port)
	log.Printf("Starting proxy on %s", proxyAddr)
	err := http.ListenAndServeTLS(proxyAddr, "cert.pem", "key.pem" , proxyMux)
	if err != nil {
		log.Fatal(err)
	}
}
