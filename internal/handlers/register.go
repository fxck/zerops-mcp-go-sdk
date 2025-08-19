package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zerops-mcp-basic/internal/handlers/tools"
	"github.com/zeropsio/zerops-go/sdk"
)

// InitializeRegistry initializes the global tool registry
// This should be called at startup before any transport is initialized
func InitializeRegistry() {
	// Register all tool handlers in the global registry
	tools.RegisterAuth()
	tools.RegisterProjects()
	tools.RegisterServices()
	tools.RegisterDeploy()
	tools.RegisterKnowledge()
}

// RegisterForMCP registers all tools with the MCP server for stdio transport
// It uses the shared registry to get tool definitions
func RegisterForMCP(server *mcp.Server, client *sdk.Handler) error {
	// Get all tools from the shared registry
	toolDefs := shared.GlobalRegistry.List()

	// Register each tool with the MCP server
	for _, toolDef := range toolDefs {
		// Create a closure to capture the tool definition
		td := toolDef

		// Create MCP tool
		mcpTool := &mcp.Tool{
			Name:        td.Name,
			Description: td.Description,
		}

		// Create handler that bridges to shared handler
		handler := mcp.ToolHandler(func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResultFor[any], error) {
			// Extract arguments from params
			args := params.Arguments

			// Add client to context if available
			if client != nil {
				ctx = context.WithValue(ctx, "zeropsClient", client)
			}
			
			// Note: Client info (name/version) is available during initialization
			// but not accessible here in tool handlers through the session

			// Call the shared handler
			result, err := td.Handler(ctx, client, args)
			if err != nil {
				// Return error as MCP result
				return &mcp.CallToolResultFor[any]{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
					},
					IsError: true,
				}, nil
			}

			// Convert result to MCP format
			if mcpResult, ok := result.(map[string]interface{}); ok {
				if contentArr, ok := mcpResult["content"].([]interface{}); ok {
					var content []mcp.Content
					for _, item := range contentArr {
						if textItem, ok := item.(map[string]interface{}); ok {
							if textItem["type"] == "text" {
								if text, ok := textItem["text"].(string); ok {
									content = append(content, &mcp.TextContent{Text: text})
								}
							}
						}
					}
					return &mcp.CallToolResultFor[any]{
						Content: content,
						IsError: mcpResult["isError"] == true,
					}, nil
				}
			}

			// Fallback for simple results
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("%v", result)},
				},
			}, nil
		})

		// Register with MCP server
		mcp.AddTool(server, mcpTool, handler)
	}

	return nil
}
