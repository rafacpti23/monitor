#!/usr/bin/env bash
set -euo pipefail
echo "== Downloading latest agent binary =="
curl -sSL -o /tmp/p-mon-agent-new https://monitor.papi.api.br/install/p-mon-agent-linux-amd64
chmod +x /tmp/p-mon-agent-new

echo "== Stopping agent =="
sudo systemctl stop p-mon-agent

echo "== Installing new agent =="
sudo install -m 0755 /tmp/p-mon-agent-new /usr/local/bin/p-mon-agent

echo "== Starting agent =="
sudo systemctl start p-mon-agent
sleep 3
sudo systemctl is-active p-mon-agent

echo "== Agent config =="
cat /etc/p-mon/agent.json

echo "== DONE =="
