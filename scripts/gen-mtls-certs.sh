#!/usr/bin/env bash
# Generate self-signed CA + server + client certs for Switch<->Backend mTLS.
# Air-gap friendly: uses only openssl, no network.
set -euo pipefail
OUT=\${1:-scripts/certs}
mkdir -p "$OUT"
cd "$OUT"
if [ ! -f ca.pem ]; then
  echo "[1] CA"
  openssl ecparam -name secp256k1 -genkey -noout -out ca.key
  openssl req -x509 -new -key ca.key -days 3650 -subj "/CN=Horizon-CA" -out ca.pem
fi
if [ ! -f server.pem ]; then
  echo "[2] server"
  openssl ecparam -name secp256k1 -genkey -noout -out server.key
  openssl req -new -key server.key -subj "/CN=horizon-backend" -out server.csr
  openssl x509 -req -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial -days 825 -out server.pem
fi
if [ ! -f client.pem ]; then
  echo "[3] client"
  openssl ecparam -name secp256k1 -genkey -noout -out client.key
  openssl req -new -key client.key -subj "/CN=horizon-switch" -out client.csr
  openssl x509 -req -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial -days 825 -out client.pem
fi
echo done
