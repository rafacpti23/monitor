package agent

import (
	"bytes"
	"strings"
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
cat <<'PMON_EOF' > /etc/p-mon/agent.json
{{.ConfigJSON}}
PMON_EOF

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

// ScriptData holds all variables for the install script template.
type ScriptData struct {
	BackendURL string
	ServerKey  string
	ConfigJSON string
}

// CollectOpts represents the user's monitoring choices.
type CollectOpts struct {
	System   bool
	Docker   bool
	PM2      bool
	Services []string
	Interval int
}

// buildConfigJSON generates the agent config JSON from user choices.
func buildConfigJSON(backendURL, serverKey string, opts CollectOpts) string {
	if opts.Interval <= 0 {
		opts.Interval = 180
	}

	// Build services array string
	svcJSON := "[]"
	if len(opts.Services) > 0 {
		parts := make([]string, len(opts.Services))
		for i, s := range opts.Services {
			parts[i] = `"` + s + `"`
		}
		svcJSON = "[" + strings.Join(parts, ", ") + "]"
	}

	var b strings.Builder
	b.WriteString("{\n")
	b.WriteString(`  "backend_url": "` + backendURL + "\",\n")
	b.WriteString(`  "server_key": "` + serverKey + "\",\n")
	b.WriteString(`  "interval_seconds": ` + itoa(opts.Interval) + ",\n")
	b.WriteString("  \"collect\": {\n")
	b.WriteString(`    "system": ` + boolStr(opts.System) + ",\n")
	b.WriteString(`    "docker": ` + boolStr(opts.Docker) + ",\n")
	b.WriteString(`    "pm2": ` + boolStr(opts.PM2) + ",\n")
	b.WriteString(`    "services": ` + svcJSON + "\n")
	b.WriteString("  }\n")
	b.WriteString("}")
	return b.String()
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Render produces the install script for a given backend URL, server key, and collect options.
func Render(backendURL, serverKey string, opts CollectOpts) (string, error) {
	tmpl, err := template.New("install").Parse(installScriptTemplate)
	if err != nil {
		return "", err
	}

	data := ScriptData{
		BackendURL: backendURL,
		ServerKey:  serverKey,
		ConfigJSON: buildConfigJSON(backendURL, serverKey, opts),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
