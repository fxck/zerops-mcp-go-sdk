package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	// Parse JSON-RPC request
	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
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

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processRequest handles JSON-RPC requests using shared registry
func (h *HTTPHandler) processRequest(ctx context.Context, request map[string]interface{}) map[string]interface{} {
	method, _ := request["method"].(string)
	id := request["id"]
	params, _ := request["params"].(map[string]interface{})

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
				"instructions": getHTTPInstructions(),
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

// getHTTPInstructions returns instructions for HTTP mode
func getHTTPInstructions() string {
	return `
# Zerops MCP - HTTP Mode

## CRITICAL: Always Use Knowledge Base First
Before creating services, ALWAYS search the knowledge base:
1. knowledge_search("service_type") - Find available services and recipes
2. knowledge_get("services/mongodb") - Get exact configuration details
3. knowledge_get("recipe/laravel") - Get complete working templates

## Service Import Workflow
1. Search KB for service types: knowledge_search("mongodb") 
2. Get exact type info: knowledge_get("services/mongodb")
3. Use the EXACT type string from KB (e.g., "mongodb@7" not "mongodb@7.0")
4. Hostname must be alphanumeric only (no hyphens)
5. For utility services (Adminer, Mailpit, S3Browser), KEEP buildFromGit field!

## Common Service Types (always verify with KB first)
- postgresql@16, mariadb@11, mongodb@7
- nodejs@20, python@3.11, php@8.3 (NOT php-apache!)
- valkey@7, keydb@6, elasticsearch@8

## Error Recovery
If you get "serviceStackTypeNotFound":
1. Use knowledge_search to find correct type
2. Check hostname has no special characters
3. Verify mode is "HA" or "NON_HA" (for databases)

Remember: Knowledge base has 159+ working recipes - use them!`
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
