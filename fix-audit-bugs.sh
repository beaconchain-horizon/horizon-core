#!/usr/bin/env bash
set -e
cd ~/Beaconchain/horizon-core

echo "1. Fixing Transaction Limit Bug (chain_handlers.go)..."
# 1. Delete the old check (which was after DB save)
awk '
/if trialTxLimitReached\(\) \{/ { skip=1 }
/if err := chain\.Enqueue\(chain\.PendingTx\{/ { skip=0 }
!skip { print }
' cmd/api/chain_handlers.go > /tmp/ch1.go

# 2. Insert the check BEFORE saving the license
awk '
/\/\/ Persist the license\./ {
  print "\tif trialTxLimitReached() {"
  print "\t\tc.JSON(http.StatusForbidden, gin.H{\"error\": \"TRIAL: transaction limit reached (1,000,000)\"})"
  print "\t\treturn"
  print "\t}"
  print ""
}
{ print }
' /tmp/ch1.go > cmd/api/chain_handlers.go

echo "2. Fixing Trial Timer Bypass (trial.go)..."
# Hide trial-state.json in AppData/Roaming/HorizonCore
awk '
/dir := os.Getenv\("HORIZON_TRIAL_DIR"\)/ {
  print "\tdir, _ := os.UserConfigDir()"
  print "\tif dir == \"\" {"
  print "\t\tdir = os.Getenv(\"HORIZON_TRIAL_DIR\")"
  print "\t}"
  print "\tif dir == \"\" {"
  print "\t\tdir = \".\""
  print "\t}"
  print "\tdir = filepath.Join(dir, \"HorizonCore\")"
  print "\tos.MkdirAll(dir, 0755)"
  next
}
{ print }
' cmd/api/trial.go > /tmp/tr.go && mv /tmp/tr.go cmd/api/trial.go

echo "3. Fixing Hardcoded Admin Token (build-trial-package.sh)..."
# Generate a random token at build time and inject it into the BAT file
awk '
/# ZIP/ {
  print "RANDOM_TOKEN=$(openssl rand -hex 16)"
  print "sed -i \"s|set ADMIN_TOKEN=12345678901234567890123456789012|set ADMIN_TOKEN=$RANDOM_TOKEN|\" \"$OUT/START-TRIAL.bat\""
  print ""
}
{ print }
' build-trial-package.sh > /tmp/bp.sh && mv /tmp/bp.sh build-trial-package.sh

echo "4. Rebuilding backend..."
go build -o backend.exe ./cmd/api/

echo "5. Repackaging Trial..."
rm -rf dist/beaconchain-horizon-trial-v1.4
rm -f dist/beaconchain-horizon-trial-v1.4.zip
bash build-trial-package.sh beaconchain-horizon-trial-v1.4

echo "6. Creating final ZIP..."
cd dist
powershell -Command "Compress-Archive -Path 'beaconchain-horizon-trial-v1.4' -DestinationPath 'beaconchain-horizon-trial-v1.4.zip' -Force"
echo "New SHA-256 hash:"
sha256sum beaconchain-horizon-trial-v1.4.zip
