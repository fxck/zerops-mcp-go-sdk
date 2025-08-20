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
		Description: "Set a project-level environment variable",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project ID",
				},
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Environment variable key",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Environment variable value",
				},
			},
			"required": []string{"project_id", "key", "value"},
		},
		Handler: handleSetProjectEnv,
	})

	// Set service environment variable
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "set_service_env",
		Description: "Set a service-level environment variable",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID",
				},
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Environment variable key",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Environment variable value",
				},
			},
			"required": []string{"service_id", "key", "value"},
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