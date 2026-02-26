# IN PROGRESS
# 🛰️ SCADA Monitoring System (Go)

A lightweight SCADA-style backend service built in Go for monitoring industrial devices, ingesting telemetry, and generating alerts.
This project simulates how real-world control systems track equipment (like wind turbines, pipelines, or manufacturing systems) using modern backend practices.

## Features

- REST API built with Go (net/http)
- PostgreSQL database with Docker
- Type-safe queries using SQLC
- Device management (create + list + filter)
- Pagination support for large datasets
- Clean project structure (internal packages)

## Tech Stack
- Go (Golang) – backend service
- PostgreSQL – relational database
- Docker Compose – local dev environment
- SQLC – type-safe SQL → Go code
- pgx – PostgreSQL driver


## 📂 Project Structure
```text
monitoringSystem/
├── cmd/
│   └── scada/              # application entrypoint
├── internal/
│   ├── api/                # HTTP handlers
│   └── database/           # DB connection + SQLC
├── sql/
│   └── queries/            # SQLC query definitions
├── migrations/             # database schema
├── docker-compose.yml      # postgres setup
├── sqlc.yaml               # SQLC config
└── README.md
```

## Run locally

### 1. Start database
```bash
docker compose up -d
```
### 2. Verify its running 
```bash
docker compose ps
```
### 3. Set Env variable
```bash 
export DATABASE_URL="postgres://scada:scada@localhost:5433/scada?sslmode=disable"
```
### 4. Run the server
```bash
go run ./cmd/scada
```
### 5. Health Check
```bash
curl http://localhost:8080/healthz
```


Optional query params:
- site
- status
- limit (default 50, max 200)
- offset

Examples:
```bash
# all devices
curl http://localhost:8080/devices

# filter by site
curl "http://localhost:8080/devices?site=DemoSite"

# pagination
curl "http://localhost:8080/devices?limit=2&offset=0"
```

#### Development Notes
PostgreSQL runs on port 5433 (mapped from container 5432)
SQLC generates code in internal/database/db
Use sqlc generate after updating queries

### Why This Project
This project is designed to mirror real-world backend systems used in:
- SCADA / industrial monitoring
- IoT platforms
- Energy systems (wind, pipelines, utilities)
- Distributed telemetry systems
