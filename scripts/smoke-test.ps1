# Smoke-test the running stack through the gateway (PowerShell version).
# Register -> login -> profile -> update -> refresh. Throws on failure.
#
#   ./scripts/smoke-test.ps1
#   ./scripts/smoke-test.ps1 -Base http://localhost:8081

param([string]$Base = "http://localhost:8080")

$ErrorActionPreference = "Stop"
function Step($m) { Write-Host "`n> $m" -ForegroundColor Magenta }

$email = "smoke+$([DateTimeOffset]::Now.ToUnixTimeSeconds())@example.com"
$pass  = "supersecret123"
$json  = @{ "Content-Type" = "application/json" }

Step "Health check"
Invoke-RestMethod "$Base/healthz" | ConvertTo-Json -Compress

Step "Register ($email)"
$reg = Invoke-RestMethod -Method Post "$Base/api/v1/auth/register" -Headers $json `
  -Body (@{ email = $email; name = "Smoke Tester"; password = $pass } | ConvertTo-Json)
$access  = $reg.data.accessToken
$refresh = $reg.data.refreshToken
if (-not $access) { throw "no access token returned" }
Write-Host "registered: $($reg.data.user.id)"

Step "Get profile (/users/me)"
$me = Invoke-RestMethod "$Base/api/v1/users/me" -Headers @{ Authorization = "Bearer $access" }
$me.data | ConvertTo-Json -Compress

Step "Update profile (PATCH /users/me)"
Invoke-RestMethod -Method Patch "$Base/api/v1/users/me" `
  -Headers @{ Authorization = "Bearer $access"; "Content-Type" = "application/json" } `
  -Body (@{ name = "Smoke Tester II" } | ConvertTo-Json) | Out-Null
Write-Host "updated ok"

Step "Refresh tokens"
Invoke-RestMethod -Method Post "$Base/api/v1/auth/refresh" -Headers $json `
  -Body (@{ refreshToken = $refresh } | ConvertTo-Json) | Out-Null
Write-Host "refreshed ok"

Step "Login with the new account"
Invoke-RestMethod -Method Post "$Base/api/v1/auth/login" -Headers $json `
  -Body (@{ email = $email; password = $pass } | ConvertTo-Json) | Out-Null
Write-Host "login ok"

Write-Host "`n[OK] All smoke tests passed" -ForegroundColor Green
