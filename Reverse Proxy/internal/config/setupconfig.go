package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

func SetupConfigurations() ProxyConfig {
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
		log.Printf("could not read config file %s: %v, using defaults", *configPath, err)
		return proxyConfig
	}

	if err := json.Unmarshal(fileData, &proxyConfig); err != nil {
		log.Printf("invalid config file %s: %v, using defaults", *configPath, err)
		return proxyConfig
	}

	return proxyConfig
}
