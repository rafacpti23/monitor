# P-mon 🖥️📡

> Monitoramento leve, self-hosted, para VPS, websites, Docker e PM2.
> Feito para rodar tranquilo num VPS de 1 vCPU.

**P-mon** é um sistema de monitoramento moderno inspirado no nmon, reescrito do zero em **Go + SQLite + Vue 3**. Um único binário no backend, um agente minúsculo nas suas VPS, e um dashboard em tempo real.

---

## ✨ Features

- 📊 **Servidores** — CPU, RAM, Disco, Load, Rede, Uptime via agente leve
- 🐳 **Docker** — monitora containers, alerta se algum cair
- ⚙️ **PM2** — monitora processos Node.js
- 🔧 **Serviços** — nginx, mysql, redis, etc. via systemctl
- 🌐 **Websites** — HTTP, tempo de resposta, string de busca, expiração de SSL
- ✅ **Checks** — ping, TCP, DNS, SSL expiry
- 🚨 **Alertas** — WhatsApp (PAPI/Chat API), Email (SMTP), Webhook
- 👥 **Multi-usuário (SAAS)** — isolamento por tenant, roles (admin/member/viewer)
- ⚡ **Tempo real** — dashboard atualiza via WebSocket, sem reload
- 🗺️ **Mapa** — servidores geolocalizados
- 📈 **Uptime** — histórico 24h / 7d / 30d / 12 meses

---

## 🏗️ Arquitetura

```
p-mon/
├── backend/     # Go — API REST + WebSocket + SQLite + scheduler + alertas
├── agent/       # Go — binário leve que roda nas VPS monitoradas
└── frontend/    # Vue 3 — dashboard "mission control" dark
```

**Stack:**
- Backend: Go 1.21 + Gin + SQLite (mattn/go-sqlite3) + gorilla/websocket
- Agent: Go + gopsutil (binário único, ~5-10 MB, < 20 MB RAM)
- Frontend: Vue 3 + Vite + Pinia (SVG charts, sem dependências pesadas)

---

## 🚀 Como rodar

### Backend
```bash
cd backend
cp config.example.json config.json   # edite o config
go mod download
go run ./cmd/server
# Servidor sobe em http://localhost:8080
```

### Frontend
```bash
cd frontend
npm install
npm run dev      # dev em http://localhost:5173
npm run build    # build de produção em dist/
```

### Agente (na VPS monitorada)
```bash
# Instalação em 1 linha (gerada pelo dashboard):
curl -sSL https://seu-p-mon.com/install/agent.sh | bash -s -- SEU_SERVER_KEY
```

---

## 📦 Deploy num VPS de 1 vCPU

1. Compile o backend: `go build -o p-mon ./cmd/server`
2. Build o frontend: `npm run build` → sirva o `dist/` (o backend pode servir estático)
3. Configure o `config.json` com seu domínio, SMTP e WhatsApp API
4. Rode como serviço systemd
5. SQLite fica num arquivo só — backup é copiar `p-mon.db`

Consumo esperado: **~50-100 MB RAM total** (vs. 500 MB+ do stack PHP).

---

## 🔔 Configurando Alertas

No dashboard → **Configurações → Canais de Alerta**:

- **WhatsApp:** cole a URL da API (PAPI, Chat API, etc.) + sua chave
- **Email:** configure SMTP (host, porta, user, senha)
- **Webhook:** URL para integrar com Slack, Discord, etc.

Cada monitor tem regras configuráveis (threshold, ocorrências, cooldown).

---

## 📄 Licença

Projeto privado — Rafael.
