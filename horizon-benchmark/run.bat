@echo off
chcp 65001 >nul
cd /d "%%~dp0"
echo === Horizon Core Benchmark ===
echo.
taskkill /F /IM switch.exe >nul 2>&1
timeout /t 3 /nobreak >nul
if exist test-keys\private.pem del test-keys\private.pem
if exist test-keys\public.pem del test-keys\public.pem
bin\sensortool.exe genkey --out=test-keys
set SWITCH_DB=:memory:
set CHAIN_CONFIG=%%cd%%\config\chain.json
set ADMIN_TOKEN=bench
set GIN_MODE=release
start /B bin\switch.exe > %%TEMP%%\horizon.log 2>&1
timeout /t 5 /nobreak >nul
echo.
echo Test: 200K readings, c=50, bs=500
bin\tpsbench.exe -url=http://127.0.0.1:8080/api/v1/industrial/reading/batch -key=test-keys\private.pem -sensor=tps-001 -n=200000 -c=50 -bs=500
echo.
echo === Integrity Check ===
curl -s http://127.0.0.1:8080/api/v1/industrial/dashboard -H "X-Admin-Token: bench"
echo.
taskkill /F /IM switch.exe >nul 2>&1
echo.
echo === Done ===
pause
