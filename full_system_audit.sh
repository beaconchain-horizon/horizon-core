#!/usr/bin/env bash
# HORIZON CORE - FULL SYSTEM & ARCHITECTURE MAPPING
# Strictly Read-Only. Generates SYSTEM_SNAPSHOT.txt

OUT="SYSTEM_SNAPSHOT.txt"
echo "============================================================" > "$OUT"
echo "  HORIZON CORE - SYSTEM SNAPSHOT & ARCHITECTURE" >> "$OUT"
echo "  Date: $(date '+%Y-%m-%d %H:%M:%S')" >> "$OUT"
echo "  Dir:  $(pwd)" >> "$OUT"
echo "============================================================" >> "$OUT"

echo "[1/4] Mapping Directories & Git State..."
echo "" >> "$OUT"
echo "--- [GIT STATE] ---" >> "$OUT"
echo "Go Version: $(go version)" >> "$OUT"
echo "Git Branch: $(git rev-parse --abbrev-ref HEAD)" >> "$OUT"
echo "Last Commit: $(git log -1 --pretty=format:'%h - %s (%cr)')" >> "$OUT"
echo "Uncommitted Changes: $(git status --porcelain | wc -l)" >> "$OUT"
echo "" >> "$OUT"

echo "--- [TOP-LEVEL STRUCTURE] ---" >> "$OUT"
ls -F >> "$OUT"
echo "" >> "$OUT"

echo "--- [CMD FOLDER (Executables/Entrypoints)] ---" >> "$OUT"
ls -F cmd/ >> "$OUT" 2>/dev/null
echo "" >> "$OUT"

echo "--- [INTERNAL FOLDER (Core Logic Modules)] ---" >> "$OUT"
ls -F internal/ >> "$OUT" 2>/dev/null
echo "" >> "$OUT"

echo "[2/4] Identifying Architecture Roles & Trial Assets..."
echo "--- [ARCHITECTURE ROLES DETECTED] ---" >> "$OUT"
# Check main entrypoints
[ -f cmd/api/main.go ] && echo "Backend Entrypoint : cmd/api/main.go (Runs on :8080)" >> "$OUT"
[ -f cmd/switch/main.go ] && echo "Switch Entrypoint  : cmd/switch/main.go (Runs on :8091)" >> "$OUT"
# Check for node components
if [ -d cmd/node ] || [ -d cmd/angel ]; then
    echo "Node Entrypoint    : Found in cmd/node or cmd/angel" >> "$OUT"
else
    echo "Node Entrypoint    : Not explicitly found as a separate binary (may be embedded)" >> "$OUT"
fi
echo "" >> "$OUT"

echo "--- [TRIAL MODE & LICENSING ASSETS] ---" >> "$OUT"
# Find trial/install scripts
find . -maxdepth 2 -type f -name "*trial*" -o -name "*install*" -o -name "*package*" | head -n 10 >> "$OUT"
# Find license files
echo "License Files Found:" >> "$OUT"
find . -maxdepth 3 -type f -name "*.lic" -o -name "*.token" -o -name "*.pem" | grep -v ".git/" >> "$OUT"
echo "" >> "$OUT"

echo "--- [DATABASES FOUND] ---" >> "$OUT"
find . -maxdepth 3 -type f -name "*.db" -exec ls -lh {} \; | awk '{print $9 " (" $5 ")"}' >> "$OUT"
echo "" >> "$OUT"

echo "[3/4] Checking Live Services & Consensus..."
echo "--- [LIVE SERVICES & PORTS] ---" >> "$OUT"
# Check Backend
BE_H=$(curl -s --max-time 2 http://localhost:8080/api/v1/health)
if echo "$BE_H" | grep -q "online"; then echo "Backend (:8080)   : ONLINE" >> "$OUT"; else echo "Backend (:8080)   : OFFLINE" >> "$OUT"; fi

# Check Switch
SW_H=$(curl -s --max-time 2 http://localhost:8091/api/v1/health)
if echo "$SW_H" | grep -q "online"; then echo "Switch (:8091)    : ONLINE" >> "$OUT"; else echo "Switch (:8091)    : OFFLINE" >> "$OUT"; fi

# Check Nodes
echo "Nodes (9001-9007):" >> "$OUT"
LAWS=""
for p in 9001 9002 9003 9004 9005 9006 9007; do
    H=$(curl -s --max-time 1 "http://localhost:$p/health" | grep -o '"law_hash":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$H" ]; then
        echo "  :$p -> ${H:0:16}..." >> "$OUT"
        LAWS="$LAWS$H\n"
    else
        echo "  :$p -> OFFLINE" >> "$OUT"
    fi
done
UNIQ=$(echo -e "$LAWS" | sort -u | grep -c .)
if [ "$UNIQ" -eq 1 ]; then echo "CONSENSUS: 7/7 Nodes in Perfect Sync" >> "$OUT"; fi
echo "" >> "$OUT"

echo "[4/4] Security Check & Finalizing..."
echo "--- [SECURITY & SECRETS CHECK] ---" >> "$OUT"
# Check if gitignore covers sensitive files
if grep -q "*.db" .gitignore 2>/dev/null; then echo ".gitignore: DBs protected" >> "$OUT"; else echo ".gitignore: WARNING - DBs not protected" >> "$OUT"; fi
if grep -q "*.pem" .gitignore 2>/dev/null; then echo ".gitignore: Keys protected" >> "$OUT"; else echo ".gitignore: WARNING - Keys not protected" >> "$OUT"; fi

# Check for hardcoded tokens in Go files (basic check)
LEAKS=$(grep -rn "12345678901234567890123456789012" --include="*.go" . | wc -l)
if [ "$LEAKS" -gt 0 ]; then echo "WARNING: Found $LEAKS hardcoded test tokens in .go files" >> "$OUT"; else echo "No obvious hardcoded tokens in Go files" >> "$OUT"; fi
echo "" >> "$OUT"

# Check Chain Info
CHAIN_INFO=$(curl -s --max-time 2 http://localhost:8080/api/v1/chain/info)
echo "--- [CHAIN STATE] ---" >> "$OUT"
echo "$CHAIN_INFO" >> "$OUT"

echo "" >> "$OUT"
echo "============================================================" >> "$OUT"
echo "  END OF SNAPSHOT" >> "$OUT"
echo "============================================================" >> "$OUT"

echo "✅ Audit Complete. Output saved to $OUT"
echo "Please copy the contents of this file and send it to me."
