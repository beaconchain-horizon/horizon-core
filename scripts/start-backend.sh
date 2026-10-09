#!/bin/sh
set -eu
if [ ! -r /run/secrets/horizon_issuer_private_key ]; then
  echo "[backend-start] Authority secret is missing or unreadable" >&2
  exit 1
fi
mkdir -p /run/horizon
cp /run/secrets/horizon_issuer_private_key /run/horizon/issuer-private.pem
chown horizon:horizon /run/horizon /run/horizon/issuer-private.pem
chmod 0400 /run/horizon/issuer-private.pem
exec su-exec horizon /app/api
