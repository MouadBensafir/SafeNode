package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	proxyAddr := ":" + fmt.Sprint(cfg.Port)
	proxyServer := &http.Server{
		Addr:         proxyAddr,        
		Handler:      proxy.ProxyHandler(mainPool),     
		ReadTimeout:  time.Duration(cfg.RequestTimeout) * time.Millisecond, 
		WriteTimeout: time.Duration(cfg.RequestTimeout) * time.Millisecond,
		IdleTimeout:  time.Duration(cfg.RequestTimeout) * 12 * time.Millisecond,
	}
	err := proxyServer.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}

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

}
