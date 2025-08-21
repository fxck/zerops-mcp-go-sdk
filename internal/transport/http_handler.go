package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
)

// HTTPServerConfig contains configuration for the HTTP server
type HTTPServerConfig struct {
	Host   string
	Port   string
	Server *mcp.Server
}

// HTTPHandler handles HTTP requests using the global tool registry
type HTTPHandler struct {
	mcpServer *mcp.Server
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(mcpServer *mcp.Server) *HTTPHandler {
	return &HTTPHandler{
		mcpServer: mcpServer,
	}
}

// ServeHTTP handles incoming HTTP requests using shared registry
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Log User-Agent and other headers that might contain model info
	if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
		fmt.Fprintf(os.Stderr, "User-Agent: %s\n", userAgent)
	}
	if xModel := r.Header.Get("X-Model"); xModel != "" {
		fmt.Fprintf(os.Stderr, "X-Model: %s\n", xModel)
	}
	if xClaudeModel := r.Header.Get("X-Claude-Model"); xClaudeModel != "" {
		fmt.Fprintf(os.Stderr, "X-Claude-Model: %s\n", xClaudeModel)
	}

	// Health check endpoint
	if r.URL.Path == "/health" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "healthy",
			"service":   "zerops-mcp",
			"transport": "http",
		})
		return
	}

	// Only accept POST for JSON-RPC
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract API key
	apiKey := extractBearerToken(r.Header.Get("Authorization"))
	if apiKey == "" {
		http.Error(w, "Authorization header with Bearer token required", http.StatusUnauthorized)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Log the raw request for debugging
	fmt.Fprintf(os.Stderr, "\n=== RAW REQUEST ===\n")
	fmt.Fprintf(os.Stderr, "Body: %s\n", string(body))
	fmt.Fprintf(os.Stderr, "==================\n\n")

	// Parse JSON-RPC request
	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		fmt.Fprintf(os.Stderr, "JSON Parse Error: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create context with API key and HTTP mode flag
	ctx := r.Context()
	ctx = context.WithValue(ctx, "httpMode", true) // Flag for HTTP mode

	if apiKey != "" {
		ctx = context.WithValue(ctx, "apiKey", apiKey)
		client := createZeropsClient(apiKey)
		ctx = context.WithValue(ctx, "zeropsClient", client)
	}

	// Process the request
	response := h.processRequest(ctx, request)

	// Log the response for debugging
	if responseBytes, err := json.MarshalIndent(response, "", "  "); err == nil {
		fmt.Fprintf(os.Stderr, "\n=== RESPONSE ===\n")
		fmt.Fprintf(os.Stderr, "%s\n", string(responseBytes))
		fmt.Fprintf(os.Stderr, "================\n\n")
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processRequest handles JSON-RPC requests using shared registry
func (h *HTTPHandler) processRequest(ctx context.Context, request map[string]interface{}) map[string]interface{} {
	method, _ := request["method"].(string)
	id := request["id"]
	params, _ := request["params"].(map[string]interface{})

	// Extract client information if available
	var clientName, clientVersion string
	if method == "initialize" && params != nil {
		if clientInfo, ok := params["clientInfo"].(map[string]interface{}); ok {
			clientName, _ = clientInfo["name"].(string)
			clientVersion, _ = clientInfo["version"].(string)
			clientTitle, _ := clientInfo["title"].(string)
			
			fmt.Fprintf(os.Stderr, "\n=== CLIENT IDENTIFICATION (HTTP) ===\n")
			fmt.Fprintf(os.Stderr, "Client: %s\n", clientName)
			fmt.Fprintf(os.Stderr, "Version: %s\n", clientVersion)
			if clientTitle != "" {
				fmt.Fprintf(os.Stderr, "Title: %s\n", clientTitle)
			}
			if protocol, ok := params["protocolVersion"].(string); ok {
				fmt.Fprintf(os.Stderr, "Protocol: %s\n", protocol)
			}
			
			// Check for _meta field which might contain model info
			if meta, ok := params["_meta"].(map[string]interface{}); ok {
				fmt.Fprintf(os.Stderr, "Meta fields:\n")
				for key, value := range meta {
					fmt.Fprintf(os.Stderr, "  %s: %v\n", key, value)
				}
			}
			
			// Check capabilities for any model hints
			if caps, ok := params["capabilities"].(map[string]interface{}); ok {
				// Check if there's any model-specific info in capabilities
				if len(caps) > 0 {
					fmt.Fprintf(os.Stderr, "Capabilities: %v\n", caps)
				}
			}
			
			fmt.Fprintf(os.Stderr, "===========================\n\n")
			
			// Store client info in context for use in tools
			ctx = context.WithValue(ctx, "clientName", clientName)
			ctx = context.WithValue(ctx, "clientVersion", clientVersion)
		}
	}

	switch method {
	case "initialize":
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools":     map[string]interface{}{},
					"resources": map[string]interface{}{},
					"prompts":   map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "zerops-mcp",
					"version": "1.0.0",
				},
				// No special instructions for simplified MCP
			},
		}

	case "tools/list":
		tools := h.getRegisteredTools()
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]interface{}{
				"tools": tools,
			},
		}

	case "tools/call":
		toolName, _ := params["name"].(string)
		toolArgs, _ := params["arguments"].(map[string]interface{})

		// Note: Client info was stored in context during initialize
		// but context is per-request in HTTP mode, so it's lost
		// Call tool using shared registry
		result, err := shared.GlobalRegistry.CallTool(ctx, toolName, toolArgs)
		if err != nil {
			return map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error": map[string]interface{}{
					"code":    -32603,
					"message": err.Error(),
				},
			}
		}

		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result":  result,
		}

	default:
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"error": map[string]interface{}{
				"code":    -32601,
				"message": "Method not found: " + method,
			},
		}
	}
}

// getRegisteredTools returns all tools from shared registry
func (h *HTTPHandler) getRegisteredTools() []map[string]interface{} {
	tools := shared.GlobalRegistry.List()
	result := make([]map[string]interface{}, 0, len(tools))

	for _, tool := range tools {
		result = append(result, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	return result
}

// extractBearerToken extracts the token from "Bearer <token>" format
func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// createZeropsClient creates a Zerops SDK client with the given API key
func createZeropsClient(apiKey string) *sdk.Handler {
	config := sdkBase.Config{
		Endpoint: "https://api.app-prg1.zerops.io",
	}
	baseSDK := sdk.New(config, http.DefaultClient)
	authorizedSDK := sdk.AuthorizeSdk(baseSDK, apiKey)
	return &authorizedSDK
}

// StartHTTPServer starts the HTTP server using the global registry
func StartHTTPServer(ctx context.Context, config HTTPServerConfig) error {
	handler := NewHTTPHandler(config.Server)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler: handler,
	}

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
