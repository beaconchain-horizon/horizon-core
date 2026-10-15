@echo off
REM موتور: ۷ نود، ۷ پورت، ۷ معادله
set GO=go
set NODE_PUBKEY=%HORIZON_LICENSE_PUBLIC_KEY%

echo === Starting Horizon Angel Engine - 7 nodes ===

start "AngelV1-Merkle" cmd /k "set NODE_ROLE=merkle&& set NODE_PORT=9001&& %GO% run ./cmd/angel-node"
start "AngelV2-Hash"   cmd /k "set NODE_ROLE=hash&&   set NODE_PORT=9002&& %GO% run ./cmd/angel-node"
start "AngelV3-ECDSA"  cmd /k "set NODE_ROLE=ecdsa&&  set NODE_PORT=9003&& %GO% run ./cmd/angel-node"
start "AngelV4-Type"   cmd /k "set NODE_ROLE=type&&   set NODE_PORT=9004&& %GO% run ./cmd/angel-node"
start "AngelV5-Volume" cmd /k "set NODE_ROLE=volume&& set NODE_PORT=9005&& %GO% run ./cmd/angel-node"
start "AngelV6-Time"   cmd /k "set NODE_ROLE=time&&   set NODE_PORT=9006&& %GO% run ./cmd/angel-node"
start "AngelV7-User"   cmd /k "set NODE_ROLE=user&&   set NODE_PORT=9007&& %GO% run ./cmd/angel-node"

echo === 7 nodes launched in separate windows ===
