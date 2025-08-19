#!/bin/bash

# Deploy script for Zerops MCP HTTP server

echo "Building for Linux x86_64..."
GOOS=linux GOARCH=amd64 go build -o zerops-mcp-linux ./cmd/mcp-server/main.go

echo "Deploying to server..."
scp zerops-mcp-linux mcp.zerops:/var/www/zerops-mcp

echo "Restarting service..."
ssh mcp.zerops "pkill -f zerops-mcp; cd /var/www && nohup ./zerops-mcp --transport http --host 0.0.0.0 --port 8080 > app.log 2>&1 &"

echo "Waiting for service to start..."
sleep 2

echo "Checking status..."
ssh mcp.zerops "ps aux | grep zerops-mcp | grep -v grep"

echo "Testing health endpoint..."
curl -s https://mcp-16cb-8080.prg1.zerops.app/health | jq '.'

echo "Deployment complete!"