#!/usr/bin/env bash
set -e

echo "1. Updating Go modules..."
go get -u ./...
go mod tidy

echo "2. Rebuilding all packages to ensure nothing broke..."
go build ./... && echo "✅ Build successful."

echo "3. Committing and pushing updated dependencies..."
git add go.mod go.sum
git commit -m "security: update dependencies to resolve Dependabot alerts"
git push origin main

echo "🚀 Done. Dependabot alerts should clear shortly."
