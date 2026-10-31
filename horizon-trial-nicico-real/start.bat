@echo off
echo Starting Horizon Core Trial (Real License)...
set HORIZON_TRIAL_MODE=true
set ADMIN_TOKEN=12345678901234567890123456789012
set HORIZON_LICENSE_PUBLIC_KEY=cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
set SERVER_PORT=8080
set SWITCH_PORT=8091

start "Horizon-API" bin\horizon-api.exe
start "Horizon-Switch" bin\horizon-switch.exe

echo Waiting for services to start...
timeout /t 5 /nobreak >nul

echo Registering REAL license to Backend...
curl.exe -s -X POST http://localhost:8080/api/v1/license/register -H "Content-Type: application/json" -H "X-Admin-Key: %ADMIN_TOKEN%" -d @config/trial.lic
echo.
echo System Ready! Open http://localhost:8080 in your browser.
pause
