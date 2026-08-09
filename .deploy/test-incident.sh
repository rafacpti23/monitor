#!/bin/bash
set -e

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"rafacpti@gmail.com","password":"Ramel@2026"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

echo "Token OK"

# Criar website fake que vai falhar
RESP=$(curl -s -X POST http://localhost:8080/api/v1/websites \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"TESTE-ALERTA","url":"http://site-que-nao-existe-123456789.invalid/","check_interval_sec":30}')
echo "Website criado: $RESP"

WID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
echo "ID=$WID — aguardando 40s para o scheduler rodar..."
sleep 40

echo ""
echo "=== Incidentes recentes ==="
curl -s http://localhost:8080/api/v1/incidents \
  -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
data = json.load(sys.stdin)
items = data if isinstance(data, list) else data.get('incidents', data.get('data', []))
for i in items[:5]:
    print(i.get('id'), i.get('monitor_type'), i.get('alert_type'), i.get('status'), i.get('message','')[:60])
"

echo ""
echo "=== Limpando website de teste ==="
curl -s -X DELETE "http://localhost:8080/api/v1/websites/$WID" \
  -H "Authorization: Bearer $TOKEN"
echo "Limpo."
