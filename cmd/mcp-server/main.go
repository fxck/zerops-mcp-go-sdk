// Zerops MCP Server - Supports both stdio and HTTP transports with shared tool logic
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zerops-mcp-basic/internal/handlers"
	"github.com/zerops-mcp-basic/internal/instructions"
	"github.com/zerops-mcp-basic/internal/transport"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
)

const (
	serverName    = "zerops-mcp"
	serverVersion = "1.0.0"
	apiEndpoint   = "https://api.app-prg1.zerops.io"
)

func main() {
	// Parse command-line flags
	var (
		transportMode = flag.String("transport", getEnvOrDefault("MCP_TRANSPORT", "stdio"), "Transport mode: stdio or http")
		httpHost      = flag.String("host", getEnvOrDefault("MCP_HTTP_HOST", "0.0.0.0"), "HTTP server host (http mode only)")
		httpPort      = flag.String("port", getEnvOrDefault("MCP_HTTP_PORT", "8080"), "HTTP server port (http mode only)")
	)
	flag.Parse()

	// Initialize shared tool registry first
	handlers.RegisterShared()

	// Create MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    serverName,
			Version: serverVersion,
		},
		&mcp.ServerOptions{
			Instructions: instructions.GetWorkflowInstructions(),
		},
	)

	// Handle transport-specific setup
	var client *sdk.Handler
	if *transportMode == "stdio" {
		// Stdio mode: API key from environment
		apiKey := os.Getenv("ZEROPS_API_KEY")
		if apiKey == "" {
			log.Fatal("ZEROPS_API_KEY environment variable is required for stdio mode")
		}
		client = createZeropsClient(apiKey)

		// Register tools with MCP server for stdio
		if err := handlers.RegisterForMCP(server, client); err != nil {
			log.Fatalf("Failed to register handlers: %v", err)
		}
	} else if *transportMode == "http" {
		// HTTP mode: API key will come from client requests
		log.Println("HTTP mode: API keys will be provided by clients via Authorization header")
		// No need to register with MCP server - HTTP will use shared registry directly
	} else {
		log.Fatalf("Invalid transport mode: %s (must be 'stdio' or 'http')", *transportMode)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nShutting down...")
		cancel()
	}()

	// Start server based on transport mode
	switch *transportMode {
	case "stdio":
		startStdioServer(ctx, server)
	case "http":
		startHTTPServer(ctx, server, *httpHost, *httpPort)
	}
}

func startStdioServer(ctx context.Context, server *mcp.Server) {
	fmt.Fprintf(os.Stderr, "Starting %s v%s in stdio mode...\n", serverName, serverVersion)

	stdioTransport := mcp.NewStdioTransport()
	if err := server.Run(ctx, stdioTransport); err != nil {
		if err != context.Canceled {
			log.Fatalf("Stdio server error: %v", err)
		}
	}
}

func startHTTPServer(ctx context.Context, server *mcp.Server, host, port string) {
	fmt.Fprintf(os.Stderr, "Starting %s v%s in HTTP mode on %s:%s...\n", serverName, serverVersion, host, port)
	fmt.Fprintf(os.Stderr, "Authentication: Bearer token with ZEROPS_API_KEY\n")

	config := transport.HTTPServerConfig{
		Host:   host,
		Port:   port,
		Server: server,
	}

	// Use the shared handler for HTTP
	if err := transport.StartHTTPServerShared(ctx, config); err != nil {
		if err != http.ErrServerClosed && err != context.Canceled {
			log.Fatalf("HTTP server error: %v", err)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createZeropsClient(apiKey string) *sdk.Handler {
	config := sdkBase.Config{
		Endpoint: apiEndpoint,
	}

	baseSDK := sdk.New(config, http.DefaultClient)
	authorizedSDK := sdk.AuthorizeSdk(baseSDK, apiKey)

	return &authorizedSDK
}
