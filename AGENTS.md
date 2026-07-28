# P-mon — Lightweight Monitoring Platform

## Overview

P-mon is a lightweight, self-hosted monitoring platform for VPS, websites, Docker containers, and PM2 services.
Designed for a 1-vCPU AWS VPS. Single Go binary + SQLite + modern Vue SPA.

**Repository root:** `C:\Users\pichau\Documents\Rafael\p-mon`

---

## Architecture

```
p-mon/
├── backend/          # Go backend (single binary)
│   ├── cmd/server/   # Main entry point
│   ├── internal/     # Core logic
│   └── pkg/          # Shared packages
├── agent/            # VPS-side binary (Go)
│   ├── monitor/      # System metrics
│   ├── collector/     # Docker + PM2 collectors
│   └── cmd/          # Agent entry point
└── frontend/         # Vue 3 SPA
    ├── src/views/    # Pages
    └── src/components/
```

---

## Design Language

**Aesthetic:** Terminal-inspired dark dashboard — monospace typography, phosphor-green accents, grid-based layout.
Think "mission control" meets modern SaaS.

### Colors (CSS Variables)
```css
--bg-base: #0a0f14;
--bg-surface: #111820;
--bg-card: #1a2332;
--accent-green: #00e676;
--accent-green-dim: #00c853;
--accent-amber: #ffab00;
--accent-red: #ff5252;
--text-primary: #e8eaed;
--text-secondary: #8b9ab4;
--border: #2a3545;
```

### Typography
- **Display:** JetBrains Mono (monospace, for metrics/numbers)
- **Body:** IBM Plex Sans (clean, readable)
- **Scale:** 11px / 13px / 15px / 18px / 24px / 36px

---

## Core Features

### Backend (Go)
- [ ] SQLite database with Litestream (optional S3 backup)
- [ ] REST API (Gin framework)
- [ ] WebSocket for real-time dashboard
- [ ] JWT authentication with roles (admin, member, viewer)
- [ ] Agent receiver endpoint (POST /api/v1/agent/:key)
- [ ] Scheduler for website/checks (cron-like, every 30s)
- [ ] Alert engine: evaluate thresholds, send notifications
- [ ] Alert channels: Email (SMTP), WhatsApp (via Chat API / Papi-style)
- [ ] Tenant isolation (SAAS multi-user)

### Frontend (Vue 3)
- [ ] Dashboard (overview cards, incidents, uptime)
- [ ] Servers list + detail view
- [ ] Websites list + detail view
- [ ] Checks list (TCP, ping, DNS, SSL)
- [ ] Incidents / alerts log
- [ ] Settings: alert channels, user management
- [ ] Real-time updates via WebSocket

### Agent (Go binary for VPS)
- [ ] System metrics: CPU, RAM, Disk, Load, Net
- [ ] Auto-detection: Docker containers running
- [ ] PM2 process list (if pm2 is installed)
- [ ] Configurable: which metrics to report
- [ ] POST compressed data to backend every 30s
- [ ] One-line install command

---

## Agent Communication Protocol

**Endpoint:** `POST /api/v1/agent/{server_key}`

**Payload (gzipped JSON):**
```json
{
  "hostname": "web-01",
  "os": "linux",
  "arch": "amd64",
  "uptime_seconds": 86400,
  "cpu_percent": 23.5,
  "cpu_cores": 2,
  "cpu_model": "Intel Xeon",
  "load_avg": [0.12, 0.08, 0.05],
  "ram_total_bytes": 4294967296,
  "ram_used_bytes": 2147483648,
  "disks": [
    {"mount": "/", "total_bytes": 50000000000, "used_bytes": 25000000000}
  ],
  "net_rx_bytes": 1000000,
  "net_tx_bytes": 500000,
  "docker_containers": [
    {"name": "nginx", "status": "running", "image": "nginx:latest"}
  ],
  "pm2_processes": [
    {"name": "api", "status": "online", "cpu": 2.1, "memory": 104857600}
  ],
  "services": ["nginx", "mysql", "redis"],
  "agent_version": "0.1.0"
}
```

---

## Alert System

### Alert Types
- `nodata` — server hasn't reported in X minutes
- `cpu` — CPU usage >= threshold (%)
- `ram` — RAM usage >= threshold (%)
- `disk` — Disk usage >= threshold (%)
- `load` — Load average >= threshold
- `service_down` — named service/PM2 process not running
- `docker_down` — Docker container stopped
- `website_down` — HTTP check failed
- `ssl_expiry` — SSL cert expires in X days
- `ping_latency` — ping > threshold (ms)

### Alert Channels (configurable per user)
- **Email:** SMTP direct or via API
- **WhatsApp:** Via Chat API / Papi (user provides API key/URL)
- **Webhook:** POST to user URL

### Alert Rules
Each monitor has configurable alert rules:
```json
{
  "type": "cpu",
  "comparison": ">=",
  "threshold": 90,
  "occurrences": 3,
  "cooldown_minutes": 30,
  "channels": ["whatsapp", "email"]
}
```

---

## API Endpoints

### Auth
- `POST /api/v1/auth/register` — Create account
- `POST /api/v1/auth/login` — Login, returns JWT
- `POST /api/v1/auth/refresh` — Refresh token
- `GET /api/v1/auth/me` — Current user

### Servers
- `GET /api/v1/servers` — List servers
- `POST /api/v1/servers` — Create server
- `GET /api/v1/servers/:id` — Server detail + latest metrics
- `PUT /api/v1/servers/:id` — Update server
- `DELETE /api/v1/servers/:id` — Delete server
- `GET /api/v1/servers/:id/history` — Metrics history
- `GET /api/v1/servers/:id/incidents` — Incidents

