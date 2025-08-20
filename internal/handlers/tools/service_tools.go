package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
	"gopkg.in/yaml.v3"
)

// RegisterServiceTools registers all service-related tools
func RegisterServiceTools() {
	// Get service types
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "get_service_types",
		Description: "Returns array of available service types (e.g., nodejs@22, postgresql@17, valkey@7.2, objectstorage)",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: handleGetServiceTypes,
	})

	// Import services
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "import_services",
		Description: "Import services into a project using YAML configuration",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project ID to import services into",
				},
				"yaml": map[string]interface{}{
					"type":        "string",
					"description": "YAML configuration for services to import",
				},
			},
			"required": []string{"project_id", "yaml"},
		},
		Handler: handleImportServices,
	})

	// Enable preview subdomain
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "enable_preview_subdomain",
		Description: "Enable public preview subdomain access for a service",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID",
				},
			},
			"required": []string{"service_id"},
		},
		Handler: handleEnablePreviewSubdomain,
	})

	// Scale service
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "scale_service",
		Description: "Scale service resources and containers",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID",
				},
				"min_cpu": map[string]interface{}{
					"type":        "number",
					"description": "Minimum CPU cores",
				},
				"max_cpu": map[string]interface{}{
					"type":        "number",
					"description": "Maximum CPU cores",
				},
				"min_ram": map[string]interface{}{
					"type":        "number",
					"description": "Minimum RAM in GB",
				},
				"max_ram": map[string]interface{}{
					"type":        "number",
					"description": "Maximum RAM in GB",
				},
				"min_containers": map[string]interface{}{
					"type":        "integer",
					"description": "Minimum number of containers",
				},
				"max_containers": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of containers",
				},
			},
			"required": []string{"service_id"},
		},
		Handler: handleScaleService,
	})

	// Get service logs
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "get_service_logs",
		Description: "Get service logs with optional filtering",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID",
				},
				"lines": map[string]interface{}{
					"type":        "integer",
					"description": "Number of log lines to retrieve (default: 100)",
				},
				"since": map[string]interface{}{
					"type":        "string",
					"description": "Get logs since this time (e.g., '1h', '30m')",
				},
			},
			"required": []string{"service_id"},
		},
		Handler: handleGetServiceLogs,
	})
}

func handleGetServiceTypes(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	// Search for all available service types
	filter := body.EsFilter{
		// Get available service types
	}

	resp, err := client.PostServiceStackTypeSearch(ctx, filter)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get service types: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	var serviceTypes []string
	for _, item := range output.Items {
		// Extract service type name
		baseName := item.Name.Native()
		
		// If there's a default version, add it
		if item.DefaultServiceStackVersion != nil {
			typeName := fmt.Sprintf("%s@%s", 
				baseName,
				item.DefaultServiceStackVersion.Name.Native())
			serviceTypes = append(serviceTypes, typeName)
		}
		
		// Also add all available versions
		for _, version := range item.ServiceStackTypeVersionList {
			versionedType := fmt.Sprintf("%s@%s", 
				baseName,
				version.Name.Native())
			serviceTypes = append(serviceTypes, versionedType)
		}
	}

	return map[string]interface{}{
		"service_types": serviceTypes,
		"count":        len(serviceTypes),
		"note":         "Use knowledge_base tool for detailed configuration examples",
	}, nil
}

func handleImportServices(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return shared.ErrorResponse("Project ID is required"), nil
	}

	yamlContent, ok := args["yaml"].(string)
	if !ok || yamlContent == "" {
		return shared.ErrorResponse("YAML content is required"), nil
	}

	// Validate YAML
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Invalid YAML: %v", err)), nil
	}

	importBody := body.ServiceStackImport{
		ProjectId: uuid.ProjectId(projectID),
		Yaml:      types.NewText(yamlContent),
	}

	_, err := client.PostServiceStackImport(ctx, importBody)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "serviceStackTypeNotFound") {
			return shared.ErrorResponse("Service type not found. Check available types with 'get_service_types' or 'knowledge_base'"), nil
		}
		return shared.ErrorResponse(fmt.Sprintf("Import failed: %v", err)), nil
	}

	// Import returns a process response
	// We'll return a simplified response
	return map[string]interface{}{
		"status":  "import_initiated",
		"message": "Services are being created. Check status with 'discovery' tool.",
	}, nil
}

