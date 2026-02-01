# Concurrent Load-Balancing Reverse Proxy

This project implements a concurrent reverse proxy with load balancing, health checks, and an admin API.

## Features
- Round-robin or least-connections load balancing
- Background health checks with configurable frequency and path
- Admin API to add/remove backends and inspect status
- Graceful shutdown and request timeouts

## Configuration
Edit `config.json`:
```json
{
  "port": 8080,
  "admin_port": 8081,
  "strategy": "round-robin",
  "health_check_frequency": 5000,
  "health_check_path": "/health",
  "request_timeout_ms": 10000,
  "backends": [
    "http://localhost:9001",
    "http://localhost:9002",
    "http://localhost:9003"
  ]
}
```

## Run the proxy
```bash
cd Reverse\ Proxy/
go run cmd/proxy/main.go --config=config.json
```

## Run dummy servers
```bash
go run Dummy_Servers/Server1/Server1.go
go run Dummy_Servers/Server2/Server2.go
go run Dummy_Servers/Server3/Server3.go
```

## Admin API
- `GET /status` (on `admin_port`)
- `POST /backends` with JSON `{ "url": "http://localhost:9004" }`
- `DELETE /backends` with JSON `{ "url": "http://localhost:9004" }`

## Test Extreme Loads
You can stress test the load balancer using hey, a modern HTTP load generator. 
A Windows executable (hey.exe) is included in the Reverse Proxy directory, or you can download it from the official repository : https://github.com/rakyll/hey

Run the following command in your terminal to send 200 requests with 100 concurrent workers:
```bash
.\hey.exe -n 200 -c 100 https://www.google.com/
```
