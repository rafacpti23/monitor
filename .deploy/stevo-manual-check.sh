#!/bin/bash
TOKEN=$(curl -s http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"rafacpti@gmail.com","password":"Ramel@2026"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')

echo "Triggering check..."
curl -s -X POST http://127.0.0.1:8080/api/v1/papi/panels/3/check \
  -H "Authorization: Bearer $TOKEN"
echo ""

sleep 3

echo ""
echo "Panel status:"
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, name, status, last_checked, last_error, total_instances, connected_instances FROM papi_panels WHERE id=3;"

echo ""
echo "Instances:"
sqlite3 /var/lib/p-mon/p-mon.db "SELECT id, instance_id, name, phone_number, status FROM papi_instances WHERE panel_id=3;"
