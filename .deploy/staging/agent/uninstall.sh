#!/bin/bash

# Stop and disable the service if it exists
if systemctl list-units --full --all | grep -Fq "p-mon-agent.service"; then
    echo "Stopping and disabling p-mon-agent service..."
    sudo systemctl stop p-mon-agent || true
    sudo systemctl disable p-mon-agent || true
    sudo rm -f /etc/systemd/system/p-mon-agent.service
    sudo systemctl daemon-reload
fi

# Remove binary
if [ -f "/usr/local/bin/p-mon-agent" ]; then
    echo "Removing agent binary..."
    sudo rm -f /usr/local/bin/p-mon-agent
fi

# Remove config
if [ -f "/etc/p-mon-agent.json" ]; then
    echo "Removing config file..."
    sudo rm -f /etc/p-mon-agent.json
fi

echo "✅ P-mon Agent uninstalled successfully."
