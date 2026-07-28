param(
    [string]$Version = "0.1.0"
)

$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Force -Path "dist" | Out-Null

Write-Host "Building linux/amd64..."
$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -ldflags "-s -w -X main.Version=$Version" -o dist/p-mon-agent-linux-amd64 ./cmd/agent

Write-Host "Building linux/arm64..."
$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"
go build -ldflags "-s -w -X main.Version=$Version" -o dist/p-mon-agent-linux-arm64 ./cmd/agent

Write-Host "Building windows/amd64..."
$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -ldflags "-s -w -X main.Version=$Version" -o dist/p-mon-agent-windows-amd64.exe ./cmd/agent

Write-Host "Done. Binaries in dist/"
