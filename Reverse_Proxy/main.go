package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"os/signal"
	"context"

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

    adminAddr := ":" + fmt.Sprint(cfg.AdminPort)
    adminServer := &http.Server{
        Addr:    adminAddr,
        Handler: adminMux,
    }

	// Proxy endpoint
	proxyAddr := ":" + fmt.Sprint(cfg.Port)
	proxyServer := &http.Server{
		Addr:         proxyAddr,        
		Handler:      proxy.ProxyHandler(mainPool),     
		ReadTimeout:  time.Duration(cfg.RequestTimeout) * time.Millisecond, 
		WriteTimeout: time.Duration(cfg.RequestTimeout) * time.Millisecond,
		IdleTimeout:  time.Duration(cfg.RequestTimeout) * 12 * time.Millisecond,
	}

	// Strat Proxy
	go func() {
		log.Printf("Starting proxy on %s", proxyAddr)
		err := proxyServer.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil {
			log.Fatal("Proxy server error: ", err)
		}
	} ()

	// Start Health Checker
	go healthcheck.StartHealthChecker(cfg.HealthCheckFreq, cfg.HealthCheckPath, mainPool)

	// Start the Admin API
	go func() {
		adminAddr := ":" + fmt.Sprint(cfg.AdminPort)
		log.Printf("Starting Admin API on %s", adminAddr)
		err := adminServer.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Handle Interrupt Signal
	graceful_exit_signal := make(chan os.Signal)
	signal.Notify(graceful_exit_signal, os.Interrupt)
	<-graceful_exit_signal

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

	// Shutdown Proxy
    if err := proxyServer.Shutdown(ctx); err != nil {
        log.Printf("Proxy forced to shutdown: %v", err)
    } else {
        log.Println("Proxy stopped gracefully")
    }

    // Shutdown Admin API
    if err := adminServer.Shutdown(ctx); err != nil {
        log.Printf("Admin API forced to shutdown: %v", err)
    } else {
        log.Println("Admin API stopped gracefully")
    }
}
