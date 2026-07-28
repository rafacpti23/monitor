#!/usr/bin/env bash
set -uo pipefail
echo "== systemd =="
sudo systemctl is-active p-mon-server
echo "== servers =="
sudo sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, name, status, interval_seconds, datetime(last_seen) FROM servers;"
echo "== health =="
curl -s -o /dev/null -w "https=%{http_code}\n" https://monitor.papi.api.br/health
echo "== recent nodata incidents =="
sudo sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, monitor_id, alert_type, severity, status, datetime(start_time), datetime(resolved_at) FROM incidents WHERE alert_type='nodata' ORDER BY id DESC LIMIT 5;"
