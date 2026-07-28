#!/bin/bash
set -e

echo "Uninstalling P-mon Agent..."

# Stop and disable the service if it exists
if systemctl list-unit-files 2>/dev/null | grep -Fq "p-mon-agent.service"; then
    echo "-> Stopping p-mon-agent service..."
    sudo systemctl stop p-mon-agent 2>/dev/null || true
    sudo systemctl disable p-mon-agent 2>/dev/null || true
    sudo rm -f /etc/systemd/system/p-mon-agent.service
    sudo systemctl daemon-reload
fi

# Remove binary
if [ -f "/usr/local/bin/p-mon-agent" ]; then
    echo "-> Removing agent binary..."
    sudo rm -f /usr/local/bin/p-mon-agent
fi

# Remove config directory (correct path used by install.sh)
if [ -d "/etc/p-mon" ]; then
    echo "-> Removing config /etc/p-mon..."
    sudo rm -rf /etc/p-mon
fi

# Also clean up legacy config path (older installations)
if [ -f "/etc/p-mon-agent.json" ]; then
    echo "-> Removing legacy config /etc/p-mon-agent.json..."
    sudo rm -f /etc/p-mon-agent.json
fi

# Remove log file
if [ -f "/var/log/p-mon-agent.log" ]; then
    echo "-> Removing log file..."
    sudo rm -f /var/log/p-mon-agent.log
fi

echo ""
echo "P-mon Agent uninstalled successfully."
