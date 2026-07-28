#!/usr/bin/env bash
set -euo pipefail

echo "== Stopping p-mon-server =="
sudo systemctl stop p-mon-server

echo "== Installing server binary =="
sudo install -m 0755 /tmp/p-mon-server-linux /opt/p-mon/bin/p-mon-server

echo "== Deploying frontend =="
sudo rm -rf /var/www/p-mon/assets
sudo tar -xzf /tmp/dist.tar.gz -C /var/www/p-mon

echo "== Clean stale Docker services for Ramel (server_id=1) =="
sudo sqlite3 /var/lib/p-mon/p-mon.db "DELETE FROM server_services WHERE server_id = 1 AND name LIKE 'docker:%';"
echo "Deleted docker entries"

echo "== Starting p-mon-server =="
sudo systemctl start p-mon-server
sleep 2
sudo systemctl is-active p-mon-server

echo "== Verify server_services for Ramel =="
sudo sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, server_id, name, status FROM server_services WHERE server_id = 1;"

echo "== DONE =="
