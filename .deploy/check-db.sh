#!/bin/bash
echo "=== Incidentes ativos ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, monitor_type, monitor_id, alert_type, status, message FROM incidents WHERE status IN ('active','acknowledged') ORDER BY id;"

echo ""
echo "=== Todos websites no DB ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, name, url, status FROM websites ORDER BY id;"
