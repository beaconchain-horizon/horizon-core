@echo off
cd /d "C:\Users\NIL STORE\Beaconchain\horizon-core"
echo Waiting for Windows and PostgreSQL to initialize...
timeout /t 15 /nobreak >nul

set ADMIN_TOKEN=bench0000000000000000000000000000000000000000000000000000000bench
set DB_ENGINE=postgres
set DB_DSN=host=localhost port=5432 user=postgres password=postgres dbname=horizon_switch sslmode=disable
set SWITCH_PORT=8091

:: Start 8 Nodes
set NODE_ROLE=Merkle&& set NODE_PORT=9001&& start /B horizon-angel.exe > logs\angel-9001.log 2>&1
set NODE_ROLE=Hash&& set NODE_PORT=9002&& start /B horizon-angel.exe > logs\angel-9002.log 2>&1
set NODE_ROLE=ECDSA&& set NODE_PORT=9003&& start /B horizon-angel.exe > logs\angel-9003.log 2>&1
set NODE_ROLE=Type&& set NODE_PORT=9004&& start /B horizon-angel.exe > logs\angel-9004.log 2>&1
set NODE_ROLE=Volume&& set NODE_PORT=9005&& start /B horizon-angel.exe > logs\angel-9005.log 2>&1
set NODE_ROLE=Time&& set NODE_PORT=9006&& start /B horizon-angel.exe > logs\angel-9006.log 2>&1
set NODE_ROLE=User&& set NODE_PORT=9007&& start /B horizon-angel.exe > logs\angel-9007.log 2>&1
set AGENT_NODE_PORT=9008&& start /B horizon-agent.exe > logs\agent.log 2>&1

:: Start Online Switch
start /B horizon-switch.exe > logs\switch-online.log 2>&1

echo Network is UP and RUNNING (but LOCKED for Air-Gap security).
echo Please insert your USB and run the manual unlock command to start signing.
