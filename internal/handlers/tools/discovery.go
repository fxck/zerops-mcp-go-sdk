package tools

import (
	"context"
	"fmt"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// RegisterDiscovery registers the discovery tool
func RegisterDiscovery() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "discovery",
		Description: "Returns all services with IDs, hostnames, types, and environment variables availability for a specific project",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project ID to get services from",
				},
			},
			"required": []string{"project_id"},
		},
		Handler: handleDiscovery,
	})
}

func handleDiscovery(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	// Get project ID parameter
	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return shared.ErrorResponse("Project ID is required"), nil
	}

	// Get project details first (we need clientId for searches)
	projectPath := path.ProjectId{Id: uuid.ProjectId(projectID)}
	projectResp, err := client.GetProject(ctx, projectPath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get project: %v", err)), nil
	}

	projectOutput, err := projectResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse project: %v", err)), nil
	}

	// Search for the project to get envList
	projectFilter := body.EsFilter{
		Search: []body.EsSearchItem{
			{
				Name:     "id",
				Operator: "eq",
				Value:    types.String(projectID),
			},
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    projectOutput.ClientId.TypedString(),
			},
		},
	}

	projectSearchResp, err := client.PostProjectSearch(ctx, projectFilter)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to search project: %v", err)), nil
	}

	projectSearchOutput, err := projectSearchResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse project search: %v", err)), nil
	}

	if len(projectSearchOutput.Items) == 0 {
		return shared.ErrorResponse("Project not found"), nil
	}

	project := projectSearchOutput.Items[0]
	
	// Get project environment variables from envList
	var projectEnvKeys []string
	for _, envItem := range project.EnvList {
		projectEnvKeys = append(projectEnvKeys, envItem.Key.Native())
	}

	// Search for services in this specific project
	serviceFilter := body.EsFilter{
		Search: []body.EsSearchItem{
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    types.String(projectID),
			},
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    projectOutput.ClientId.TypedString(),
			},
		},
	}

	serviceResp, err := client.PostServiceStackSearch(ctx, serviceFilter)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to search services: %v", err)), nil
	}

	serviceOutput, err := serviceResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse services: %v", err)), nil
	}

	if len(serviceOutput.Items) == 0 {
		return map[string]interface{}{
			"services": []interface{}{},
			"project": map[string]interface{}{
				"id":   projectID,
				"name": project.Name.Native(),
				"environment_variables": map[string]interface{}{
					"project_env_keys": projectEnvKeys,
				},
			},
			"message": "No services found in this project. Use 'import_services' to add services.",
		}, nil
	}

	// Build service information for this project
	var services []map[string]interface{}
	for _, service := range serviceOutput.Items {
		// Get service environment variables
		var serviceEnvKeys []string
		servicePath := path.ServiceStackId{Id: service.Id}
		serviceEnvResp, err := client.GetServiceStackEnv(ctx, servicePath)
		if err == nil {
			if envOutput, err := serviceEnvResp.Output(); err == nil {
				// Extract env variable keys
				for _, envItem := range envOutput.Items {
					serviceEnvKeys = append(serviceEnvKeys, envItem.Key.Native())
				}
			}
		}

		serviceInfo := map[string]interface{}{
			"id":       string(service.Id),
			"hostname": service.Name.Native(),
			"type":     string(service.ServiceStackTypeVersionId),
			"environment_variables": map[string]interface{}{
				"service_env_keys": serviceEnvKeys,
			},
		}
		services = append(services, serviceInfo)
	}

	return map[string]interface{}{
		"services": services,
		"count":    len(services),
		"project": map[string]interface{}{
			"id":   projectID,
			"name": project.Name.Native(),
			"environment_variables": map[string]interface{}{
				"project_env_keys": projectEnvKeys,
			},
		},
	}, nil
}