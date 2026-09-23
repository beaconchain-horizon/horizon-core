#!/bin/bash
export SWITCH_PORT=8080
export SWITCH_DB="$(pwd)/data/horizon-switch.db"
export CHAIN_CONFIG="$(pwd)/config/chain.json"
export CUSTOMERS_CONFIG="$(pwd)/config/customers.json"
export ADMIN_TOKEN="1a3c9f06f4db55c8847d86bf9af20b583712fc54666d87a01c64fb84eadf9e19"
export ADMIN_PASSWORD_HASH="\$2a\$10\$BOJ9T6ZC9fbG5n1y4yeLDORlccFh3iLEZJ.axiW7IG.tSWep2nBd6"
export HORIZON_LICENSE_PUBLIC_KEY="048bea3fcc2a793347da31386d15ef67b44b8e8d47143b5c7ed2c7132f241bba969a6973a25dc510f8f091fa9f18b7e03fb1d7ffb10b215e977075ee6c43e5d34a"
export BACKEND_URL="http://localhost:8080"
export CORE_ENGINE_URL="http://localhost:8080"
export LICENSE_SERVER_URL="http://localhost:8080"
echo "Hash: \$ADMIN_PASSWORD_HASH"
echo "PubKey: \$HORIZON_LICENSE_PUBLIC_KEY"
echo "Starting..."
./switch.exe
