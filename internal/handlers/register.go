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
	// Register simplified MCP tool handlers
	tools.RegisterDiscovery()        // discovery tool
	tools.RegisterServiceTools()     // get_service_types, import_services, enable_preview_subdomain, scale_service, get_service_logs
	tools.RegisterEnvironment()      // set_project_env, set_service_env
	tools.RegisterProcesses()        // get_running_processes
	tools.RegisterKnowledgeBase()    // knowledge_base
}

// RegisterForMCP registers all tools with the MCP server for stdio transport
// It uses the shared registry to get tool definitions
func RegisterForMCP(server *mcp.Server, client *sdk.Handler) error {
	return RegisterForMCPWithClientInfo(server, client, nil)
}

// RegisterForMCPWithClientInfo registers all tools with client info support
func RegisterForMCPWithClientInfo(server *mcp.Server, client *sdk.Handler, clientInfo **mcp.Implementation) error {
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
			
			// Add client info to context if available
			if clientInfo != nil && *clientInfo != nil {
				ctx = context.WithValue(ctx, "clientName", (*clientInfo).Name)
				ctx = context.WithValue(ctx, "clientVersion", (*clientInfo).Version)
			}

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
