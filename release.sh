#!/bin/bash

# Release script for Zerops MCP Server
set -e

VERSION="${1:-1.0.0}"

echo "Building Zerops MCP Server v${VERSION}..."

# Create releases directory
mkdir -p releases

# Build for all platforms
echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -o releases/zerops-mcp-win-x64.exe -ldflags="-X main.serverVersion=${VERSION}" cmd/mcp-server/main.go

echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o releases/zerops-mcp-linux-amd64 -ldflags="-X main.serverVersion=${VERSION}" cmd/mcp-server/main.go

echo "Building for Linux 386..."
GOOS=linux GOARCH=386 go build -o releases/zerops-mcp-linux-i386 -ldflags="-X main.serverVersion=${VERSION}" cmd/mcp-server/main.go

echo "Building for macOS Intel..."
GOOS=darwin GOARCH=amd64 go build -o releases/zerops-mcp-darwin-amd64 -ldflags="-X main.serverVersion=${VERSION}" cmd/mcp-server/main.go

echo "Building for macOS Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build -o releases/zerops-mcp-darwin-arm64 -ldflags="-X main.serverVersion=${VERSION}" cmd/mcp-server/main.go

# Create release archives
echo "Creating release archives..."
cd releases

for file in *; do
    if [[ "$file" == *.exe ]]; then
        zip "${file%.exe}-v${VERSION}.zip" "$file"
    else
        tar czf "${file}-v${VERSION}.tar.gz" "$file"
    fi
done

cd ..

echo "Release v${VERSION} complete!"
echo "Binaries available in releases/ directory"
ls -lh releases/