#!/bin/bash
set -e

echo "=== Deploying PAPI feature ==="

# 1. Copy schema for migration
sudo cp /tmp/schema.sql /opt/p-mon/internal/db/schema.sql
echo "Schema updated"

# 2. Stop server
sudo systemctl stop p-mon-server || true
echo "Server stopped"

# 3. Replace server binary
sudo cp /tmp/p-mon-server-linux /opt/p-mon/bin/p-mon-server
sudo chmod +x /opt/p-mon/bin/p-mon-server
echo "Server binary updated"

# 4. Deploy frontend
sudo rm -rf /var/www/p-mon/assets
sudo tar -xzf /tmp/dist.tar.gz -C /var/www/p-mon/
echo "Frontend updated"

# 5. Start server
sudo systemctl start p-mon-server
sleep 2
echo "Server started"

# 6. Health check
STATUS=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/health)
if [ "$STATUS" = "200" ]; then
    echo "✅ Health check OK (200)"
else
    echo "❌ Health check FAILED ($STATUS)"
    sudo journalctl -u p-mon-server --no-pager -n 20
    exit 1
fi

echo "=== Deploy complete ==="
