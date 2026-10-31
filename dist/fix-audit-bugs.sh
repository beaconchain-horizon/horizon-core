#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

echo "1. Fixing Transaction Limit Bug (chain_handlers.go)..."
# Move the trial check BEFORE saving the license record
awk '
/\/\/ Persist the license\./ {
  print "\t\tif trialTxLimitReached() {"
  print "\t\t\tc.JSON(http.StatusForbidden, gin.H{\"error\": \"TRIAL: transaction limit reached (1,000,000)\"})"
  print "\t\t\treturn"
  print "\t\t}"
  print ""
}
/if trialTxLimitReached\(\) \{/ { skip=1 }
/if err := chain\.Enqueue\(chain\.PendingTx\{/ { skip=0 }
!skip { print }
' cmd/api/chain_handlers.go > /tmp/ch.go && mv /tmp/ch.go cmd/api/chain_handlers.go

echo "2. Fixing Trial Timer Bypass (trial.go)..."
# Hide trial-state.json in AppData so deleting the folder doesn't reset it
sed -i 's|dir := os.Getenv("HORIZON_TRIAL_DIR")|dir, _ := os.UserConfigDir()\n\tif dir == "" {\n\t\tdir = os.Getenv("HORIZON_TRIAL_DIR")\n\t}\n\tdir = filepath.Join(dir, "HorizonCore")\n\tos.MkdirAll(dir, 0755)|' cmd/api/trial.go

echo "3. Fixing Hardcoded Admin Token (build-trial-package.sh)..."
# Generate a random token in the BAT file using PowerShell
sed -i 's|set ADMIN_TOKEN=12345678901234567890123456789012|for /f "delims=" %%a in ('"'"'powershell -Command "[guid]::NewGuid().ToString().Replace("-","")"'"'"') do set ADMIN_TOKEN=%%a|' build-trial-package.sh

echo "4. Rebuilding backend..."
go build -o backend.exe ./cmd/api/

echo "5. Repackaging Trial..."
cd dist
rm -rf beaconchain-horizon-trial-v1.4
rm -f beaconchain-horizon-trial-v1.4.zip
cd ..
bash build-trial-package.sh beaconchain-horizon-trial-v1.4

echo "6. Creating final ZIP..."
cd dist
powershell -Command "Compress-Archive -Path 'beaconchain-horizon-trial-v1.4' -DestinationPath 'beaconchain-horizon-trial-v1.4.zip' -Force"
echo "New SHA-256:"
sha256sum beaconchain-horizon-trial-v1.4.zip
