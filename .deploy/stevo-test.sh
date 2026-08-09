#!/bin/bash
TOKEN=$(sqlite3 /var/lib/p-mon/p-mon.db "SELECT panel_token FROM papi_panels WHERE id=3;")
echo "Token: ${TOKEN:0:10}..."

# Try different MCP method names
for method in "list_instances" "instances/list" "tools/list" "resources/list" "initialize" "tools/call"; do
  echo "--- Trying method: $method ---"
  RESP=$(curl -s -w "\nHTTP_CODE:%{http_code}" https://openapi.stevo.chat/mcp \
    --request POST \
    --header 'Content-Type: application/json' \
    --header 'Accept: application/json, text/event-stream' \
    --header "Authorization: Bearer $TOKEN" \
    --data "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"$method\"}")
  echo "$RESP" | tail -5
  echo ""
done
