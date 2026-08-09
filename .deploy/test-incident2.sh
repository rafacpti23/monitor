#!/bin/bash
set -e

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"rafacpti@gmail.com","password":"Ramel@2026"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

echo "Token OK"

# Criar website fake
RESP=$(curl -s -X POST http://localhost:8080/api/v1/websites \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"TESTE-ALERTA-2","url":"http://site-que-nao-existe-999.invalid/","check_interval_sec":30}')
WID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
echo "Website id=$WID criado. Aguardando 40s..."
sleep 40

echo ""
echo "=== Incidentes recentes ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, monitor_type, monitor_id, alert_type, status FROM incidents ORDER BY id DESC LIMIT 5;"

echo ""
echo "=== notifications_sent ==="
sqlite3 /var/lib/p-mon/p-mon.db "SELECT ns.id, ns.incident_id, ns.channel_id, ns.success, ns.sent_at FROM notifications_sent ns ORDER BY ns.id DESC LIMIT 10;"

echo ""
echo "=== Limpando ==="
curl -s -X DELETE "http://localhost:8080/api/v1/websites/$WID" -H "Authorization: Bearer $TOKEN"
echo "Done."
