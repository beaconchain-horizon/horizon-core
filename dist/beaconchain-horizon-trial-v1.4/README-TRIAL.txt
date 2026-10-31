BEACONCHAIN HORIZON — TRIAL PACKAGE
====================================
WHAT IS THIS
An offline-first tamper-evident ledger. Trial is valid for 48
hours from first launch, maximum 1,000,000 transactions.
HOW TO START
1. Extract this folder anywhere
2. Double-click START-TRIAL.bat
3. Wait 15 seconds
4. Verify: curl -s localhost:8080/api/v1/health
INDEPENDENT TESTS
  bash tps-report.sh
  ./pentest.exe -url http://localhost:8080
AFTER 48 HOURS
The backend will refuse to start. To continue, contact the
maintainers for a production license bound to your hardware.
(c) 2026 Beaconchain Horizon
