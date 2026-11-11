$ErrorActionPreference = "SilentlyContinue"
Write-Host "1. Stopping old services..." -ForegroundColor Cyan
Get-Process horizon-switch,horizon-angel,horizon-agent,horizon-api,brain | Stop-Process -Force
Start-Sleep -Seconds 2

if (!(Test-Path logs)) { New-Item -ItemType Directory -Path logs | Out-Null }

# --- تنظیمات مشترک ---
$env:DB_ENGINE="postgres"
$env:DB_DSN="host=localhost port=5432 user=postgres password=postgres dbname=horizon_switch sslmode=disable"
$env:HORIZON_LICENSE_PUBLIC_KEY="04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
$env:ADMIN_TOKEN="bench0000000000000000000000000000000000000000000000000000000bench"

# --- 1. Switch ---
Write-Host "2. Starting Switch (8091)..." -ForegroundColor Green
$env:SWITCH_PORT="8091"
Start-Process -FilePath ".\horizon-switch.exe" -RedirectStandardOutput "logs\switch.log" -RedirectStandardError "logs\switch.err" -WindowStyle Hidden

# --- 2. API ---
Write-Host "3. Starting API (8080)..." -ForegroundColor Green
$env:PORT="8080"
$env:CN_TOKEN="bench0000000000000000000000000000000000000000000000000000000bench"
$env:API_KEY="bench0000000000000000000000000000000000000000000000000000000bench"
Start-Process -FilePath ".\horizon-api.exe" -RedirectStandardOutput "logs\api.log" -RedirectStandardError "logs\api.err" -WindowStyle Hidden

# --- 3. Angels ---
Write-Host "4. Starting 8 Angels..." -ForegroundColor Green
$roles = @{
    "9001"="Merkle"; "9002"="Hash"; "9003"="ECDSA"; "9004"="Type";
    "9005"="Volume"; "9006"="Time"; "9007"="User"; "9008"="AgentGuardian"
}
foreach ($port in $roles.Keys) {
    $env:NODE_ROLE = $roles[$port]
    $env:NODE_PORT = "$port"
    $exe = if ($port -eq "9008") { ".\horizon-agent.exe" } else { ".\horizon-angel.exe" }
    Start-Process -FilePath $exe -RedirectStandardOutput "logs\angel-$port.log" -RedirectStandardError "logs\angel-$port.err" -WindowStyle Hidden
}

# --- 4. Brain ---
Write-Host "5. Starting Brain (9300)..." -ForegroundColor Green
Start-Process -FilePath ".\brain.exe" -RedirectStandardOutput "logs\brain.log" -RedirectStandardError "logs\brain.err" -WindowStyle Hidden

Write-Host "Waiting 5 seconds for services to boot..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# --- بررسی نهایی ---
Write-Host "`n--- Health Check ---" -ForegroundColor Cyan
$fail = 0
try { Invoke-RestMethod -Uri "http://localhost:8091/health" -TimeoutSec 2 | Out-Null; Write-Host "  :8091 Switch     OK" -ForegroundColor Green } catch { Write-Host "  :8091 Switch     FAIL" -ForegroundColor Red; $fail++ }
# تغییر مسیر چک کردن API به /health تا نیازی به توکن نداشته باشد
try { Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 2 | Out-Null; Write-Host "  :8080 API        OK" -ForegroundColor Green } catch { Write-Host "  :8080 API        FAIL" -ForegroundColor Red; $fail++ }
foreach ($port in $roles.Keys) {
    try { Invoke-RestMethod -Uri "http://localhost:$port/health" -TimeoutSec 2 | Out-Null; Write-Host "  :$port Angel      OK" -ForegroundColor Green } catch { Write-Host "  :$port Angel      FAIL" -ForegroundColor Red; $fail++ }
}
try { Invoke-RestMethod -Uri "http://localhost:9300/health" -TimeoutSec 2 | Out-Null; Write-Host "  :9300 Brain      OK" -ForegroundColor Green } catch { Write-Host "  :9300 Brain      (Optional)" -ForegroundColor Yellow }

if ($fail -eq 0) {
    Write-Host "`n✅ ALL SYSTEMS GO." -ForegroundColor Green
} else {
    Write-Host "`n⚠️ $fail services failed to start. Check the 'logs' folder." -ForegroundColor Red
}
