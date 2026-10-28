#!/usr/bin/env bash
# HORIZON CORE - FULL ECOSYSTEM AUDIT
# Read-Only. Does not modify system.

BACKEND="http://localhost:8080"
SWITCH="http://localhost:8091"
PASS=0; FAIL=0; WARN=0
ok()   { echo -e "  [\033[32mPASS\033[0m] $1"; PASS=$((PASS+1)); }
bad()  { echo -e "  [\033[31mFAIL\033[0m] $1"; FAIL=$((FAIL+1)); }
warn() { echo -e "  [\033[33mWARN\033[0m] $1"; WARN=$((WARN+1)); }

echo "============================================================"
echo "  HORIZON CORE - ECOSYSTEM AUDIT REPORT"
echo "  Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo "  Dir:  $(pwd)"
echo "============================================================"

echo ""
echo "[1/6] Environment & Versioning"
GO_VER=$(go version 2>/dev/null | awk '{print $3}')
if [[ "$GO_VER" == go1.24* ]]; then
    ok "Go version: $GO_VER (Red Line Intact)"
elif [ -n "$GO_VER" ]; then
    bad "Go version: $GO_VER (Expected 1.24)"
else
    bad "Go is not installed or not in PATH"
fi

GIT_DIRTY=$(git status --porcelain 2>/dev/null | wc -l)
if [ "$GIT_DIRTY" -eq 0 ]; then
    ok "Git tree: CLEAN (Build Provenance intact)"
else
    warn "Git tree: $GIT_DIRTY uncommitted changes found"
fi

echo ""
echo "[2/6] Services Health"
if curl -sf --max-time 2 "$BACKEND/api/v1/health" >/dev/null; then ok "Backend (8080) is ONLINE"; else bad "Backend (8080) is DOWN"; fi
if curl -sf --max-time 2 "$SWITCH/api/v1/health" >/dev/null; then ok "Switch (8091) is ONLINE"; else bad "Switch (8091) is DOWN"; fi

echo ""
echo "[3/6] 7-Node Consensus (Horizon Angel)"
LAWS=""
ALIVE=0
for p in 9001 9002 9003 9004 9005 9006 9007; do
    HASH=$(curl -s --max-time 2 "http://localhost:$p/health" | grep -o '"law_hash":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$HASH" ]; then
        LAWS="$LAWS$HASH\n"
        ALIVE=$((ALIVE+1))
        ok "Node :$p -> ${HASH:0:12}..."
    else
        bad "Node :$p -> UNREACHABLE"
    fi
done
UNIQ=$(echo -e "$LAWS" | sort -u | grep -c .)
if [ "$UNIQ" -eq 1 ] && [ "$ALIVE" -eq 7 ]; then
    ok "Consensus: 7/7 nodes share IDENTICAL law_hash"
elif [ "$UNIQ" -gt 1 ]; then
    bad "Consensus: MISMATCH! ($UNIQ unique hashes)"
else
    warn "Consensus: Only $ALIVE/7 nodes reachable"
fi

echo ""
echo "[4/6] Database & Chain State"
for DB in "horizon-dev.db" "horizon-switch.db"; do
    if [ -f "$DB" ]; then
        SIZE=$(stat -c %s "$DB" 2>/dev/null || wmic datafile where name="" 2>/dev/null)
        ok "DB found: $DB ($(du -h "$DB" | cut -f1))"
    else
        warn "DB not found in current dir: $DB"
    fi
done

INFO=$(curl -s --max-time 2 "$BACKEND/api/v1/chain/info")
if echo "$INFO" | grep -q '"chain_id"'; then
    HEIGHT=$(echo "$INFO" | grep -o '"height":[0-9]*' | cut -d: -f2)
    TXS=$(echo "$INFO" | grep -o '"tx_total":[0-9]*' | cut -d: -f2)
    ok "Chain API active: Height=$HEIGHT, Total_TXs=$TXS"
else
    bad "Chain API unreachable or invalid response"
fi

echo ""
echo "[5/6] Security & Secrets (OpSec Check)"
KEYFILE=".horizon-dev-key.pem"
if [ -f "$KEYFILE" ]; then
    KHASH=$(sha256sum "$KEYFILE" | cut -c1-16)
    ok "Private key exists (sha256: $KHASH...)"
    if git ls-files --error-unmatch "$KEYFILE" >/dev/null 2>&1; then
        bad "CRITICAL: Private key is TRACKED by Git!"
    else
        ok "Private key is safely ignored by Git"
    fi
else
    warn "Private key $KEYFILE not found in current dir"
fi

if [ -n "$ADMIN_TOKEN" ]; then
    ok "ADMIN_TOKEN is set in environment (len: ${#ADMIN_TOKEN})"
else
    warn "ADMIN_TOKEN is NOT set in current shell"
fi

echo ""
echo "[6/6] Recent Errors (if any log files exist)"
LOG_COUNT=$(ls *.log 2>/dev/null | wc -l)
if [ "$LOG_COUNT" -gt 0 ]; then
    ERRS=$(grep -i "panic\|fatal\|error" *.log 2>/dev/null | tail -n 3)
    if [ -n "$ERRS" ]; then
        warn "Recent errors found in logs:"
        echo "$ERRS" | sed 's/^/         /'
    else
        ok "Logs exist but no recent PANIC/FATAL errors found"
    fi
else
    ok "No *.log files in root dir (logs might be in stdout or elsewhere)"
fi

echo ""
echo "============================================================"
echo "  AUDIT SUMMARY"
echo "============================================================"
echo -e "  [PASS]: \033[32m$PASS\033[0m"
echo -e "  [WARN]: \033[33m$WARN\033[0m"
echo -e "  [FAIL]: \033[31m$FAIL\033[0m"
echo "============================================================"
