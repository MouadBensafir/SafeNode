package main

import (
	"flag"
	"log"
	"os"
	"encoding/json"
	"net/http"	

)

func handleRequests(w http.ResponseWriter, r *http.Request) {
	backnd := mainPool.GetNextValidPeer()
	if backnd == nil {
		log.Fatal("ERROR: 503 Service Unavailable")
	}
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