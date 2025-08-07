package handlers

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zerops-mcp-basic/internal/handlers/tools"
	"github.com/zeropsio/zerops-go/sdk"
)

// Register registers all MCP handlers with the server
func Register(server *mcp.Server, client *sdk.Handler) error {
	// Register authentication tools
	tools.RegisterAuth(server, client)
	
	// Register project management tools
	tools.RegisterProjects(server, client)
	
	// Register service management tools
	tools.RegisterServices(server, client)
	
	// Register deployment tools
	tools.RegisterDeploy(server, client)
	
	// Register knowledge base tools
	tools.RegisterKnowledge(server, client)
	
	return nil
}