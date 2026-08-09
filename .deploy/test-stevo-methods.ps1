$ErrorActionPreference = 'Stop'
$key = "C:\Users\pichau\Documents\Rafael\p-mon\.deploy\vps-key.pem"
$base = "https://monitor.papi.api.br"

# Login
$loginResp = Invoke-RestMethod -Uri "$base/api/v1/auth/login" -Method POST -ContentType "application/json" -Body '{"email":"rafacpti@gmail.com","password":"Ramel@2026"}'
$token = $loginResp.token
$headers = @{ Authorization = "Bearer $token" }

# Get the stevo token from the panel config
$panel = Invoke-RestMethod -Uri "$base/api/v1/papi/panels/3" -Method GET -Headers $headers
Write-Host "Panel base_url: $($panel.base_url)"

# Try calling the Stevo API directly from the VPS
$sshKey = "C:\Users\pichau\Documents\Rafael\p-mon\.deploy\vps-key.pem"
