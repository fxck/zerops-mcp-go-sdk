package shared

import (
	"context"
	"fmt"
	"sync"

	"github.com/zeropsio/zerops-go/sdk"
)

// ToolFunc is a function that handles a tool call
type ToolFunc func(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error)

// ToolDefinition describes a tool
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
	Handler     ToolFunc
}

// ToolRegistry manages tool registrations
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*ToolDefinition
}

// GlobalRegistry is the shared tool registry
var GlobalRegistry = &ToolRegistry{
	tools: make(map[string]*ToolDefinition),
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(tool *ToolDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (*ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *ToolRegistry) List() []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// CallTool executes a tool by name
func (r *ToolRegistry) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// Get client from context (may be nil for some tools)
	client, _ := ctx.Value("zeropsClient").(*sdk.Handler)

	return tool.Handler(ctx, client, args)
}

// Helper function to create standard text response
func TextResponse(text string) interface{} {
	return map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": text,
			},
		},
	}
}

// Helper function to create error response
func ErrorResponse(message string) interface{} {
	return map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": fmt.Sprintf("❌ Error: %s", message),
			},
		},
		"isError": true,
	}
}
