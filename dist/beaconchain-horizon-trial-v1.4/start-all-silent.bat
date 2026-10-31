@echo off
cd /d "C:\Users\NIL STORE\Beaconchain\horizon-core"

set ADMIN_TOKEN=12345678901234567890123456789012
set SWITCH_URL=http://localhost:8091
set HORIZON_DB_PATH=./horizon-dev.db
set SWITCH_PORT=8091
set HORIZON_LICENSE_PUBLIC_KEY=cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
set NODE_PUBKEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214

echo [%date% %time%] Horizon Angel boot sequence starting...

call :killall
ping -n 3 127.0.0.1 >nul

echo Starting 7 engine nodes...
start "" /b cmd /c "set NODE_ROLE=merkle&& set NODE_PORT=9001&& angel-node.exe >> logs\node-9001.log 2>&1"
start "" /b cmd /c "set NODE_ROLE=hash&& set NODE_PORT=9002&& angel-node.exe >> logs\node-9002.log 2>&1"
start "" /b cmd /c "set NODE_ROLE=ecdsa&& set NODE_PORT=9003&& angel-node.exe >> logs\node-9003.log 2>&1"
start "" /b cmd /c "set NODE_ROLE=type&& set NODE_PORT=9004&& angel-node.exe >> logs\node-9004.log 2>&1"
start "" /b cmd /c "set NODE_ROLE=volume&& set NODE_PORT=9005&& angel-node.exe >> logs\node-9005.log 2>&1"
start "" /b cmd /c "set NODE_ROLE=time&& set NODE_PORT=9006&& angel-node.exe >> logs\node-9006.log 2>&1"
start "" /b cmd /c "set NODE_ROLE=user&& set NODE_PORT=9007&& angel-node.exe >> logs\node-9007.log 2>&1"

ping -n 4 127.0.0.1 >nul

echo Starting switch 8091...
start "" /b cmd /c "switch.exe >> logs\switch.log 2>&1"

ping -n 4 127.0.0.1 >nul

echo Starting backend 8080...
start "" /b cmd /c "backend.exe >> logs\backend.log 2>&1"

echo [%date% %time%] Boot sequence complete.
exit /b 0

:killall
for %%P in (8080 8091 9001 9002 9003 9004 9005 9006 9007) do (
  for /f "tokens=5" %%A in ('netstat -ano ^| findstr ":%%P " ^| findstr "LISTENING"') do (
    taskkill /F /PID %%A >nul 2>&1
  )
)
exit /b 0