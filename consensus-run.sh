#!/bin/bash
set -e
export ADMIN_TOKEN="bench0000000000000000000000000000000000000000000000000000000bench"
export PATH="/c/Program Files/PostgreSQL/18/bin:$PATH"
export PGPASSWORD="postgres"
H="X-Admin-Token: $ADMIN_TOKEN"

echo "⏳ [1/4] Stopping old processes..."
taskkill //F //IM horizon-switch.exe 2>/dev/null || true
taskkill //F //IM horizon-angel.exe 2>/dev/null || true
sleep 2

echo "⏳ [2/4] Starting Switch and 8 Consensus Nodes..."
DB_ENGINE=postgres DB_DSN="host=localhost port=5432 user=postgres password=postgres dbname=horizon_switch sslmode=disable" SWITCH_PORT=8091 nohup ./horizon-switch.exe > logs/switch-online.log 2>&1 &

ROLES=(Merkle Hash ECDSA Type Volume Time User)
PORTS=(9001 9002 9003 9004 9005 9006 9007)
for i in 0 1 2 3 4 5 6; do
  NODE_ROLE="${ROLES[$i]}" NODE_PORT="${PORTS[$i]}" nohup ./horizon-angel.exe > logs/angel-${PORTS[$i]}.log 2>&1 &
done
AGENT_NODE_PORT=9008 nohup ./horizon-agent.exe > logs/agent.log 2>&1 &
sleep 3

echo "⏳ [3/4] Checking Node Status (Row by Row)..."
printf "%-10s %-6s %-10s\n" "Node" "Port" "Status"
printf "%-10s %-6s %-10s\n" "----" "----" "------"
for i in 0 1 2 3 4 5 6; do
  port="${PORTS[$i]}"
  role="${ROLES[$i]}"
  code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 2 http://localhost:$port/health)
  if [ "$code" = "200" ]; then status="✅ ONLINE"; else status="❌ OFFLINE ($code)"; fi
  printf "%-10s %-6s %-10s\n" "$role" "$port" "$status"
done
code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 2 http://localhost:9008/health)
if [ "$code" = "200" ]; then status="✅ ONLINE"; else status="❌ OFFLINE ($code)"; fi
printf "%-10s %-6s %-10s\n" "Agent" "9008" "$status"

echo ""
echo "⏳ [4/4] Running Consensus Test (8/8 Vote)..."
psql -U postgres -h localhost -d horizon_switch -c "DELETE FROM transactions WHERE synced = false;" > /dev/null
PRIV=$(sed 's/^0[xX]//;s/[[:space:]]//g' /d/mykey.txt)
curl -s -X POST http://localhost:8091/api/v1/key/setup -H "Content-Type: application/json" -H "$H" -d "{\"private_key\":\"$PRIV\",\"password\":\"authority\"}" > /dev/null
curl -s -X POST http://localhost:8091/api/v1/key/unlock -H "Content-Type: application/json" -H "$H" -d '{"password":"authority"}' > /dev/null

LIC_ID="CONSENSUS-TEST-$(date +%s)"
curl -s -X POST http://localhost:8091/api/v1/admin/license/issue -H "Content-Type: application/json" -H "$H" -d "{\"license_id\":\"$LIC_ID\",\"user_id\":\"sender_1\",\"product_id\":\"horizon-core\",\"volume\":1000000,\"duration\":365}" > /dev/null

LICENSE_ID="$LIC_ID" ./loadtest-bin.exe raw 100 > /dev/null

echo "Mining Result:"
curl -s -X POST http://localhost:8091/api/v1/block/mine -H "Content-Type: application/json" -H "$H" -d '{}'
echo ""
