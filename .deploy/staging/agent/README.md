# P-mon Agent

The P-mon Agent is a lightweight cross-platform Go binary that collects system metrics, Docker containers, and PM2 processes, and reports them securely to the P-mon backend.

## Design Goals
- **Minimal Footprint:** Single binary (~5-10MB). Uses less than 20MB of RAM.
- **Cross-Platform:** Runs flawlessly on Linux and Windows.
- **Auto-Discovery:** Automatically detects and monitors running Docker containers and PM2 applications (no explicit configuration needed).

## Installation

### Linux (One-Line Script)
If you already have your server key from the backend dashboard:

```bash
curl -sSL https://your-p-mon-backend.com/install/agent.sh | bash -s YOUR_SERVER_KEY
```

This script will automatically detect your architecture, download the correct binary, create a default configuration, and register a `systemd` service to run the agent in the background.

### Windows (Manual Setup)
1. Download the latest `p-mon-agent-windows-amd64.exe` from your backend release or UI.
2. Put it in `C:\p-mon\` or another persistent directory.
3. Create a config file named `.p-mon-agent.json` in your User Profile folder (e.g., `C:\Users\Administrator\.p-mon-agent.json`).
4. Run the executable in the background (or install as a service using tools like [NSSM](https://nssm.cc/)).

## Configuration

The agent loads configuration in the following order:
1. Command-line `-config` argument path.
2. `/etc/p-mon-agent.json`
3. `~/.p-mon-agent.json`

### `config.example.json`

```json
{
  "backend_url": "https://api.p-mon.example.com",
  "server_key": "abc123def456xyz987",
  "interval_seconds": 30,
  "collect": {
    "system": true,
    "docker": true,
    "pm2": true,
    "services": [
      "nginx",
      "mysql"
    ]
  },
  "log_file": "/var/log/p-mon-agent.log"
}
```

- **`backend_url`**: Your self-hosted or managed P-mon instance URL.
- **`server_key`**: Secret token providing access for this explicit server.
- **`interval_seconds`**: Loop cycle frequency (defaults to 30 seconds).
- **`collect.system`**: Enable/disable base server node metrics (CPU, Memory, Disk, etc).
- **`collect.docker`**: Auto-detect docker daemon and send container states. (Linux only for now).
- **`collect.pm2`**: Auto-detect `pm2` CLI on PATH and send process states.
- **`collect.services`**: Explicit services (like Systemd units) to track via `systemctl is-active`.

## Compiling from Source

Ensure you have Go 1.21+ installed on your system. 

```bash
# Clone and enter directory
cd p-mon/agent

# Fetch dependencies
go mod tidy

# Build for all main platforms (Linux amd64/arm64 + Windows amd64)
bash build.sh
# or if on Windows:
# .\build.ps1
```

The resulting binaries will be placed in the `dist/` folder.

## Clean Uninstall (Linux)

You can purge the service, binary, and configs cleanly utilizing the `uninstall.sh` script:

```bash
sudo bash uninstall.sh
```

---

# (Português) P-mon Agent

O P-mon Agent é um binário Go leve e multiplataforma que coleta métricas do sistema, containers Docker e processos do PM2, enviando as informações com segurança ao backend do P-mon.

## Objetivos do Design
- **Leveza:** Binário único (~5-10MB). Consumo de memória abaixo de 20MB.
- **Multiplataforma:** Roda tanto em Linux quanto em Windows sem dependências.
- **Auto-Discovery:** Detecta e monitora containers Docker e aplicações PM2 automaticamente (sem configuração complexa).

## Instalação

### Linux (Instalação Rápida em 1 Linha)
Se você já possui a `server_key` fornecida pelo seu dashboard Web:

```bash
curl -sSL https://seu-p-mon-backend.com/install/agent.sh | bash -s SUA_SERVER_KEY
```

O script detecta automaticamente a arquitetura, faz o download do binário correto, cria uma configuração base e registra um serviço no `systemd` para o agente trabalhar em background.

### Windows
1. Baixe o executável `p-mon-agent-windows-amd64.exe` pela dashboard ou página de releases.
2. Crie uma pasta, exemplo: `C:\p-mon\`.
3. Adicione o arquivo de configuração `.p-mon-agent.json` na pasta do usuário raiz (ex. `C:\Users\Administrator\.p-mon-agent.json`).
4. Rode como um serviço usando programas como o [NSSM](https://nssm.cc/).

## Configuração Básica

O agente irá buscar o arquivo de configuração seguindo a prioridade:
1. Argumento livre `--config` durante a execução
2. `/etc/p-mon-agent.json`
3. `~/.p-mon-agent.json`

(Consulte a aba de documentação acima em Inglês para um exemplo em arquivo `.json`).

## Compilando pelo Código

Requer Go 1.21+ em sua máquina.

```bash
cd p-mon/agent
go mod tidy

# Script de compilação multi-arquitetura (Gera pasta dist/)
bash build.sh 
# ou
# .\build.ps1
```

## Desinstalação (Linux)

Para revogar o serviço, binário e configurações permanentemente:

```bash
sudo bash uninstall.sh
```
