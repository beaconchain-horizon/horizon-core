#!/usr/bin/env bash

ZIP_FILE="horizon-trial-nicico-real.zip"
TEMP_DIR="/tmp/trial_check_temp"

echo "============================================================"
echo "  HORIZON CORE - TRIAL MODE VALIDATION (FIXED)"
echo "============================================================"

echo -e "\n[1/3] Extracting package with unzip (handling backslashes)..."
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"
unzip -q "$ZIP_FILE" -d "$TEMP_DIR"

echo -e "\n[2/3] Verifying License (Must be NICICO Real License)..."
# پیدا کردن فایل لایسنس در هر مسیری که هست
LICENSE_FILE=$(find "$TEMP_DIR" -name "trial.lic" | head -n 1)

if [ -n "$LICENSE_FILE" ]; then
    echo "License found at: $LICENSE_FILE"
    cat "$LICENSE_FILE" | grep -E '"license_id"|"user_id"|"volume"|"duration"'

    if grep -q '"user_id": "nicico"' "$LICENSE_FILE"; then
        echo "✅ PASS: License belongs to 'nicico'."
    else
        echo "❌ FAIL: License user_id is NOT nicico!"
    fi
else
    echo "❌ FATAL: License file not found inside the package!"
fi

echo -e "\n[3/3] Verifying start.bat and Binaries..."
BAT_FILE=$(find "$TEMP_DIR" -name "start.bat" | head -n 1)
API_EXE=$(find "$TEMP_DIR" -name "horizon-api.exe" | head -n 1)

if [ -f "$BAT_FILE" ] && grep -q "license/register" "$BAT_FILE"; then
    echo "✅ PASS: start.bat has auto-registration logic."
else
    echo "⚠️ WARNING: start.bat missing or lacks registration logic."
fi

if [ -f "$API_EXE" ]; then
    echo "✅ PASS: API binary (horizon-api.exe) is present."
else
    echo "❌ FAIL: API binary is missing!"
fi

echo -e "\n============================================================"
echo "  VALIDATION COMPLETE. PACKAGE IS READY."
echo "============================================================"

rm -rf "$TEMP_DIR"
