#!/bin/sh
set -eu
mkdir -p /data
chown -R horizon:horizon /data 2>/dev/null || true
if [ ! -f /data/chain.json ]; then
  if [ -f /app/config-seed/chain.json ]; then
    cp /app/config-seed/chain.json /data/chain.json
  elif [ -f /app/config-seed/chain.json.example ]; then
    cp /app/config-seed/chain.json.example /data/chain.json
  else
    echo "[start.sh] missing chain configuration; refusing to start" >&2
    exit 1
  fi
fi
if [ ! -f /data/customers.json ]; then
  printf '[]\n' > /data/customers.json
fi
chown horizon:horizon /data/chain.json /data/customers.json
echo "[start.sh] initialized persistent data directory; no demo customers seeded"
exec su-exec horizon /app/horizon-switch
