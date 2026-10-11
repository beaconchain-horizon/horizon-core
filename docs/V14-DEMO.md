# Horizon v1.4 — Live Demo

## Prerequisites

- Go 1.24
- Git Bash on Windows
- USB with the Authority private key (or a local PEM)

## Terminal 1 — Switch (with USB public key)

```bash
cd ~/Beaconchain/horizon-core
unset HORIZON_LICENSE_PUBLIC_KEY
HORIZON_ISSUER_KEY_PATH="/d/mykey.txt" bash scripts/start-switch.sh
```

Expected: `Horizon Switch v1.4 running on port 8081`.

## Terminal 2 — Browser (sign + verify)

1. Open `file:///C:/Users/NIL%20STORE/Desktop/offline-license.html`
2. **Ctrl+F5** (hard refresh)
3. Click "Choose key file from USB" → select the USB PEM
4. Fill fields (License ID, User ID, Product ID, Volume, Duration)
5. Click **Sign license**
6. Click **Verify on Switch (v1.4)**

Expected: green bar `Switch ACCEPTED - valid=true verified=true`.

## Terminal 3 — Manual verification (optional)

```bash
LID="HORIZON-GOLD-001"
LUSER="bank-markazi"
PID="horizon-core"
VOL="50000000"
MR=""
EXP_TS="1823257127"
SIG="<paste signature from browser>"

EXP=$(date -u -d "@$EXP_TS" +"%Y-%m-%dT%H:%M:%SZ")
DATA="$LID|$PID|$LUSER|$VOL|$MR|active|$EXP"

curl -s -X POST http://localhost:8081/api/v1/license/check \
  -H "Content-Type: application/json" \
  -d "{\"license_id\":\"$LID\",\"user_id\":\"$LUSER\",\"signature\":\"$SIG\",\"data\":\"$DATA\"}"
```

Expected: `{"valid":true,"verified":true,...}`.

## Tamper tests

```bash
# 1) flip first byte of signature
BAD_SIG="0${SIG:1}"
curl -s -X POST http://localhost:8081/api/v1/license/check ... \
  -d "{\"signature\":\"$BAD_SIG\",...}"
# → valid:false

# 2) change volume in data
BAD_DATA=$(echo "$DATA" | sed 's/|50000000|/|999|/')
curl -s -X POST http://localhost:8081/api/v1/license/check ... \
  -d "{\"data\":\"$BAD_DATA\",...}"
# → valid:false
```

## What this proves

1. Private key on USB never leaves the file system.
2. Signature is secp256k1, R||S, deterministic.
3. Switch verifies with public key only — no private key, no backend.
4. Any tamper is rejected.

## Known limits

- Browser page must be opened from `file://` for verify to work
  (GitHub Pages HTTPS → localhost HTTP is blocked by browsers).
- Current signing path uses the browser's noble-secp256k1;
  server-side signing via `cmd/license-signer` is available for CI.
