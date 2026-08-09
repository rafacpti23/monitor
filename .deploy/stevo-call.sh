#!/bin/bash
TOKEN=$(sqlite3 /var/lib/p-mon/p-mon.db "SELECT panel_token FROM papi_panels WHERE id=3;")

echo "=== Calling tools/call list_instances ==="
RESP=$(curl -s https://openapi.stevo.chat/mcp \
  --request POST \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/json, text/event-stream' \
  --header "Authorization: Bearer $TOKEN" \
  --data '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_instances","arguments":{}}}')

echo "$RESP"
