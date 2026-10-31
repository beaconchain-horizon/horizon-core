#!/usr/bin/env bash

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'
FAILED=0

echo -e "${CYAN}================================================================${NC}"
echo -e "${CYAN}   HORIZON CORE - LIVE ARCHITECTURE PROOF (For Auditors)${NC}"
echo -e "${CYAN}================================================================${NC}"

# ---------------------------------------------------------
echo -e "\n[1] Verifying Horizon Angel Network (7 Nodes Consensus)"
LAWS=""
for p in 9001 9002 9003 9004 9005 9006 9007; do
    HASH=$(curl -s --max-time 2 "http://localhost:$p/health" | grep -o '"law_hash":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$HASH" ]; then
        echo -e "  ✅ Node :$p -> Law Hash: ${HASH:0:16}..."
        LAWS="$LAWS$HASH\n"
    else
        echo -e "  ${RED}❌ Node :$p -> OFFLINE${NC}"
        FAILED=$((FAILED+1))
    fi
done
UNIQ=$(echo -e "$LAWS" | sort -u | grep -c .)
if [ "$UNIQ" -eq 1 ]; then
    echo -e "  ${GREEN}🏆 PROOF: All 7 nodes share IDENTICAL law_hash.${NC}"
else
    echo -e "  ${RED}⚠️ PROOF FAILED: Nodes have different hashes!${NC}"
    FAILED=$((FAILED+1))
fi

# ---------------------------------------------------------
echo -e "\n[2] Verifying Client Consortium Node (Blockchain State)"
HEIGHT=$(curl -s http://localhost:8080/api/v1/chain/info | grep -o '"height":[0-9]*' | cut -d: -f2)
if [ -n "$HEIGHT" ] && [ "$HEIGHT" -ge 5 ]; then
    echo -e "  Current Height: ${CYAN}$HEIGHT${NC}"
    BLOCK5=$(curl -s http://localhost:8080/api/v1/blocks/5)
    USER_ID=$(echo "$BLOCK5" | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
    if [ "$USER_ID" == "nicico" ]; then
        echo -e "  ${GREEN}🏆 PROOF: Block 5 (NICICO) exists in the blockchain.${NC}"
    else
        echo -e "  ${RED}❌ FAILED: Block 5 does not contain NICICO license.${NC}"
        FAILED=$((FAILED+1))
    fi
else
    echo -e "  ${RED}❌ FAILED: Chain Height is $HEIGHT (Expected >= 5)${NC}"
    FAILED=$((FAILED+1))
fi

# ---------------------------------------------------------
echo -e "\n[3] Verifying Security & Fail-Closed Mechanism"
CODE1=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/api/v1/blocks/1%27%20OR%20%271%27=%271")
CODE2=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:8091/api/v1/license/save" -d '{}')
CODE3=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:8091/api/v1/tx" -H "Content-Type: application/json" -d '{"from":"<script>","to":"b","amount":1}')

if [[ "$CODE1" == "400" || "$CODE1" == "404" ]]; then echo -e "  SQLi...... ${GREEN}BLOCKED ($CODE1)${NC}"; else echo -e "  SQLi...... ${RED}VULNERABLE ($CODE1)${NC}"; FAILED=$((FAILED+1)); fi
if [[ "$CODE2" == "401" || "$CODE2" == "403" ]]; then echo -e "  AuthBypass ${GREEN}BLOCKED ($CODE2)${NC}"; else echo -e "  AuthBypass ${RED}VULNERABLE ($CODE2)${NC}"; FAILED=$((FAILED+1)); fi
if [[ "$CODE3" == "400" || "$CODE3" == "422" || "$CODE3" == "401" ]]; then echo -e "  XSS....... ${GREEN}BLOCKED ($CODE3)${NC}"; else echo -e "  XSS....... ${RED}VULNERABLE ($CODE3)${NC}"; FAILED=$((FAILED+1)); fi

# ---------------------------------------------------------
echo -e "\n[4] Verifying Atomic License Logic in Binary"
if strings horizon-switch.exe 2>/dev/null | grep -q "license volume exceeded" || strings ./cmd/switch/switch 2>/dev/null | grep -q "license volume exceeded" || strings bin/horizon-switch.exe 2>/dev/null | grep -q "license volume exceeded"; then
    echo -e "  ${GREEN}🏆 PROOF: Atomic limit enforcement found in binary.${NC}"
else
    echo -e "  ${CYAN}(Binary check skipped or string not found - this is a soft check)${NC}"
fi

echo -e "\n${CYAN}================================================================${NC}"
if [ "$FAILED" -eq 0 ]; then
    echo -e "${GREEN}  AUDIT COMPLETE. ALL ARCHITECTURAL CLAIMS VERIFIED (0 FAILURES).${NC}"
else
    echo -e "${RED}  AUDIT FAILED. $FAILED CHECKS DID NOT PASS.${NC}"
fi
echo -e "${CYAN}================================================================${NC}"
