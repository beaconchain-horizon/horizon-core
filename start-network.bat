@echo off
cd /d "%~dp0"
set ADMIN_TOKEN=bench0000000000000000000000000000000000000000000000000000000bench
set SWITCH_PORT=8091
set SWITCH_DB=./data/horizon-switch.db

echo Starting Horizon Network in background...

start /b horizon-api.exe > logs\backend.log 2>&1
start /b horizon-switch.exe > logs\switch.log 2>&1

set NODE_ROLE=Merkle&& set NODE_PORT=9001&& start /b horizon-angel.exe > logs\angel-9001.log 2>&1
set NODE_ROLE=Hash&& set NODE_PORT=9002&& start /b horizon-angel.exe > logs\angel-9002.log 2>&1
set NODE_ROLE=ECDSA&& set NODE_PORT=9003&& start /b horizon-angel.exe > logs\angel-9003.log 2>&1
set NODE_ROLE=Type&& set NODE_PORT=9004&& start /b horizon-angel.exe > logs\angel-9004.log 2>&1
set NODE_ROLE=Volume&& set NODE_PORT=9005&& start /b horizon-angel.exe > logs\angel-9005.log 2>&1
set NODE_ROLE=Time&& set NODE_PORT=9006&& start /b horizon-angel.exe > logs\angel-9006.log 2>&1
set NODE_ROLE=User&& set NODE_PORT=9007&& start /b horizon-angel.exe > logs\angel-9007.log 2>&1

set AGENT_NODE_PORT=9008&& start /b horizon-agent.exe > logs\agent.log 2>&1

echo All services launched silently.
exit
