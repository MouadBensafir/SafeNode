package shutdown

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"context"
	"time"
	"net/http"
)

func HandleGracefulShutdown(proxyServer, adminServer *http.Server) {
	// Handle Interrupt Signal
	graceful_exit_signal := make(chan os.Signal, 1)
	signal.Notify(graceful_exit_signal, os.Interrupt)
	<-graceful_exit_signal

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

	// Shutdown Proxy
	err := proxyServer.Shutdown(ctx);
	fmt.Println()
    if err != nil {
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