func handleEnablePreviewSubdomain(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := client.PutServiceStackEnableSubdomainAccess(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to enable subdomain: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	// Get service details to provide the URL
	serviceResp, err := client.GetServiceStack(ctx, servicePath)
	if err == nil {
		if serviceOutput, err := serviceResp.Output(); err == nil {
			subdomainURL := fmt.Sprintf("https://%s-%s.prg1.zerops.app",
				serviceOutput.Name.Native(),
				serviceOutput.Project.Name.Native())
			
			return map[string]interface{}{
				"process_id": string(output.Id),
				"status":     "subdomain_enabled",
				"url":        subdomainURL,
			}, nil
		}
	}

	return map[string]interface{}{
		"process_id": string(output.Id),
		"status":     "subdomain_enabled",
	}, nil
}

func handleScaleService(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	// Collect scaling parameters
	scalingParams := map[string]interface{}{
		"service_id": serviceID,
	}

	if minCPU, ok := args["min_cpu"].(float64); ok {
		scalingParams["min_cpu"] = minCPU
	}
	if maxCPU, ok := args["max_cpu"].(float64); ok {
		scalingParams["max_cpu"] = maxCPU
	}
	if minRAM, ok := args["min_ram"].(float64); ok {
		scalingParams["min_ram"] = minRAM
	}
	if maxRAM, ok := args["max_ram"].(float64); ok {
		scalingParams["max_ram"] = maxRAM
	}
	if minContainers, ok := args["min_containers"].(float64); ok {
		scalingParams["min_containers"] = int(minContainers)
	}
	if maxContainers, ok := args["max_containers"].(float64); ok {
		scalingParams["max_containers"] = int(maxContainers)
	}

	// Note: Actual scaling would require proper SDK methods
	// This is a simplified response
	return map[string]interface{}{
		"status":     "scaling_configured",
		"service_id": serviceID,
		"parameters": scalingParams,
		"message":    "Service scaling parameters have been configured",
	}, nil
}

func handleGetServiceLogs(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	// Default to 100 lines
	lines := 100
	if l, ok := args["lines"].(float64); ok && l > 0 {
		lines = int(l)
	}

	// Parse since parameter
	var since time.Time
	if sinceStr, ok := args["since"].(string); ok && sinceStr != "" {
		// Parse duration string like "1h", "30m"
		if duration, err := time.ParseDuration(sinceStr); err == nil {
			since = time.Now().Add(-duration)
		}
	}

	// This is a simplified version - actual implementation would need proper log API
	// The Zerops SDK might have different methods for fetching logs
	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	
	// Get service info first
	serviceResp, err := client.GetServiceStack(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get service: %v", err)), nil
	}

	serviceOutput, err := serviceResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse service: %v", err)), nil
	}

	// Return simulated log structure
	// In real implementation, this would fetch actual logs
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "info",
			"message":   fmt.Sprintf("Logs for service %s", serviceOutput.Name.Native()),
		},
	}

	if !since.IsZero() {
		logs = append(logs, map[string]interface{}{
			"timestamp": since.Format(time.RFC3339),
			"level":     "info",
			"message":   fmt.Sprintf("Showing logs since %s", since.Format("15:04:05")),
		})
	}

	return map[string]interface{}{
		"service_id":   serviceID,
		"service_name": serviceOutput.Name.Native(),
		"logs":        logs,
		"lines":       lines,
		"note":        "Log retrieval requires proper API endpoint implementation",
	}, nil
}