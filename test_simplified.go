package main

import (
	"context"
	"fmt"
	"log"

	"github.com/zerops-mcp-basic/internal/handlers"
	"github.com/zerops-mcp-basic/internal/handlers/shared"
)

func main() {
	// Initialize the registry
	handlers.InitializeRegistry()

	// List all registered tools
	tools := shared.GlobalRegistry.List()
	
	fmt.Printf("Simplified MCP Tools (%d registered):\n\n", len(tools))
	
	for _, tool := range tools {
		fmt.Printf("• %s\n", tool.Name)
		fmt.Printf("  %s\n\n", tool.Description)
	}

	// Test tool execution (without API key)
	fmt.Println("Testing tools without API key:\n")
	
	// Test discovery
	testTool("discovery", map[string]interface{}{})
	
	// Test get_service_types
	testTool("get_service_types", map[string]interface{}{})
	
	// Test knowledge_base
	testTool("knowledge_base", map[string]interface{}{
		"runtime": "nodejs",
	})
}

func testTool(name string, args map[string]interface{}) {
	tool, found := shared.GlobalRegistry.Get(name)
	if !found || tool == nil {
		log.Printf("Tool %s not found\n", name)
		return
	}
	
	fmt.Printf("Testing %s:\n", name)
	result, err := tool.Handler(context.Background(), nil, args)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Result: %v\n", result)
	}
	fmt.Println()
}