#!/usr/bin/env bash
set -euo pipefail

echo "== Stopping p-mon-server =="
sudo systemctl stop p-mon-server

echo "== Installing binaries =="
sudo install -m 0755 /tmp/p-mon-server-linux  /opt/p-mon/bin/p-mon-server
sudo install -m 0755 /tmp/p-mon-migrate-linux /opt/p-mon/bin/p-mon-migrate
sudo install -m 0755 /tmp/p-mon-agent-linux   /opt/p-mon/bin/p-mon-agent
# also refresh the agent binary served by Caddy
sudo install -m 0755 /tmp/p-mon-agent-linux   /var/www/p-mon/install/p-mon-agent-linux-amd64

echo "== Updating schema.sql for migrate =="
sudo mkdir -p /opt/p-mon/internal/db
sudo cp /tmp/schema.sql /opt/p-mon/internal/db/schema.sql

echo "== Running migration (adds servers.interval_seconds) =="
cd /opt/p-mon
sudo /opt/p-mon/bin/p-mon-migrate || true

echo "== Verifying servers table has interval_seconds =="
sudo sqlite3 /var/lib/p-mon/p-mon.db "PRAGMA table_info(servers);" | grep interval_seconds || echo "!! interval_seconds column NOT found"

echo "== Deploying frontend =="
sudo rm -rf /var/www/p-mon/assets
sudo tar -xzf /tmp/dist.tar.gz -C /var/www/p-mon
sudo chown -R www-data:www-data /var/www/p-mon || true

echo "== Starting p-mon-server =="
sudo systemctl start p-mon-server
sleep 2
sudo systemctl is-active p-mon-server

echo "== Recent server logs =="
sudo journalctl -u p-mon-server -n 20 --no-pager

echo "== DONE =="
