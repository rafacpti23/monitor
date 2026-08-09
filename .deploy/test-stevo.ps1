$ErrorActionPreference = 'Stop'
$key = "C:\Users\pichau\Documents\Rafael\p-mon\.deploy\vps-key.pem"
$base = "https://monitor.papi.api.br"

# Login
$loginResp = Invoke-RestMethod -Uri "$base/api/v1/auth/login" -Method POST -ContentType "application/json" -Body '{"email":"rafacpti@gmail.com","password":"Ramel@2026"}'
$token = $loginResp.token
$headers = @{ Authorization = "Bearer $token" }

# Trigger manual check on Stevo panel (id=3)
$checkResp = Invoke-RestMethod -Uri "$base/api/v1/papi/panels/3/check" -Method POST -Headers $headers
Write-Host "Check triggered:"
$checkResp | ConvertTo-Json

Start-Sleep -Seconds 5

# Get panel status
$panel = Invoke-RestMethod -Uri "$base/api/v1/papi/panels/3" -Method GET -Headers $headers
Write-Host "`nPanel status:"
Write-Host "  status: $($panel.status)"
Write-Host "  last_error: $($panel.last_error)"
Write-Host "  total_instances: $($panel.total_instances)"
Write-Host "  connected_instances: $($panel.connected_instances)"
Write-Host "  instances count: $($panel.instances.Count)"
if ($panel.instances.Count -gt 0) {
    foreach ($inst in $panel.instances) {
        Write-Host "  - $($inst.name) ($($inst.instance_id)): $($inst.status) / $($inst.phone_number)"
    }
}
