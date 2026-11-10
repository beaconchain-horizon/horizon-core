@echo off
cd /d "%~dp0"
echo ========================================
echo   HORIZON CORE - DUAL NETWORK START
echo   Online (8091) + Offline (8092) + 8 Nodes
echo ========================================
echo.

echo [1/4] Ensuring PostgreSQL is running...
net start postgresql-x64-18 >nul 2>&1
timeout /t 3 /nobreak >nul

echo [2/4] Starting 8 Consensus Nodes (9001-9008)...
set ADMIN_TOKEN=bench0000000000000000000000000000000000000000000000000000000bench
set HORIZON_LICENSE_PUBLIC_KEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214

set NODE_ROLE=Merkle&& set NODE_PORT=9001&& start /B horizon-angel.exe > logs\angel-9001.log 2>&1
set NODE_ROLE=Hash&& set NODE_PORT=9002&& start /B horizon-angel.exe > logs\angel-9002.log 2>&1
set NODE_ROLE=ECDSA&& set NODE_PORT=9003&& start /B horizon-angel.exe > logs\angel-9003.log 2>&1
set NODE_ROLE=Type&& set NODE_PORT=9004&& start /B horizon-angel.exe > logs\angel-9004.log 2>&1
set NODE_ROLE=Volume&& set NODE_PORT=9005&& start /B horizon-angel.exe > logs\angel-9005.log 2>&1
set NODE_ROLE=Time&& set NODE_PORT=9006&& start /B horizon-angel.exe > logs\angel-9006.log 2>&1
set NODE_ROLE=User&& set NODE_PORT=9007&& start /B horizon-angel.exe > logs\angel-9007.log 2>&1
set AGENT_NODE_PORT=9008&& start /B horizon-agent.exe > logs\agent.log 2>&1

timeout /t 3 /nobreak >nul

echo [3/4] Starting ONLINE Switch (8091 - PostgreSQL)...
set DB_ENGINE=postgres
set DB_DSN=host=localhost port=5432 user=postgres password=postgres dbname=horizon_switch sslmode=disable
set SWITCH_PORT=8091
start /B horizon-switch.exe > logs\switch-online.log 2>&1

timeout /t 2 /nobreak >nul

echo [4/4] Starting OFFLINE Switch (8092 - SQLite)...
set DB_ENGINE=sqlite
set DB_DSN=
set SWITCH_DB=./data/horizon-switch-offline.db
set SWITCH_PORT=8092
start /B horizon-switch.exe > logs\switch-offline.log 2>&1

timeout /t 3 /nobreak >nul
echo.
echo ========================================
echo   ALL SERVICES STARTED
echo   Online  Switch: http://localhost:8091/health
echo   Offline Switch: http://localhost:8092/health
echo   Nodes: 9001-9008
echo ========================================
pause
