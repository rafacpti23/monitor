#!/bin/bash
set -e

VERSION=${1:-0.1.0}
mkdir -p dist

echo "Building linux/amd64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=$VERSION" -o dist/p-mon-agent-linux-amd64 ./cmd/agent

echo "Building linux/arm64..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=$VERSION" -o dist/p-mon-agent-linux-arm64 ./cmd/agent

echo "Building windows/amd64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=$VERSION" -o dist/p-mon-agent-windows-amd64.exe ./cmd/agent

echo "Done. Binaries in dist/"
