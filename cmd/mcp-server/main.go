// Zerops MCP Server - Simple, clean implementation
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zerops-mcp-basic/internal/handlers"
	"github.com/zerops-mcp-basic/internal/instructions"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
)

const (
	serverName    = "zerops-mcp"
	serverVersion = "1.0.0"
	apiEndpoint   = "https://api.app-prg1.zerops.io"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ZEROPS_API_KEY")
	if apiKey == "" {
		log.Fatal("ZEROPS_API_KEY environment variable is required")
	}

	// Create Zerops SDK client
	client := createZeropsClient(apiKey)

	// Create and configure MCP server with workflow instructions
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    serverName,
			Version: serverVersion,
		},
		&mcp.ServerOptions{
			Instructions: instructions.GetWorkflowInstructions(),
		},
	)

	// Register all handlers
	if err := handlers.Register(server, client); err != nil {
		log.Fatalf("Failed to register handlers: %v", err)
	}

	// Start server
	fmt.Fprintf(os.Stderr, "Starting %s v%s...\n", serverName, serverVersion)
	
	transport := mcp.NewStdioTransport()
	if err := server.Run(context.Background(), transport); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func createZeropsClient(apiKey string) *sdk.Handler {
	config := sdkBase.Config{
		Endpoint: apiEndpoint,
	}
	
	baseSDK := sdk.New(config, http.DefaultClient)
	authorizedSDK := sdk.AuthorizeSdk(baseSDK, apiKey)
	
	return &authorizedSDK
}