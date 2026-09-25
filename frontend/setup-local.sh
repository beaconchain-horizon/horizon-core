#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
echo "Horizon Store + Angel SOC"
node -v
npm -v
if [ ! -f .env.local ]; then cp .env.example .env.local; echo "Created .env.local — set your Liara URLs and Angel credentials before production use."; fi
npm install
npm run dev
