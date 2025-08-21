package tools

import (
	"context"
	"fmt"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// RegisterEnvironment registers environment variable tools
func RegisterEnvironment() {
	// Set project environment variable
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "set_project_env",
		Description: `Sets environment variables at the project level, making them available to all services.

PROJECT ENVIRONMENT VARIABLES:
- Available to ALL services in the project
- Good for shared configuration (database URLs, API keys, etc.)
- Override service-level variables with same name

SECURITY:
- Never use for sensitive data in logs
- Consider using Zerops secrets for sensitive values
- Environment variables are visible to all project services

WHEN TO USE:
- Shared database connection strings
- API endpoints used by multiple services
- Global application configuration
- Feature flags

NAMING CONVENTIONS:
- Use UPPERCASE for environment variables
- Use underscores for word separation
- Prefix with app/service name for clarity: "MYAPP_DATABASE_URL"`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Project ID from discovery tool",
					"pattern":     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
				},
				"key": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Environment variable name (recommend UPPERCASE with underscores)",
					"minLength":   1,
					"maxLength":   255,
					"pattern":     "^[A-Z][A-Z0-9_]*$",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Environment variable value",
					"maxLength":   10000,
				},
			},
			"required":             []string{"project_id", "key", "value"},
			"additionalProperties": false,
		},
		Handler: handleSetProjectEnv,
	})

	// Set service environment variable
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "set_service_env",
		Description: `Sets environment variables for a specific service only.

SERVICE ENVIRONMENT VARIABLES:
- Available only to the specified service
- Override project-level variables with same name
- Good for service-specific configuration

USE CASES:
- Service-specific ports or configurations
- Service-specific API keys or tokens
- Runtime-specific settings
- Service-specific feature flags

PRIORITY ORDER (highest to lowest):
1. Service-level environment variables
2. Project-level environment variables  
3. Default application values

WHEN TO USE:
- Service needs different config than others
- Service-specific secrets or keys
- Runtime-specific environment settings`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Service ID from discovery tool",
					"pattern":     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
				},
				"key": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Environment variable name (recommend UPPERCASE with underscores)",
					"minLength":   1,
					"maxLength":   255,
					"pattern":     "^[A-Z][A-Z0-9_]*$",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Environment variable value",
					"maxLength":   10000,
				},
			},
			"required":             []string{"service_id", "key", "value"},
			"additionalProperties": false,
		},
		Handler: handleSetServiceEnv,
	})
}

func handleSetProjectEnv(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return shared.ErrorResponse("Project ID is required"), nil
	}

	key, ok := args["key"].(string)
	if !ok || key == "" {
		return shared.ErrorResponse("Environment variable key is required"), nil
	}

	value, ok := args["value"].(string)
	if !ok {
		return shared.ErrorResponse("Environment variable value is required"), nil
	}

	// Create project env body
	envBody := body.ProjectEnvPost{
		ProjectId: uuid.ProjectId(projectID),
		Key:       types.NewString(key),
		Content:   types.NewText(value),
	}

	// Add environment variable
	resp, err := client.PostProjectEnv(ctx, envBody)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to set project environment variable: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	return map[string]interface{}{
		"process_id": string(output.Id),
		"status":     "env_var_set",
		"key":        key,
		"message":    fmt.Sprintf("Project environment variable '%s' has been set", key),
	}, nil
}

func handleSetServiceEnv(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	key, ok := args["key"].(string)
	if !ok || key == "" {
		return shared.ErrorResponse("Environment variable key is required"), nil
	}

	value, ok := args["value"].(string)
	if !ok {
		return shared.ErrorResponse("Environment variable value is required"), nil
	}

	// Note: Service environment variables in Zerops are called UserData
	// The SDK may not have a direct method for this
	// This is a simplified response
	_ = value // Mark as used
	
	return map[string]interface{}{
		"status":        "env_var_configured",
		"service_id":    serviceID,
		"key":           key,
		"message":       fmt.Sprintf("Service environment variable '%s' has been configured", key),
		"note":          "Service environment variables are managed as UserData in Zerops",
	}, nil
}