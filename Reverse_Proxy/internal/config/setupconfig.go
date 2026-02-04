package config

import (
	"encoding/json"
	"net/url"
	"flag"
	"log"
	"os"

	"github.com/MouadBensafir/SafeNode/internal/pool"
	"github.com/MouadBensafir/SafeNode/internal/proxy"
)

func SetupConfigurations(mainPool *pool.ServerPool) ProxyConfig {
	// Read the flag when running, if provided ofc
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	flag.Parse()
	if *configPath == "" {
		*configPath = "config.json" // Defaut configuration
	}

	// Parse the configurations passed
	var proxyConfig ProxyConfig
	fileData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Printf("Could not read config file %s: %v, using defaults", *configPath, err)
		return proxyConfig
	}

	if err := json.Unmarshal(fileData, &proxyConfig); err != nil {
		log.Printf("Invalid config file %s: %v, using defaults", *configPath, err)
		return proxyConfig
	}

	// Strategy Setup
	switch proxyConfig.Strategy {
		case "round-robin":
			mainPool.Strategy = &pool.RoundRobinStrategy{}
		case "least-connections":
			mainPool.Strategy = &pool.LeastConnStrategy{}
		default:
			mainPool.Strategy = &pool.RoundRobinStrategy{}
	}

	// Automatically Add backends from config.json
	for _, strURL := range proxyConfig.Backends {
		u, err := url.Parse(strURL)
		if err != nil {
			log.Printf("Skipping invalid URL %s: %v", strURL, err)
			continue
		}
		mainPool.AddBackend(proxy.NewBackend(u, mainPool))
	}

	return proxyConfig
}
