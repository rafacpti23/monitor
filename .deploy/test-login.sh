#!/bin/bash
set -e

# Login - ver resposta completa
echo "=== Login ==="
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"rafacpti@gmail.com","password":"123456"}'
