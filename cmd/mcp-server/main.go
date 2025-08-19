// Zerops MCP Server - Supports both stdio and HTTP transports
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
		transportMode   = flag.String("transport", getEnvOrDefault("MCP_TRANSPORT", "stdio"), "Transport mode: stdio or http")
		httpHost       = flag.String("host", getEnvOrDefault("MCP_HTTP_HOST", "0.0.0.0"), "HTTP server host (http mode only)")
		httpPort       = flag.String("port", getEnvOrDefault("MCP_HTTP_PORT", "8080"), "HTTP server port (http mode only)")
		skipValidation = flag.Bool("skip-validation", false, "Skip API key validation (for testing only)")
	)
	flag.Parse()

	// Get API key from environment
	apiKey := os.Getenv("ZEROPS_API_KEY")
	if apiKey == "" && !*skipValidation {
		log.Fatal("ZEROPS_API_KEY environment variable is required")
	}

	// Create Zerops SDK client (can be nil if skip-validation is used)
	var client *sdk.Handler
	if apiKey != "" {
		client = createZeropsClient(apiKey)
	} else if *skipValidation {
		log.Println("WARNING: Running without ZEROPS_API_KEY - API calls will fail")
		// Create a dummy client for testing
		client = nil
	}

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
	default:
		log.Fatalf("Invalid transport mode: %s (must be 'stdio' or 'http')", *transportMode)
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
		Host:      host,
		Port:      port,
		Server:    server,
	}

	if err := transport.StartHTTPServer(ctx, config); err != nil {
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