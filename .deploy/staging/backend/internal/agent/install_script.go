package agent

import (
	"bytes"
	"text/template"
)

const installScriptTemplate = `#!/bin/bash
set -e

echo "Installing P-mon Agent..."

BACKEND_URL="{{.BackendURL}}"
SERVER_KEY="{{.ServerKey}}"

# Download binary
curl -sSL "${BACKEND_URL}/install/p-mon-agent-linux-amd64" -o /usr/local/bin/p-mon-agent
chmod +x /usr/local/bin/p-mon-agent

# Write config
mkdir -p /etc/p-mon
cat <<EOF > /etc/p-mon/agent.json
{
  "backend_url": "${BACKEND_URL}",
  "server_key": "${SERVER_KEY}",
  "interval_seconds": 30,
  "collect": {
    "system": true,
    "docker": true,
    "pm2": true
  }
}
EOF

# Write systemd unit
cat <<EOF > /etc/systemd/system/p-mon-agent.service
[Unit]
Description=P-mon Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/p-mon-agent -config /etc/p-mon/agent.json
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable p-mon-agent
systemctl restart p-mon-agent

echo "P-mon agent installed and running."
`

type ScriptData struct {
	BackendURL string
	ServerKey  string
}

// Render produces the install script for a given backend URL and server key.
func Render(backendURL, serverKey string) (string, error) {
	tmpl, err := template.New("install").Parse(installScriptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ScriptData{BackendURL: backendURL, ServerKey: serverKey}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
