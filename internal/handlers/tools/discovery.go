package tools

import (
	"context"
	"fmt"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
)

// RegisterDiscovery registers the discovery tool
func RegisterDiscovery() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "discovery",
		Description: "Returns all services with IDs, hostnames, types, and environment variables availability",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: handleDiscovery,
	})
}

func handleDiscovery(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	// Get user info to get all projects
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get user info: %v", err)), nil
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse user info: %v", err)), nil
	}

	var allServices []map[string]interface{}

	// Iterate through all organizations
	for _, clientUser := range userOutput.ClientUserList {
		// Get projects for this organization
		filter := body.EsFilter{
			Search: []body.EsSearchItem{{
				Name:     "clientId",
				Operator: "eq",
				Value:    clientUser.ClientId.TypedString(),
			}},
		}

		projectResp, err := client.PostProjectSearch(ctx, filter)
		if err != nil {
			continue
		}

		projectOutput, err := projectResp.Output()
		if err != nil {
			continue
		}

		// For each project, get services
		for _, project := range projectOutput.Items {
			// Get project environment variables (simplified)
			var projectEnvKeys []string
			// Note: Project env vars would be fetched here if available in SDK
			projectEnvKeys = append(projectEnvKeys, "env_configured")

			// Search for services in this project
			serviceFilter := body.EsFilter{
				Search: []body.EsSearchItem{
					{
						Name:     "projectId",
						Operator: "eq",
						Value:    types.String(string(project.Id)),
					},
					{
						Name:     "clientId",
						Operator: "eq",
						Value:    clientUser.ClientId.TypedString(),
					},
				},
			}

			serviceResp, err := client.PostServiceStackSearch(ctx, serviceFilter)
			if err != nil {
				continue
			}

			serviceOutput, err := serviceResp.Output()
			if err != nil {
				continue
			}

			// Build service information
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
					"project": map[string]interface{}{
						"id":   string(project.Id),
						"name": project.Name.Native(),
					},
					"organization": map[string]interface{}{
						"id":   string(clientUser.ClientId),
						"name": clientUser.Client.AccountName.Native(),
					},
					"environment_variables": map[string]interface{}{
						"project_env_keys": projectEnvKeys,
						"service_env_keys": serviceEnvKeys,
					},
				}
				allServices = append(allServices, serviceInfo)
			}
		}
	}

	if len(allServices) == 0 {
		return map[string]interface{}{
			"services": []interface{}{},
			"message":  "No services found. Use 'import_services' to add services to a project.",
		}, nil
	}

	return map[string]interface{}{
		"services": allServices,
		"count":    len(allServices),
	}, nil
}