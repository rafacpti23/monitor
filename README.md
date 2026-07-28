<p align="center">
  <img src="./docs/logo-full.png" alt="P-mon" width="280" />
</p>

# P-mon 🖥️📡

> Monitoramento leve, self-hosted, para VPS, websites, Docker e PM2.
> Feito para rodar tranquilo num VPS de 1 vCPU.

**P-mon** é um sistema de monitoramento moderno inspirado no nmon, reescrito do zero em **Go + SQLite + Vue 3**. Um único binário no backend, um agente minúsculo nas suas VPS, e um dashboard em tempo real.

🇧🇷 **PT-BR** · [🇺🇸 English](./README.en.md)

---

## ✨ Features

- 📊 **Servidores** — CPU, RAM, Disco, Load, Rede, Uptime via agente leve
- 🐳 **Docker** — monitora containers, alerta se algum cair
- ⚙️ **PM2** — monitora processos Node.js
- 🔧 **Serviços** — nginx, mysql, redis, etc. via systemctl
- 🌐 **Websites** — HTTP, tempo de resposta, string de busca, expiração de SSL
- ✅ **Checks** — ping, TCP, DNS, SSL expiry
- 🚨 **Alertas** — WhatsApp (PAPI), Email (SMTP), Webhook
- 👥 **Multi-usuário (SAAS)** — isolamento por tenant, roles (admin/member/viewer)
- ⚡ **Tempo real** — dashboard atualiza via WebSocket, sem reload
- 🎨 **Whitelabel** — logo e cores por empresa
- 📈 **Uptime** — histórico com retenção configurável

---

## 🏗️ Arquitetura

```
p-mon/
├── backend/     # Go — API REST + WebSocket + SQLite + scheduler + alertas
├── agent/       # Go — binário leve que roda nas VPS monitoradas
└── frontend/    # Vue 3 — dashboard "mission control" dark
```

**Stack:**
- Backend: Go 1.21 + Gin + SQLite (modernc.org/sqlite, puro Go) + gorilla/websocket
- Agent: Go + gopsutil (binário único, ~5-10 MB, ~5-7 MB RAM em produção)
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
# Instalação em 1 linha (gerada pelo dashboard, com flags de coleta):
curl -sSL "https://seu-p-mon.com/install/agent.sh?key=SEU_SERVER_KEY&docker=1&pm2=1&interval=180" | sudo bash
```

---

## 📦 Deploy num VPS de 1 vCPU

1. Compile o backend: `go build -o p-mon-server ./cmd/server`
2. Build do frontend: `npm run build` → `frontend/dist/`
3. Configure o `config.json` (JWT, SMTP, PAPI WhatsApp, base URL)
4. Rode como serviço systemd + Caddy na frente (HTTPS automático via Let's Encrypt)
5. SQLite é um arquivo só — backup é copiar `p-mon.db`

**Consumo real em produção:** ~5-7 MB RAM (backend) + ~5-7 MB (agente por host).

---

## 🔔 Configurando Alertas

No dashboard → **Configurações → Canais de Alerta**:

- **WhatsApp (PAPI):** cole a URL da API PAPI + sua chave
- **Email:** configure SMTP (host, porta, user, senha)
- **Webhook:** URL para integrar com Slack, Discord, etc.

Cada monitor tem regras configuráveis (threshold, ocorrências, cooldown).
Alerta de servidor offline dispara automaticamente em `2× intervalo do agente`.

---

## 📱 PAPI WhatsApp API

Para o canal de alertas via WhatsApp, o P-mon usa a **PAPI** — API oficial do WhatsApp Business (compatível com Cloud API e não-oficial).

**Crie sua instância PAPI aqui:** 👉 **[https://papi.api.br](https://papi.api.br)**

---

## 📞 Contato

**Rafa Martins** — 📱 **+55 27 99908-2624**

---

## 📄 Licença

**P-mon License** — source-available, uso próprio livre, venda/SaaS mediante autorização.

- ✅ Você **pode**: baixar, instalar, usar internamente (pessoal ou na sua empresa), estudar e modificar o código para uso próprio, fazer backup.
- 🔒 Você **precisa de autorização escrita** para: vender, revender, sublicenciar, oferecer como SaaS/serviço a terceiros, redistribuir versões modificadas ou remover as marcas.

Para licença comercial: **Rafa Martins — +55 27 99908-2624**.
Veja [LICENSE](./LICENSE) para os termos completos.
