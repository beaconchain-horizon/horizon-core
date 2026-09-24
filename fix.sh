#!/bin/bash
cd ~/Beaconchain/horizon-core

# 1. kill switch
taskkill //F //IM switch.exe 2>/dev/null
sleep 2

# 2. ساخت hash با $2a$
node -e "const b=require('bcryptjs');const h=b.hashSync('admin123',10).replace(/^\\\$2b\\\$/, '\\\$2a\\\$');require('fs').writeFileSync('hash.txt',h);console.log('Hash:',h)"

# 3. تست
node -e "const b=require('bcryptjs');const h=require('fs').readFileSync('hash.txt','utf8').trim();console.log('Match:',b.compareSync('admin123',h))"

# 4. اجرای switch
export ADMIN_TOKEN="1a3c9f06f4db55c8847d86bf9af20b583712fc54666d87a01c64fb84eadf9e19"
export ADMIN_PASSWORD_HASH="$(cat hash.txt)"
export SWITCH_PORT=8080
export SWITCH_DB="$(pwd)/data/horizon-switch.db"
export CHAIN_CONFIG="$(pwd)/config/chain.json"
export CUSTOMERS_CONFIG="$(pwd)/config/customers.json"
export HORIZON_LICENSE_PUBLIC_KEY="048bea3fcc2a793347da31386d15ef67b44b8e8d47143b5c7ed2c7132f241bba969a6973a25dc510f8f091fa9f18b7e03fb1d7ffb10b215e977075ee6c43e5d34a"

echo ""
echo "=== Hash: ${ADMIN_PASSWORD_HASH:0:30}..."
echo "=== Starting switch..."
echo ""

./switch.exe
