#!/usr/bin/env bash
cd ~/Beaconchain/horizon-core
export HORIZON_DB_PATH=/tmp/horizon-chain.db
export HORIZON_LICENSE_PUBLIC_KEY=cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
export SERVER_PORT=8080
export ENABLE_SWITCH=false
export ADMIN_TOKEN=ad44e42de34b83cc196719512b91c8314798e6825629aa77753310e091a95b58
export GIN_MODE=release
nohup ./backend.exe > /tmp/backend.log 2>&1 &
disown
echo "backend started"
