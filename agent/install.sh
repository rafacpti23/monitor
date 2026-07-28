#!/bin/bash
set -e
SERVER_KEY="$1"
BACKEND_URL="${BACKEND_URL:-https://p-mon.example.com}"

if [ -z "$SERVER_KEY" ]; then 
  echo "Usage: install.sh <SERVER_KEY>"
  exit 1
fi

# Detect arch
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) BIN="p-mon-agent-linux-amd64" ;;
  aarch64|arm64) BIN="p-mon-agent-linux-arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

# Download binary
echo "Downloading $BIN from $BACKEND_URL..."
sudo curl -sSL "$BACKEND_URL/download/$BIN" -o /usr/local/bin/p-mon-agent
sudo chmod +x /usr/local/bin/p-mon-agent

# Write config
echo "Creating /etc/p-mon-agent.json..."
sudo tee /etc/p-mon-agent.json >/dev/null <<EOF
{
  "backend_url": "$BACKEND_URL",
  "server_key": "$SERVER_KEY",
  "interval_seconds": 30,
  "collect": {
    "system": true,
    "docker": true,
    "pm2": true,
    "services": []
  }
}
EOF

# Systemd service
echo "Configuring systemd service..."
sudo tee /etc/systemd/system/p-mon-agent.service >/dev/null <<EOF
[Unit]
Description=P-mon Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/p-mon-agent -config /etc/p-mon-agent.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now p-mon-agent

echo "✅ P-mon Agent installed and running!"
