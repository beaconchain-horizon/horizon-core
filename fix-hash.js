const fs = require('fs');
const hash = fs.readFileSync('hash.txt', 'utf8').trim();
const sh = `#!/bin/bash
export ADMIN_TOKEN=REPLACE_WITH_ADMIN_TOKEN
export ADMIN_PASSWORD_HASH='${hash}'
export SWITCH_PORT=8080
export SWITCH_DB="$(pwd)/data/horizon-switch.db"
export CHAIN_CONFIG="$(pwd)/config/chain.json"
export CUSTOMERS_CONFIG="$(pwd)/config/customers.json"
export HORIZON_LICENSE_PUBLIC_KEY=048bea3fcc2a793347da31386d15ef67b44b8e8d47143b5c7ed2c7132f241bba969a6973a25dc510f8f091fa9f18b7e03fb1d7ffb10b215e977075ee6c43e5d34a
./switch.exe
`;
fs.writeFileSync('start.sh', sh);
console.log('✓ start.sh ساخته شد');
console.log('✓ Hash length:', hash.length);
