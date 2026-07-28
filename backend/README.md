# P-mon Backend

Go backend for P-mon — lightweight server/website/check monitoring with alerts.

## Prerequisites

- **Go 1.21+**
- **GCC / C compiler** — required by `mattn/go-sqlite3` (CGO).
  - **Linux/macOS:** `gcc` is usually pre-installed.
  - **Windows:** Install [MSYS2/MinGW-w64](https://www.msys2.org/) and ensure `gcc` is on PATH.

## Quick Start

```bash
# 1. Copy and edit config
cp config.example.json config.json

# 2. Build
CGO_ENABLED=1 go build -o p-mon-server ./cmd/server

# 3. Run (migrations run automatically on startup)
./p-mon-server
```

The server listens on `0.0.0.0:8080` by default.

## Standalone Migration

```bash
CGO_ENABLED=1 go run cmd/migrate/main.go
```

## Project Structure

```
cmd/server/        Entry point — wires everything
cmd/migrate/       Standalone migration tool
internal/
  agent/           Install script template
  alerts/          Evaluator + notifier (email, whatsapp, webhook)
  api/             Gin HTTP handlers (auth, servers, websites, checks, incidents, channels, rules, settings)
  auth/            JWT + bcrypt
  checks/          Ping, TCP, DNS, HTTP, SSL expiry checkers
  config/          Config loader
  db/              SQLite init + schema.sql
  scheduler/       Background cron (30s) for website/check/nodata monitoring
  ws/              WebSocket hub (real-time push per user)
pkg/models/        Shared Go structs
```

## API Overview

| Group | Auth | Description |
|-------|------|-------------|
| `POST /api/v1/auth/register` | Public | Create account |
| `POST /api/v1/auth/login` | Public | Login, returns JWT |
| `GET /api/v1/auth/me` | JWT | Current user |
| `POST /api/v1/auth/logout` | JWT | Invalidate session |
| `/api/v1/servers` | JWT | CRUD + metrics history |
| `POST /api/v1/agent/:key` | Public (key) | Agent metric receiver (gzip JSON) |
| `/api/v1/websites` | JWT | CRUD + history |
| `/api/v1/checks` | JWT | CRUD (ping/tcp/dns/ssl) |
| `/api/v1/incidents` | JWT | List, acknowledge, resolve, ignore |
| `/api/v1/alert-rules` | JWT | CRUD |
| `/api/v1/settings/channels` | JWT | CRUD + test |
| `/api/v1/settings` | JWT | User preferences |
| `GET /ws?token=JWT` | JWT (query) | WebSocket real-time events |
| `GET /install/agent.sh?key=KEY` | Public | Agent install script |
| `GET /health` | Public | Health check |
