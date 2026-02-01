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
go run . --config=config.json
```

## Run dummy servers
```bash
go run ../Dummy_Servers/Server1/Server1.go
go run ../Dummy_Servers/Server2/Server2.go
go run ../Dummy_Servers/Server3/Server3.go
```

## Admin API
- `GET /status` (on `admin_port`)
- `POST /backends` with JSON `{ "url": "http://localhost:9004" }`
- `DELETE /backends` with JSON `{ "url": "http://localhost:9004" }`