### Websites
- `GET /api/v1/websites` — List websites
- `POST /api/v1/websites` — Create website
- `GET /api/v1/websites/:id` — Website detail
- `PUT /api/v1/websites/:id` — Update
- `DELETE /api/v1/websites/:id` — Delete

### Checks
- `GET /api/v1/checks` — List checks
- `POST /api/v1/checks` — Create check (ping/tcp/dns/ssl)
- `PUT /api/v1/checks/:id`
- `DELETE /api/v1/checks/:id`

### Incidents
- `GET /api/v1/incidents` — List incidents
- `PUT /api/v1/incidents/:id/acknowledge`
- `PUT /api/v1/incidents/:id/resolve`

### Settings
- `GET /api/v1/settings/channels` — Get alert channels
- `PUT /api/v1/settings/channels` — Update channels
- `POST /api/v1/settings/channels/test` — Test channel

### WebSocket
- `WS /ws?token=JWT` — Real-time events

---

## Database Schema (SQLite)

See `backend/internal/db/schema.sql`

---

## Team Coordination

This project uses three parallel agents:
1. **backend-agent** — Builds the Go backend (API, WebSocket, scheduler, alerts, auth)
2. **frontend-agent** — Builds the Vue 3 SPA (dashboard, pages, components)
3. **agent-bin-agent** — Builds the Go VPS agent (metrics collection, Docker/PM2)

Each agent reads this AGENTS.md and its own section below.

---

## Agent: backend-agent

**Start:** After frontend-agent-agent is initialized.
**Goal:** Complete Go backend in `C:\Users\pichau\Documents\Rafael\p-mon\backend`

Priority order:
1. `db/schema.sql` — SQLite schema
2. `pkg/models/*.go` — Data models
3. `internal/db/*.go` — Database layer
4. `internal/config/config.go` — Config loading
5. `internal/auth/jwt.go` — JWT middleware
6. `internal/api/servers.go` — Server CRUD + agent receiver
7. `internal/api/websites.go` — Website CRUD
8. `internal/api/checks.go` — Checks CRUD
9. `internal/api/incidents.go` — Incidents
10. `internal/api/settings.go` — Alert channels
11. `internal/ws/hub.go` — WebSocket hub
12. `internal/scheduler/scheduler.go` — Cron jobs
13. `internal/alerts/evaluator.go` — Threshold evaluation
14. `internal/alerts/notifier.go` — Email + WhatsApp sender
15. `cmd/server/main.go` — Entry point wiring everything

Use Gin as the HTTP framework. Use `github.com/mattn/go-sqlite3` for SQLite.
For WebSocket use `github.com/gorilla/websocket`.

---

## Agent: frontend-agent

**Start:** Immediately.
**Goal:** Complete Vue 3 SPA in `C:\Users\pichau\Documents\Rafael\p-mon\frontend`

Tech stack: Vue 3 + Vite + Pinia + Vue Router + Axios.
NO Tailwind by default — use scoped CSS with the design variables above.

Priority:
1. `vite.config.js` + `package.json`
2. `src/assets/main.css` — Design tokens, global styles
3. `src/main.js` — App bootstrap
4. `src/router/index.js` — Routes
5. `src/stores/auth.js` — Auth state
6. `src/stores/servers.js` — Server state
7. `src/composables/useApi.js` — Axios wrapper
8. `src/composables/useSocket.js` — WebSocket composable
9. `src/views/Login.vue` + `Register.vue`
10. `src/views/Dashboard.vue` — Overview
11. `src/views/Servers.vue` + `Servers/Detail.vue`
12. `src/views/Websites.vue` + `Websites/Detail.vue`
13. `src/views/Checks.vue`
14. `src/views/Incidents.vue`
15. `src/views/Settings.vue` — Alert channels
16. `src/views/AddServer.vue` — Shows the agent install command
17. `src/components/` — Reusable components

### Dashboard Design
- Dark terminal aesthetic
- Server list with live metrics (CPU/RAM mini-bars)
- Incident timeline on the right
- Top row: 4 stat cards (servers up, websites up, alerts, uptime %)
- Real-time updates via WebSocket

---

## Agent: agent-bin-agent

**Start:** After backend-agent starts schema design.
**Goal:** Build VPS-side agent binary in `C:\Users\pichau\Documents\Rafael\p-mon\agent`

Tech stack: Pure Go, no external dependencies except standard library + `github.com/shirou/gopsutil/v3`.

Priority:
1. `go.mod` with dependencies
2. `collector/system.go` — CPU, RAM, Disk, Load, Net
3. `collector/docker.go` — Docker container list via socket
4. `collector/pm2.go` — PM2 process list via `pm2 jlist`
5. `monitor/config.go` — Config (backend URL, server key, interval, what to collect)
6. `monitor/collector.go` — Aggregates all collectors
7. `cmd/agent/main.go` — Main loop: collect → gzip → POST → sleep
8. Build script: `build.sh` for Linux binary
9. `INSTALL.md` — One-line install command

Config file (`/etc/p-mon-agent.json` or `~/.p-mon-agent.json`):
```json
{
  "backend_url": "https://your-p-mon.com",
  "server_key": "KEY-FROM-WEB-UI",
  "interval_seconds": 30,
  "collect": {
    "system": true,
    "docker": true,
    "pm2": true,
    "services": ["nginx", "mysql"]
  }
}
```

Install command (one-liner):
```bash
curl -sSL https://your-p-mon.com/install/agent.sh | bash
```
