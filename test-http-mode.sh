#!/bin/bash

# Test script for HTTP mode of Zerops MCP server

echo "Testing Zerops MCP Server in HTTP mode"
echo "======================================="

# Set the test token (using the provided test token)
export ZEROPS_API_KEY="SbbWs0jmQyeElIA0T9qUxwfaxS351pSoahao9HneAPXg-p"

# Test 1: Health check endpoint (no auth required)
echo ""
echo "Test 1: Health check endpoint"
curl -s http://localhost:8080/health | jq '.' || echo "Health check failed"

# Test 2: MCP endpoint without authentication (should fail)
echo ""
echo "Test 2: MCP endpoint without auth (should return 401)"
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n"

# Test 3: MCP endpoint with authentication
echo ""
echo "Test 3: MCP endpoint with valid auth"
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ZEROPS_API_KEY" \
  -H "Accept: application/json, text/event-stream" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}}},"id":1}' \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "Testing complete!"