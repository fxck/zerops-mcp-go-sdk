#!/bin/bash
# Create a release with all binaries

set -e

VERSION=${1:-"v0.1.0"}

echo "Creating release $VERSION..."

# Clean and build
make clean
make all

# Create releases directory
mkdir -p releases

# Copy binaries to releases with version
cp bin/zerops-mcp-darwin-amd64 "releases/zerops-mcp-darwin-amd64"
cp bin/zerops-mcp-darwin-arm64 "releases/zerops-mcp-darwin-arm64"
cp bin/zerops-mcp-linux-amd64 "releases/zerops-mcp-linux-amd64"
cp bin/zerops-mcp-linux-i386 "releases/zerops-mcp-linux-i386"
cp bin/zerops-mcp-win-x64.exe "releases/zerops-mcp-win-x64.exe"

echo "Release binaries created in releases/"
echo ""
echo "To create a GitHub release:"
echo "1. git tag $VERSION"
echo "2. git push origin $VERSION"
echo "3. Go to https://github.com/krls2020/zerops-mcp-go-sdk/releases/new"
echo "4. Select tag $VERSION"
echo "5. Upload files from releases/ directory"
echo ""
echo "Or use GitHub CLI:"
echo "gh release create $VERSION releases/* --title 'Release $VERSION' --notes 'Release notes here'"