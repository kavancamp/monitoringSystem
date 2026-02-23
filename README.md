# monitoring_system






















## Run locally

### 1. Start database
```bash
docker compose up -d
```
### 2. Set env
```bash
cp.env.example .env
```
### 3. Run migrations
```bash 
export DATABASE_URL=postgres://scada:scada@localhost:5433/scada?sslmode=disable
goose -dir migrations postgres "$DATABASE_URL" up
```

### 4. Start API
```bash
go run ./cmd/scada
```
### 5. Health Check
```bash
curl http://localhost:8080/healthz
```
