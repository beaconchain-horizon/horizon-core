# Pre-Demo Checklist — Horizon v1.4

## 30 minutes before the meeting

### Terminal 0 — smoke test

\`\`\`bash
cd ~/Beaconchain/horizon-core
bash scripts/smoke-test.sh
\`\`\`

Must print: **PASS >= 6, FAIL: 0**. If any FAIL, do not proceed.

### Terminal 1 — Switch

\`\`\`bash
# kill previous instance
powershell -Command "Get-NetTCPConnection -LocalPort 8081 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess | Sort-Object -Unique | ForEach-Object { Stop-Process -Id \$_ -Force }" 2>/dev/null
sleep 2

cd ~/Beaconchain/horizon-core
unset HORIZON_LICENSE_PUBLIC_KEY
HORIZON_ISSUER_KEY_PATH="/d/mykey.txt" bash scripts/start-switch.sh
\`\`\`

Wait for: **Horizon Switch v1.4 running on port 8081**.

### Terminal 2 — sanity

\`\`\`bash
curl -s http://localhost:8081/health
# expect: {"service":"horizon-switch","status":"online",...}
\`\`\`

### Browser — the demo page

1. File Explorer → Desktop → `offline-license.html` (double-click, opens as file://)
2. **Ctrl+F5** to bypass cache
3. Click "Choose key file from USB" → select `mykey.txt`
4. Confirm public key starts with `0x04cf6aaaba...`
5. Fill fields:
   - License ID: `HORIZON-GOLD-001`
   - User ID: `bank-markazi`
   - Product ID: `horizon-core`
   - Volume: `50000000`
   - Duration: `365`
6. Click **Sign license**
7. Click **Verify on Switch (v1.4)**

Expected: green bar **Switch ACCEPTED - valid=true verified=true**.

## What to say if asked

**Q: Where is the private key?**
A: On the USB. It never enters the browser. The page reads the hex
bytes into JavaScript memory only to sign; nothing is sent to any server.

**Q: Which curve?**
A: secp256k1, the same curve Bitcoin and Ethereum use. Signature is
64 bytes R||S, deterministic per RFC 6979.

**Q: How does Switch verify without the key?**
A: Switch holds only the public key. It hashes the canonical payload
with SHA-256 and verifies the signature. It cannot sign anything.

**Q: What happens if someone tampers?**
A: Show the tamper test — one byte change in signature or payload,
and Switch returns `valid:false`.

**Q: Why not P-256?**
A: P-256 was the previous version. v1.4 uses secp256k1 for
Ethereum-compatible addresses and industry-standard tooling.

## Fallback if browser verify fails

Run the manual curl (from docs/V14-DEMO.md, terminal 3 section).
If that works, the issue is browser mixed-content — open the page
from `file://` not from GitHub Pages.

## Emergency — rollback

If nothing works and the meeting is starting, show:
1. \`docs/CRYPTO-CONTRACT.md\` — the formal spec
2. \`git log --oneline origin/main -20\` — the commit history
3. Run \`go test -v ./internal/crypto/...\` live — all tests green
4. Show \`curl -s http://localhost:8081/health\` — Switch is up

Even without the full browser demo, this proves the system works.
