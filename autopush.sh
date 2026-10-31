#!/usr/bin/env bash
set -euo pipefail
cd "$HOME/Beaconchain/horizon-core"

# Ensure .gitignore is correct (exclude binaries and db)
cat > .gitignore << 'GITIGNORE'
bin/
*.exe
*.db
*.db-shm
*.db-wal
*.log
node_modules/
frontend/
GITIGNORE

# Add all source files
git add -A

# Commit with a timestamped message
MSG="Auto-push: $(date '+%Y-%m-%d %H:%M:%S')"
git commit -m "$MSG" --allow-empty

# Push to GitHub
echo "Pushing to GitHub..."
git push -u origin main --force

echo "✅ Auto-push complete: $MSG"
