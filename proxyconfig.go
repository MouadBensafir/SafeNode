package main

import (
	"time"
)

type ProxyConfig struct { 
	Port int `json:"port"`          
	Strategy string `json:"strategy"` // can be "round-robin" or "least-conn" 
	HealthCheckFreq time.Duration `json:"health_check_frequency"` 
} 