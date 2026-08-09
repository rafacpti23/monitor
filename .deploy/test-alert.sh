#!/bin/bash
set -e

# Login
echo "=== Login ==="
RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"rafacpti@gmail.com","password":"Ramel@2026"}')
echo "$RESP"

TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
echo "Token OK: ${TOKEN:0:20}..."

# Listar canais
echo ""
echo "=== Canais ==="
curl -s http://localhost:8080/api/v1/settings/channels \
  -H "Authorization: Bearer $TOKEN"

# Testar canal id=2 (WhatsApp)
echo ""
echo "=== Testando canal WhatsApp (id=2) ==="
curl -s -X POST http://localhost:8080/api/v1/settings/channels/2/test \
  -H "Authorization: Bearer $TOKEN"
