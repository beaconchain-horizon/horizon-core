#!/bin/bash
cd ~/Beaconchain/horizon-core || exit 1

export ADMIN_TOKEN="bench0000000000000000000000000000000000000000000000000000000bench"
export HORIZON_LICENSE_PUBLIC_KEY="04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export PATH="/c/Program Files/PostgreSQL/18/bin:$PATH"
export PGPASSWORD="postgres"

log() { echo "[$(date +%H:%M:%S)] $*"; }

ensure_switch() {
    if ! curl -s -o /dev/null --max-time 2 http://localhost:8091/health; then
        log "Switch DOWN -> restarting"
        taskkill //F //IM horizon-switch.exe 2>/dev/null
        sleep 1
        DB_ENGINE=postgres \
          DB_DSN="host=localhost port=5432 user=postgres password=postgres dbname=horizon_switch sslmode=disable" \
          HORIZON_LICENSE_PUBLIC_KEY=$HORIZON_LICENSE_PUBLIC_KEY \
          SWITCH_PORT=8091 nohup ./horizon-switch.exe > logs/switch.log 2>&1 &
        sleep 3
        curl -s -X POST http://localhost:8091/api/v1/key/unlock \
          -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" \
          -d '{"password":"authority"}' > /dev/null
        log "Switch UP + unlocked"
    fi
}

ensure_services() {
    curl -s -o /dev/null --max-time 2 http://localhost:8080/health || (PORT=8080 nohup ./horizon-api.exe > logs/backend.log 2>&1 &)
    roles=(Merkle Hash ECDSA Type Volume Time User Guardian)
    ports=(9001 9002 9003 9004 9005 9006 9007 9008)
    for i in 0 1 2 3 4 5 6 7; do
        p=${ports[$i]}
        curl -s -o /dev/null --max-time 2 http://localhost:$p/health || \
          (NODE_ROLE=${roles[$i]} NODE_PORT=$p nohup ./horizon-angel.exe > logs/angel-$p.log 2>&1 &)
    done
    curl -s -o /dev/null --max-time 2 http://localhost:9300/api/health || (BRAIN_PORT=9300 nohup ./brain.exe > logs/brain.log 2>&1 &)
}

issue_license() {
    local lic="AUTO-$(date +%s)-$RANDOM"
    curl -s -X POST http://localhost:8091/api/v1/admin/license/issue \
      -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" \
      -d "{\"license_id\":\"$lic\",\"user_id\":\"sender_1\",\"product_id\":\"horizon-core\",\"volume\":100000,\"duration\":365}" > /dev/null
    echo "$lic"
}

fire_batch() {
    local lic=$1 count=$2
    for i in $(seq 1 $count); do
        local ses=$(curl -s -X POST http://localhost:8091/api/v1/signing/session \
                    -H "X-Admin-Token: $ADMIN_TOKEN" | grep -oE '"session_token":"[^"]*"' | cut -d'"' -f4)
        curl -s -o /dev/null -X POST http://localhost:8091/api/v1/tx \
          -H "Content-Type: application/json" \
          -H "X-Admin-Token: $ADMIN_TOKEN" \
          -H "X-License-ID: $lic" \
          -H "X-Session-Token: $ses" \
          -d '{"from":"sender_1","to":"rcv","amount":1,"type":"auto"}'
    done
}

log "=== HORIZON AUTO BOOT ==="
psql -U postgres -h localhost -d horizon_switch -c "DELETE FROM transactions WHERE synced = false;" 2>/dev/null
ensure_switch
ensure_services
sleep 2

ROUND=0; OK=0; FAIL=0; LIC=""
log "=== LOOP START ==="

while true; do
    ROUND=$((ROUND+1))
    if [ -z "$LIC" ] || [ $((ROUND % 10)) -eq 1 ]; then
        LIC=$(issue_license)
        log "New license: $LIC"
    fi
    fire_batch "$LIC" 50
    RESULT=$(curl -s -X POST http://localhost:8091/api/v1/block/mine \
             -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{}')
    if echo "$RESULT" | grep -q '"status":"mined"'; then
        BNUM=$(echo "$RESULT" | grep -oE '"block_num":[0-9]+' | cut -d: -f2)
        TXN=$(echo "$RESULT" | grep -oE '"tx_count":[0-9]+' | cut -d: -f2)
        OK=$((OK+1))
        log "R$ROUND OK  block #$BNUM  tx=$TXN  total_ok=$OK"
    else
        FAIL=$((FAIL+1))
        log "R$ROUND FAIL  total_fail=$FAIL"
        psql -U postgres -h localhost -d horizon_switch -c "DELETE FROM transactions WHERE synced = false;" 2>/dev/null
    fi
    if [ $((ROUND % 5)) -eq 0 ]; then
        ensure_switch
        ensure_services
        log "Health: OK=$OK  FAIL=$FAIL"
    fi
    sleep 3
done
