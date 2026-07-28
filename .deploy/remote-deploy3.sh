#!/usr/bin/env bash
set -euo pipefail
echo "== Stop server =="
sudo systemctl stop p-mon-server
echo "== Install binary =="
sudo install -m 0755 /tmp/p-mon-server-linux /opt/p-mon/bin/p-mon-server
echo "== Deploy frontend =="
sudo rm -rf /var/www/p-mon/assets
sudo tar -xzf /tmp/dist.tar.gz -C /var/www/p-mon
echo "== Start server =="
sudo systemctl start p-mon-server
sleep 2
sudo systemctl is-active p-mon-server
echo "== Health =="
curl -s -o /dev/null -w "https=%{http_code}\n" https://monitor.papi.api.br/health
echo "== WS endpoint reachable (expect 401 without token) =="
curl -s -o /dev/null -w "ws_noauth=%{http_code}\n" https://monitor.papi.api.br/ws
echo "== DONE =="
