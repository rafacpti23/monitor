#!/bin/bash
set -e

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"rafacpti@gmail.com","password":"Ramel@2026"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

echo "Token OK"

# Resolver incidente #11 manualmente
echo "=== Resolvendo incidente #11 ==="
curl -s -X PUT http://localhost:8080/api/v1/incidents/11/resolve \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "=== Estado dos incidentes ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, monitor_type, monitor_id, alert_type, status FROM incidents ORDER BY id DESC LIMIT 5;"

echo ""
echo "Aguardando 40s para o scheduler checar o site down e criar novo incidente..."
sleep 40

echo ""
echo "=== Incidentes após check ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, monitor_type, monitor_id, alert_type, status FROM incidents ORDER BY id DESC LIMIT 5;"

echo ""
echo "=== notifications_sent ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT ns.id, ns.incident_id, ns.channel_id, ns.success, ns.sent_at FROM notifications_sent ns ORDER BY ns.id DESC LIMIT 10;"
