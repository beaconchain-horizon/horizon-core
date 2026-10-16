@echo off
cd /d "C:\Users\NIL STORE\Beaconchain\horizon-core"

set ADMIN_TOKEN=12345678901234567890123456789012
set SWITCH_URL=http://localhost:8091
set HORIZON_DB_PATH=./horizon-dev.db
set HORIZON_LICENSE_PUBLIC_KEY=cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
set NODE_PUBKEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214

timeout /t 5 /nobreak > nul

start "" /b cmd /c "set NODE_ROLE=merkle&& set NODE_PORT=9001&& angel-node.exe"
start "" /b cmd /c "set NODE_ROLE=hash&& set NODE_PORT=9002&& angel-node.exe"
start "" /b cmd /c "set NODE_ROLE=ecdsa&& set NODE_PORT=9003&& angel-node.exe"
start "" /b cmd /c "set NODE_ROLE=type&& set NODE_PORT=9004&& angel-node.exe"
start "" /b cmd /c "set NODE_ROLE=volume&& set NODE_PORT=9005&& angel-node.exe"
start "" /b cmd /c "set NODE_ROLE=time&& set NODE_PORT=9006&& angel-node.exe"
start "" /b cmd /c "set NODE_ROLE=user&& set NODE_PORT=9007&& angel-node.exe"

timeout /t 3 /nobreak > nul

start "" /b cmd /c "set SWITCH_PORT=8091&& switch.exe"

timeout /t 3 /nobreak > nul

start "" /b backend.exe

exit
