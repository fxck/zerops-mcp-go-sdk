package tools

import (
	"fmt"
	
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// EmptyArgs represents tools with no parameters
type EmptyArgs struct{}

// textResult creates a simple text response for MCP
func textResult(message string) *mcp.CallToolResultFor[struct{}] {
	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}
}

// errorResult creates an error response for MCP
func errorResult(err error) *mcp.CallToolResultFor[struct{}] {
	return textResult(fmt.Sprintf("Error: %v", err))
}