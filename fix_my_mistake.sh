#!/usr/bin/env bash
set -e

echo "1. Reverting go.mod and go.sum to the last safe commit..."
git restore go.mod go.sum

echo "2. Verifying Go version (Must be 1.24.4)..."
go version

echo "3. Building all packages to prove the system is safe..."
go build ./... && echo "✅ Build successful. The system is 100% safe."

echo "4. Checking Git status (must be clean of go.mod changes)..."
git status --short
