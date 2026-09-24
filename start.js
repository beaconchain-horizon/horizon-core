const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

const cwd = process.cwd();
const hash = fs.readFileSync(path.join(cwd, 'hash.txt'), 'utf8').trim();

console.log('Hash:', hash);
console.log('Length:', hash.length);
console.log('');

const env = Object.assign({}, process.env, {
    ADMIN_TOKEN: '1a3c9f06f4db55c8847d86bf9af20b583712fc54666d87a01c64fb84eadf9e19',
    ADMIN_PASSWORD_HASH: hash,
    SWITCH_PORT: '8080',
    SWITCH_DB: path.join(cwd, 'data', 'horizon-switch.db'),
    CHAIN_CONFIG: path.join(cwd, 'config', 'chain.json'),
    CUSTOMERS_CONFIG: path.join(cwd, 'config', 'customers.json'),
    HORIZON_LICENSE_PUBLIC_KEY: '048bea3fcc2a793347da31386d15ef67b44b8e8d47143b5c7ed2c7132f241bba969a6973a25dc510f8f091fa9f18b7e03fb1d7ffb10b215e977075ee6c43e5d34a'
});

console.log('Launching switch.exe...');
console.log('');

const child = spawn(path.join(cwd, 'switch.exe'), [], {
    cwd: cwd,
    env: env,
    stdio: 'inherit'
});

child.on('exit', code => console.log('switch exited:', code));
