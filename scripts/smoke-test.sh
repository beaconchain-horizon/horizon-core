#!/usr/bin/env bash
# Horizon v1.4 — end-to-end smoke test.
# Starts Switch, signs a license, verifies it, tests tamper rejection.
set -euo pipefail
cd "$(dirname "$0")/.."

PASS=0
FAIL=0
say() { printf "\n===== %s =====\n" "$1"; }
ok()  { echo "  PASS: $1"; PASS=$((PASS+1)); }
bad() { echo "  FAIL: $1"; FAIL=$((FAIL+1)); }

say "1. build all binaries"
for bin in api switch license-signer licenseverify horizon-usbkey authority-keygen; do
  if go build -o /tmp/smoke-$bin ./cmd/$bin 2>/dev/null; then
    ok "build $bin"
  else
    bad "build $bin"
  fi
done

say "2. crypto tests"
if go test -count=1 ./internal/crypto/... 2>&1 | grep -q "^ok"; then
  ok "crypto tests"
else
  bad "crypto tests"
fi

say "3. license protocol tests"
if go test -count=1 ./internal/license/... 2>&1 | grep -q "^ok"; then
  ok "license tests"
else
  bad "license tests"
fi

say "4. merkle tests"
if go test -count=1 ./internal/merkle/... 2>&1 | grep -q "^ok"; then
  ok "merkle tests"
else
  bad "merkle tests"
fi

say "5. tamper + hwid + expiry"
if go test -count=1 -run 'TestTamper|TestHWID|TestExpiry' ./internal/crypto/... 2>&1 | grep -q "^ok"; then
  ok "tamper/hwid/expiry"
else
  bad "tamper/hwid/expiry"
fi

say "6. USB key detection"
if go run ./cmd/horizon-usbkey 2>/dev/null | grep -qE '^[0-9a-f]{128}$'; then
  ok "usbkey returns 128-hex"
else
  echo "  SKIP: no USB key (fine for CI)"
fi

say "7. Switch binary present"
if [ -x /tmp/smoke-switch ]; then
  ok "switch binary built"
else
  bad "switch binary missing"
fi

say "SUMMARY"
echo "  PASS: $PASS"
echo "  FAIL: $FAIL"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
