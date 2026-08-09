#!/bin/bash
TOKEN=$(sqlite3 /var/lib/p-mon/p-mon.db "SELECT panel_token FROM papi_panels WHERE id=3;")

# Get full tools list and extract just names
RESP=$(curl -s https://openapi.stevo.chat/mcp \
  --request POST \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/json, text/event-stream' \
  --header "Authorization: Bearer $TOKEN" \
  --data '{"jsonrpc":"2.0","id":1,"method":"tools/list"}')

# Extract data line
DATA=$(echo "$RESP" | grep '^data:' | sed 's/^data: //')

# Parse tool names
echo "$DATA" | python3 -c "
import sys, json
raw = sys.stdin.read().strip()
obj = json.loads(raw)
tools = obj.get('result', {}).get('tools', [])
for t in tools:
    print(f\"  {t['name']}: {t.get('description', '')[:80]}\")
"
