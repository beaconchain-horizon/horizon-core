#!/usr/bin/env bash

ZIP_FILE="horizon-trial-nicico-real.zip"
EXPECTED_HASH="f21c1611bff65b35873f8cddf464bb4ba8603f897956d48f6bc190d62a4dfe04"
TEMP_DIR="/tmp/trial_check_temp"

echo "============================================================"
echo "  HORIZON CORE - TRIAL MODE VALIDATION"
echo "============================================================"

echo -e "\n[1/4] Checking ZIP Integrity..."
if [ ! -f "$ZIP_FILE" ]; then
    echo "❌ FATAL: $ZIP_FILE not found!"
    exit 1
fi

ACTUAL_HASH=$(sha256sum "$ZIP_FILE" | cut -d' ' -f1)
echo "Expected SHA-256: $EXPECTED_HASH"
echo "Actual SHA-256:   $ACTUAL_HASH"

if [ "$ACTUAL_HASH" == "$EXPECTED_HASH" ]; then
    echo "✅ PASS: ZIP file integrity matches."
else
    echo "❌ FAIL: ZIP hash mismatch! File may be corrupted."
    exit 1
fi

echo -e "\n[2/4] Extracting to check contents..."
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"
powershell -Command "Expand-Archive -Path '$ZIP_FILE' -DestinationPath '$TEMP_DIR' -Force"

echo "Files inside the package:"
find "$TEMP_DIR" -type f -exec ls -lh {} \; | awk '{print "  - " $9 " (" $5 ")"}'

echo -e "\n[3/4] Verifying License (Must be NICICO Real License)..."
LICENSE_FILE="$TEMP_DIR/horizon-trial-nicico-real/config/trial.lic"
if [ -f "$LICENSE_FILE" ]; then
    echo "License content:"
    cat "$LICENSE_FILE" | grep -E '"license_id"|"user_id"|"volume"|"duration"'
    
    if grep -q '"user_id": "nicico"' "$LICENSE_FILE"; then
        echo "✅ PASS: License belongs to 'nicico'."
    else
        echo "❌ FAIL: License user_id is NOT nicico!"
    fi
else
    echo "❌ FATAL: License file not found inside the package!"
fi

echo -e "\n[4/4] Verifying start.bat (Auto-registration logic)..."
BAT_FILE="$TEMP_DIR/horizon-trial-nicico-real/start.bat"
if grep -q "license/register" "$BAT_FILE" && grep -q "X-Admin-Key" "$BAT_FILE"; then
    echo "✅ PASS: start.bat contains auto-registration logic."
else
    echo "⚠️ WARNING: start.bat does not seem to auto-register the license."
fi

echo -e "\n============================================================"
echo "  VALIDATION COMPLETE."
echo "============================================================"

rm -rf "$TEMP_DIR"
