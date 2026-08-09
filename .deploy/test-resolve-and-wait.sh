#!/bin/bash
set -e

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"rafacpti@gmail.com","password":"Ramel@2026"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

echo "Resolvendo incidente #18..."
curl -s -X PUT http://localhost:8080/api/v1/incidents/18/resolve \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "Aguardando 40s..."
sleep 40

echo ""
echo "=== Incidentes recentes ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, monitor_type, monitor_id, alert_type, status FROM incidents ORDER BY id DESC LIMIT 5;"

echo ""
echo "=== notifications_sent ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, incident_id, channel_id, success, sent_at FROM notifications_sent ORDER BY id DESC LIMIT 10;"
