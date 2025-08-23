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
		Description: `ESSENTIAL FIRST STEP: Discovers all services in a project with their IDs, hostnames, service types, deployment status, and environment variable availability.

CRITICAL: Requires a project ID. To get the project ID, the agent can run 'echo $projectId' in the container environment.

Returns condensed data about:
- All services with their unique IDs (required for other tools)
- Service hostnames, types, and current status
- Active app version details (for runtime services with deployments)
- Available environment variables at project and service level
- Current project configuration

Optional filters:
- service_id: Get details for a specific service by ID
- service_name: Get details for a specific service by hostname

Always use this tool first to understand the project structure before performing other operations.`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Zerops project ID. Get it by running 'echo $projectId' in the container.",
				},
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Service ID to get details for a single service only",
				},
				"service_name": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Service hostname/name to get details for a single service only",
				},
			},
			"required":             []string{"project_id"},
			"additionalProperties": false,
		},
		Handler: handleDiscovery,
	})
}

func handleDiscovery(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	// Debug: Log all received parameters
	fmt.Printf("DEBUG: Discovery received args: %+v\n", args)

	// Get project ID parameter
	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		// Check if it was passed as "projectId" instead
		if altProjectID, altOk := args["projectId"].(string); altOk && altProjectID != "" {
			projectID = altProjectID
			fmt.Printf("DEBUG: Found projectId parameter (camelCase): %s\n", projectID)
		} else {
			return shared.ErrorResponse("Project ID is required. Run 'echo $projectId' in the container to get it."), nil
		}
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

	// Get optional service filtering parameters
	serviceIDFilter, _ := args["service_id"].(string)
	serviceNameFilter, _ := args["service_name"].(string)
	
	// Search for services in this specific project
	searchItems := []body.EsSearchItem{
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
	}
	
	// Add service-specific filters if provided
	if serviceIDFilter != "" {
		searchItems = append(searchItems, body.EsSearchItem{
			Name:     "id",
			Operator: "eq",
			Value:    types.String(serviceIDFilter),
		})
	}
	
	if serviceNameFilter != "" {
		searchItems = append(searchItems, body.EsSearchItem{
			Name:     "name",
			Operator: "eq",
			Value:    types.String(serviceNameFilter),
		})
	}
	
	serviceFilter := body.EsFilter{
		Search: searchItems,
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
		message := "No services found. Use 'import_services' to add services."
		if serviceIDFilter != "" {
			message = fmt.Sprintf("No service found with ID '%s'", serviceIDFilter)
		} else if serviceNameFilter != "" {
			message = fmt.Sprintf("No service found with name '%s'", serviceNameFilter)
		}
		
		return map[string]interface{}{
			"services": []interface{}{},
			"project": map[string]interface{}{
				"id":       projectID,
				"name":     project.Name.Native(),
				"env_keys": projectEnvKeys,
			},
			"message": message,
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

		// Count running processes
		processCount := 0
		processFilter := body.EsFilter{
			Search: []body.EsSearchItem{
				{
					Name:     "serviceStackId",
					Operator: "eq",
					Value:    types.String(string(service.Id)),
				},
				{
					Name:     "status",
					Operator: "eq",
					Value:    types.String("running"),
				},
			},
		}
		processResp, err := client.PostProcessSearch(ctx, processFilter)
		if err == nil {
			if processOutput, err := processResp.Output(); err == nil {
				processCount = len(processOutput.Items)
			}
		}

		serviceInfo := map[string]interface{}{
			"id":            string(service.Id),
			"hostname":      service.Name.Native(),
			"type":          string(service.ServiceStackTypeVersionId),
			"status":        string(service.Status),
			"env_keys":      serviceEnvKeys,
			"process_count": processCount,
		}
		
		// Add active app version info if available (for runtime services)
		if service.ActiveAppVersion != nil {
			serviceInfo["active_version"] = map[string]interface{}{
				"id":         string(service.ActiveAppVersion.Id),
				"status":     string(service.ActiveAppVersion.Status),
				"created":    service.ActiveAppVersion.Created.Native(),
				"updated":    service.ActiveAppVersion.LastUpdate.Native(),
			}
		}
		services = append(services, serviceInfo)
	}

	return map[string]interface{}{
		"project": map[string]interface{}{
			"id":       projectID,
			"name":     project.Name.Native(),
			"env_keys": projectEnvKeys,
		},
		"services": services,
		"count":    len(services),
	}, nil
}