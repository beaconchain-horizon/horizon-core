#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

fail() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
for cmd in go docker openssl curl awk; do command -v "$cmd" >/dev/null 2>&1 || fail "required command missing: $cmd"; done
go version | grep -Eq 'go1\.24\.' || fail "Go 1.24.x is required; detected: $(go version)"
docker compose version >/dev/null 2>&1 || fail "Docker Compose v2 is required"

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BACKUP=".backups/horizon-v14-bootstrap-$STAMP"
mkdir -p "$BACKUP"

if [[ -f .env ]]; then cp -p .env "$BACKUP/env.before"; fi
PRIVATE="secrets/v1.4/issuer-private.pem"
PUBLIC="secrets/v1.4/issuer-public.hex"
mkdir -p "$(dirname "$PRIVATE")"

env_value() {
  local key="$1"
  [[ -f .env ]] || return 0
  awk -F= -v k="$key" '$1==k {sub(/^[^=]*=/, ""); v=$0} END {print v}' .env
}

PROTOCOL="$(env_value AUTHORITY_PROTOCOL_VERSION)"
if [[ "$PROTOCOL" == "v1.4-secp256k1" && -s "$PRIVATE" && -s "$PUBLIC" ]]; then
  printf 'Reusing the already provisioned Horizon v1.4 Authority keypair.\n'
  ADMIN_TOKEN="$(env_value ADMIN_TOKEN)"
  API_KEY="$(env_value API_KEY)"
  POSTGRES_PASSWORD="$(env_value POSTGRES_PASSWORD)"
  ANGEL_ADMIN_PASSWORD="$(env_value ANGEL_ADMIN_PASSWORD)"
  ANGEL_SESSION_SECRET="$(env_value ANGEL_SESSION_SECRET)"
else
  # The previous issuer identity is test-era material. Preserve it in a dated
  # backup, rotate auth tokens, and create a clean v1.4 Authority root of trust.
  for keyfile in "$PRIVATE" "$PUBLIC"; do
    if [[ -e "$keyfile" ]]; then
      rel="${keyfile#secrets/}"
      mkdir -p "$BACKUP/secrets/$(dirname "$rel")"
      cp -p "$keyfile" "$BACKUP/secrets/$rel"
      rm -f "$keyfile"
    fi
  done
  go run ./cmd/authority-keygen -private-out "$PRIVATE" -public-out "$PUBLIC"
  ADMIN_TOKEN="$(openssl rand -hex 32)"
  API_KEY="$(openssl rand -hex 32)"
  POSTGRES_PASSWORD="$(openssl rand -hex 32)"
  ANGEL_ADMIN_PASSWORD="$(openssl rand -hex 32)"
  ANGEL_SESSION_SECRET="$(openssl rand -hex 32)"
fi

[[ "${#ADMIN_TOKEN}" -ge 32 ]] || ADMIN_TOKEN="$(openssl rand -hex 32)"
[[ "${#API_KEY}" -ge 32 ]] || API_KEY="$(openssl rand -hex 32)"
[[ "${#POSTGRES_PASSWORD}" -ge 32 ]] || POSTGRES_PASSWORD="$(openssl rand -hex 32)"
[[ "${#ANGEL_ADMIN_PASSWORD}" -ge 32 ]] || ANGEL_ADMIN_PASSWORD="$(openssl rand -hex 32)"
[[ "${#ANGEL_SESSION_SECRET}" -ge 32 ]] || ANGEL_SESSION_SECRET="$(openssl rand -hex 32)"
HORIZON_LICENSE_PUBLIC_KEY="$(tr -d '\r\n' < "$PUBLIC")"

# Keep unrelated environment settings (for example optional AI integrations)
# but replace the keys managed by this bootstrapper.
TMP_ENV="$(mktemp)"
if [[ -f .env ]]; then
  awk -F= '
    $1=="POSTGRES_PASSWORD" || $1=="ADMIN_TOKEN" || $1=="API_KEY" ||
    $1=="HORIZON_LICENSE_PUBLIC_KEY" || $1=="AUTHORITY_PROTOCOL_VERSION" ||
    $1=="LICENSE_PRIVATE_KEY_FILE" || $1=="SWITCH_URL" || $1=="ENABLE_SWITCH" ||
    $1=="ANGEL_ADMIN_PASSWORD" || $1=="ANGEL_SESSION_SECRET" {next}
    {print}
  ' .env > "$TMP_ENV"
fi
cat >> "$TMP_ENV" <<EOF
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
ADMIN_TOKEN=$ADMIN_TOKEN
API_KEY=$API_KEY
HORIZON_LICENSE_PUBLIC_KEY=$HORIZON_LICENSE_PUBLIC_KEY
AUTHORITY_PROTOCOL_VERSION=v1.4-secp256k1
LICENSE_PRIVATE_KEY_FILE=/run/horizon/issuer-private.pem
SWITCH_URL=http://switch:8080
ENABLE_SWITCH=true
ANGEL_ADMIN_PASSWORD=$ANGEL_ADMIN_PASSWORD
ANGEL_SESSION_SECRET=$ANGEL_SESSION_SECRET
EOF
mv "$TMP_ENV" .env
chmod 600 .env

docker compose config >/dev/null || fail "docker-compose.yml configuration check failed"
printf '\n== Building and starting Horizon Core v1.4 ==\n'
docker compose up -d --build

ok=0
for _ in $(seq 1 60); do
  if curl -fsS http://127.0.0.1:8080/health >/dev/null 2>&1 &&
     curl -fsS http://127.0.0.1:8081/api/v1/health >/dev/null 2>&1 &&
     curl -fsS http://127.0.0.1:3000/angel/login >/dev/null 2>&1; then
    ok=1
    break
  fi
  sleep 2
done

if [[ "$ok" != 1 ]]; then
  printf '\nHealth check did not pass. Recent logs follow:\n' >&2
  docker compose ps >&2 || true
  docker compose logs --tail=80 backend switch postgres >&2 || true
  fail "Horizon Core did not become healthy; no automatic cleanup was performed"
fi

printf '\nHorizon Core v1.4 is up.\n'
printf 'Switch health:  http://127.0.0.1:8080/health\n'
printf 'Backend health: http://127.0.0.1:8081/api/v1/health\n'
printf 'Store:           http://127.0.0.1:3000/\n'
printf 'Owner console:   http://127.0.0.1:3000/angel/login\n'
printf 'Owner password is stored in .env; read it locally and never share it.\n'
printf 'Issuer public key: %s\n' "$PUBLIC"
printf 'Private key: %s (do not upload or open in a browser)\n' "$PRIVATE"
printf 'Backup: %s\n' "$BACKUP"
docker compose ps
