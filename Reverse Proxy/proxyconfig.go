package main

type ProxyConfig struct {
	Port                  int      `json:"port"`
	AdminPort             int      `json:"admin_port"`
	Strategy              string   `json:"strategy"` // "round-robin" or "least-conn"
	HealthCheckFreq		  int      `json:"health_check_frequency"`
	HealthCheckPath       string   `json:"health_check_path"`
	RequestTimeout 		  int      `json:"request_timeout_ms"`
	Backends              []string `json:"backends"`
}
