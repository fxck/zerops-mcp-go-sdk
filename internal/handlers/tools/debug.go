package tools

import (
	"context"
	"fmt"
	"runtime"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/sdk"
)

// RegisterDebug registers debug tools in the global registry
func RegisterDebug() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "debug_info",
		Description: "Display debug information including client identification, server version, and runtime details",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: handleDebugInfo,
	})
}

func handleDebugInfo(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	var message string

	// Client information from context
	clientName := ctx.Value("clientName")
	clientVersion := ctx.Value("clientVersion")
	
	message += "=== CLIENT INFORMATION ===\n"
	if clientName != nil && clientVersion != nil {
		message += fmt.Sprintf("Client: %s\n", clientName)
		message += fmt.Sprintf("Version: %s\n", clientVersion)
		message += fmt.Sprintf("Full ID: %s v%s\n", clientName, clientVersion)
	} else {
		message += "Client: Unknown (no clientInfo sent)\n"
		message += "Note: Client info is only available in initialize request\n"
	}

	// Transport mode
	message += "\n=== TRANSPORT MODE ===\n"
	if ctx.Value("httpMode") != nil {
		message += "Mode: HTTP (remote)\n"
		message += "Auth: Per-request Bearer token\n"
	} else {
		message += "Mode: stdio (local)\n"
		message += "Auth: Environment variable\n"
	}

	// Server information
	message += "\n=== SERVER INFORMATION ===\n"
	message += "Server: zerops-mcp\n"
	message += "Version: 1.0.0\n"
	message += fmt.Sprintf("Go Version: %s\n", runtime.Version())
	message += fmt.Sprintf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// API connection
	message += "\n=== API CONNECTION ===\n"
	if client != nil {
		message += "Zerops API: Connected\n"
		message += "Endpoint: https://api.app-prg1.zerops.io\n"
	} else {
		message += "Zerops API: Not initialized\n"
	}

	// Known client signatures
	message += "\n=== KNOWN CLIENT SIGNATURES ===\n"
	message += "• claude-desktop: Claude Desktop/Code application\n"
	message += "• ragflow-mcp-client: RAGFlow MCP client\n"
	message += "• spring-ai-mcp: Spring AI MCP integration\n"
	message += "• mcp-client: Generic MCP client\n"

	// Interpretation
	message += "\n=== CLIENT INTERPRETATION ===\n"
	if clientName != nil {
		name := clientName.(string)
		switch {
		case name == "claude-desktop" || name == "claude-code":
			message += "You are being called by: Claude (Anthropic)\n"
		case name == "chatgpt-desktop" || name == "openai-mcp":
			message += "You are being called by: ChatGPT (OpenAI)\n"
		case name == "ragflow-mcp-client":
			message += "You are being called by: RAGFlow\n"
		default:
			message += fmt.Sprintf("You are being called by: Unknown client (%s)\n", name)
		}
	} else {
		message += "Unable to determine calling client\n"
	}

	return shared.TextResponse(message), nil
}