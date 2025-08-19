package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DirectHTTPHandler creates a simpler HTTP handler that processes JSON-RPC directly
func DirectHTTPHandler(server *mcp.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS preflight
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Only accept POST requests for JSON-RPC
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
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

		// Log the request for debugging
		log.Printf("Received JSON-RPC request: %s", string(body))

		// Check if client accepts SSE
		acceptHeader := r.Header.Get("Accept")
		useSSE := strings.Contains(acceptHeader, "text/event-stream")

		if useSSE {
			// Set up SSE response
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			// Create a simple SSE response
			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
				return
			}

			// Process the request and send response as SSE
			response := processRequest(server, request)
			responseJSON, _ := json.Marshal(response)
			
			// Send as SSE event
			fmt.Fprintf(w, "data: %s\n\n", responseJSON)
			flusher.Flush()

			// Keep connection open briefly for additional events if needed
			time.Sleep(100 * time.Millisecond)
		} else {
			// Regular JSON response
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			response := processRequest(server, request)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("Failed to encode response: %v", err)
			}
		}
	}
}

// processRequest handles a JSON-RPC request
func processRequest(server *mcp.Server, request map[string]interface{}) map[string]interface{} {
	method, _ := request["method"].(string)
	id := request["id"]
	params := request["params"]

	// Handle different methods
	switch method {
	case "initialize":
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
					"resources": map[string]interface{}{
						"subscribe": false,
						"listChanged": false,
					},
				},
				"serverInfo": map[string]interface{}{
					"name":    "zerops-mcp",
					"version": "1.0.0",
				},
			},
		}

	case "tools/list":
		// Get tools from the server
		// This is a simplified response - in production, you'd enumerate actual tools
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]interface{}{
				"tools": []interface{}{
					map[string]interface{}{
						"name":        "project_list",
						"description": "List all projects",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{},
						},
					},
					map[string]interface{}{
						"name":        "service_list",
						"description": "List services in a project",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"projectId": map[string]interface{}{
									"type":        "string",
									"description": "Project ID",
								},
							},
							"required": []string{"projectId"},
						},
					},
				},
			},
		}

	case "tools/call":
		// Handle tool calls
		toolName, _ := params.(map[string]interface{})["name"].(string)
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": fmt.Sprintf("Tool %s called (simplified response)", toolName),
					},
				},
			},
		}

	default:
		// Unknown method
		return map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"error": map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			},
		}
	}
}