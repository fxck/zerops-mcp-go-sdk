package tools

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// RegisterServices registers service tools in the global registry
func RegisterServices() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "service_list",
		Description: "List all services in a project",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project ID (22-char string from project_list)",
				},
			},
			"required": []string{"project_id"},
		},
		Handler: handleServiceList,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "service_info",
		Description: "Get detailed information about a service using its ID (not name)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (22-char string from service_list, NOT the service name)",
				},
			},
			"required": []string{"service_id"},
		},
		Handler: handleServiceInfo,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "service_delete",
		Description: "Delete a service from a project",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (22-char string, NOT the service name)",
				},
				"confirm": map[string]interface{}{
					"type":        "boolean",
					"description": "Must be true to confirm deletion",
				},
			},
			"required": []string{"service_id", "confirm"},
		},
		Handler: handleServiceDelete,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "service_enable_subdomain",
		Description: "Enable public subdomain access for a service",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (22-char string)",
				},
			},
			"required": []string{"service_id"},
		},
		Handler: handleServiceEnableSubdomain,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "service_disable_subdomain",
		Description: "Disable public subdomain access for a service",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (22-char string)",
				},
			},
			"required": []string{"service_id"},
		},
		Handler: handleServiceDisableSubdomain,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "service_start",
		Description: "Start a stopped service",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (22-char string)",
				},
			},
			"required": []string{"service_id"},
		},
		Handler: handleServiceStart,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "service_stop",
		Description: "Stop a running service",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (22-char string)",
				},
			},
			"required": []string{"service_id"},
		},
		Handler: handleServiceStop,
	})
}

func handleServiceList(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return shared.ErrorResponse("Project ID is required"), nil
	}

	// Get project details
	projectPath := path.ProjectId{Id: uuid.ProjectId(projectID)}
	projectResp, err := client.GetProject(ctx, projectPath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get project: %v", err)), nil
	}

	projectOutput, err := projectResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse project: %v", err)), nil
	}

	// Search for services in this project
	filter := body.EsFilter{
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

	resp, err := client.PostServiceStackSearch(ctx, filter)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to search services: %v", err)), nil
	}

	searchOutput, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	if len(searchOutput.Items) == 0 {
		return shared.TextResponse(fmt.Sprintf("No services found in project '%s'\n\n"+
			"Use 'project_import' to add services.", projectOutput.Name.Native())), nil
	}

	// Format output
	var message strings.Builder
	message.WriteString(fmt.Sprintf("Services in project '%s' (%d):\n\n",
		projectOutput.Name.Native(), len(searchOutput.Items)))

	for i, service := range searchOutput.Items {
		message.WriteString(formatService(i+1, service))
	}

	return shared.TextResponse(message.String()), nil
}

func handleServiceInfo(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := client.GetServiceStack(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get service: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("Service: %s\n", output.Name.Native()))
	message.WriteString(fmt.Sprintf("ID: %s\n", string(output.Id)))
	message.WriteString(fmt.Sprintf("Type: %s\n", string(output.ServiceStackTypeVersionId)))
	message.WriteString(fmt.Sprintf("Status: %s\n", string(output.Status)))
	message.WriteString(fmt.Sprintf("Mode: %s\n", string(output.Mode)))

	// Add subdomain URL if enabled
	if output.SubdomainAccess.Native() {
		subdomainURL := fmt.Sprintf("https://%s-%s.prg1.zerops.app",
			output.Name.Native(),
			output.Project.Name.Native())
		message.WriteString(fmt.Sprintf("\nPublic URL: %s\n", subdomainURL))
	}

	message.WriteString(fmt.Sprintf("\nCreated: %s", output.Created.Format("2006-01-02 15:04:05")))

	return shared.TextResponse(message.String()), nil
}

func handleServiceDelete(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	confirm, _ := args["confirm"].(bool)
	if !confirm {
		return shared.TextResponse("Deletion cancelled. Set confirm=true to proceed."), nil
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := client.DeleteServiceStack(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to delete service: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	return shared.TextResponse(fmt.Sprintf("Service deletion initiated\nProcess ID: %s", string(output.Id))), nil
}

func handleServiceEnableSubdomain(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
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

	return shared.TextResponse(fmt.Sprintf("Subdomain access enabled\nProcess ID: %s", string(output.Id))), nil
}

func handleServiceDisableSubdomain(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := client.PutServiceStackDisableSubdomainAccess(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to disable subdomain: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	return shared.TextResponse(fmt.Sprintf("Subdomain access disabled\nProcess ID: %s", string(output.Id))), nil
}

func handleServiceStart(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := client.PutServiceStackStart(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to start service: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	return shared.TextResponse(fmt.Sprintf("Service started\nProcess ID: %s", string(output.Id))), nil
}

func handleServiceStop(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := client.PutServiceStackStop(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to stop service: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	return shared.TextResponse(fmt.Sprintf("Service stopped\nProcess ID: %s", string(output.Id))), nil
}

// Helper function to format service output
func formatService(index int, service interface{}) string {
	// Use reflection to access fields since the exact type may vary
	v := reflect.ValueOf(service)

	// Get required fields
	name := v.FieldByName("Name")
	id := v.FieldByName("Id")
	status := v.FieldByName("Status")
	created := v.FieldByName("Created")
	serviceStackTypeInfo := v.FieldByName("ServiceStackTypeInfo")

	if !name.IsValid() || !id.IsValid() || !status.IsValid() || !created.IsValid() {
		return fmt.Sprintf("%d. Service (unable to read details)\n\n", index)
	}

	// Extract name
	nameStr := ""
	if nameMethod := name.MethodByName("Native"); nameMethod.IsValid() {
		results := nameMethod.Call(nil)
		if len(results) > 0 {
			nameStr = results[0].String()
		}
	}

	// Extract ID
	idStr := ""
	if idStringer, ok := id.Interface().(fmt.Stringer); ok {
		idStr = idStringer.String()
	}

	// Extract status
	statusStr := fmt.Sprintf("%v", status.Interface())

	// Extract service type info
	typeStr := "unknown"
	if serviceStackTypeInfo.IsValid() {
		if nameField := serviceStackTypeInfo.FieldByName("Name"); nameField.IsValid() {
			typeStr = fmt.Sprintf("%v", nameField.Interface())
		}
	}

	// Format creation time
	createdStr := ""
	if timeMethod := created.MethodByName("Format"); timeMethod.IsValid() {
		results := timeMethod.Call([]reflect.Value{reflect.ValueOf("2006-01-02 15:04:05")})
		if len(results) > 0 {
			createdStr = results[0].String()
		}
	}

	result := fmt.Sprintf("%d. %s\n", index, nameStr)
	result += fmt.Sprintf("   ID: %s\n", idStr)
	result += fmt.Sprintf("   Type: %s\n", typeStr)
	result += fmt.Sprintf("   Status: %s\n", statusStr)
	if createdStr != "" {
		result += fmt.Sprintf("   Created: %s\n", createdStr)
	}
	result += "\n"

	return result
}
