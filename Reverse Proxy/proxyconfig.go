package main

type ProxyConfig struct {
	Port     int    `json:"port"`
	Strategy string `json:"strategy"` // can be "round-robin" or "least-conn"
	HealthCheckFreq int `json:"health_check_frequency"`
}
