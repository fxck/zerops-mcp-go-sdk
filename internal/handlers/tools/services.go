package tools

import (
	"context"
	"fmt"
	"reflect"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// RegisterServices registers service management tools
func RegisterServices(server *mcp.Server, client *sdk.Handler) {
	registerServiceList(server, client)
	registerServiceInfo(server, client)
	registerServiceDelete(server, client)
	registerServiceEnableSubdomain(server, client)
}

func registerServiceList(server *mcp.Server, client *sdk.Handler) {
	type ListArgs struct {
		ProjectID string `json:"project_id" mcp:"Project ID (22-char string from project_list or project_create)"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "service_list",
		Description: "List all services in a project (returns service IDs needed for other operations)",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[ListArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		// Get project details to find clientId (required for search)
		projectPath := path.ProjectId{
			Id: uuid.ProjectId(args.ProjectID),
		}

		projectResp, err := client.GetProject(ctx, projectPath)
		if err != nil {
			return errorResult(fmt.Errorf("failed to get project: %w", err)), nil
		}

		projectOutput, err := projectResp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse project: %w", err)), nil
		}

		// Search for services (requires both projectId AND clientId)
		filter := body.EsFilter{
			Search: []body.EsSearchItem{
				{
					Name:     "projectId",
					Operator: "eq",
					Value:    types.NewString(args.ProjectID),
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
			return errorResult(fmt.Errorf("failed to list services: %w", err)), nil
		}

		searchOutput, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		if len(searchOutput.Items) == 0 {
			return textResult(fmt.Sprintf("No services found in project '%s'\n\n"+
				"Use 'project_import' to add services.", projectOutput.Name.Native())), nil
		}

		// Format output
		message := fmt.Sprintf("Services in project '%s' (%d):\n\n", 
			projectOutput.Name.Native(), len(searchOutput.Items))
		
		for i, service := range searchOutput.Items {
			message += formatService(i+1, service)
		}

		return textResult(message), nil
	})
}

func registerServiceInfo(server *mcp.Server, client *sdk.Handler) {
	type InfoArgs struct {
		ServiceID string `json:"service_id" mcp:"Service ID (22-char string from service_list, NOT the service name)"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "service_info",
		Description: "Get detailed information about a service using its ID (not name)",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[InfoArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		servicePath := path.ServiceStackId{
			Id: uuid.ServiceStackId(args.ServiceID),
		}

		resp, err := client.GetServiceStack(ctx, servicePath)
		if err != nil {
			return errorResult(fmt.Errorf("failed to get service: %w", err)), nil
		}

		output, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		message := fmt.Sprintf("Service Details:\n\n"+
			"Name: %s\n"+
			"ID: %s\n"+
			"Type: %s\n"+
			"Status: %s\n"+
			"Project ID: %s\n"+
			"Created: %s\n",
			output.Name.Native(),
			string(output.Id),
			output.ServiceStackTypeInfo.ServiceStackTypeName.Native(),
			string(output.Status),
			string(output.ProjectId),
			output.Created.Format("2006-01-02 15:04:05"))

		// Add version info if available
		if output.VersionNumber.Native() != "" {
			message += fmt.Sprintf("Version: %s\n", output.VersionNumber.Native())
		}

		// Add mode info
		message += fmt.Sprintf("Mode: %s\n", string(output.Mode))

		// Add subdomain URL if enabled
		if output.SubdomainAccess.Native() {
			// Construct subdomain URL based on Zerops pattern: {service}-{project}.prg1.zerops.app
			subdomainURL := fmt.Sprintf("https://%s-%s.prg1.zerops.app", 
				output.Name.Native(), 
				output.Project.Name.Native())
			message += fmt.Sprintf("\nPublic URL: %s\n", subdomainURL)
		}

		// Add ports if any
		if len(output.Ports) > 0 {
			message += "\nPorts:\n"
			for _, port := range output.Ports {
				message += fmt.Sprintf("  • %d (%s)\n", 
					port.Port.Native(), port.Scheme.Native())
			}
		}

		return textResult(message), nil
	})
}

func registerServiceDelete(server *mcp.Server, client *sdk.Handler) {
	type DeleteArgs struct {
		ServiceID string `json:"service_id" mcp:"Service ID (22-char string from service_list, NOT the service name)"`
		Confirm   bool   `json:"confirm" mcp:"Must be true to confirm deletion"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "service_delete",
		Description: "Delete a service by its ID (requires confirmation)",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		if !args.Confirm {
			return textResult("Deletion cancelled. Set confirm=true to proceed."), nil
		}

		servicePath := path.ServiceStackId{
			Id: uuid.ServiceStackId(args.ServiceID),
		}

		resp, err := client.DeleteServiceStack(ctx, servicePath)
		if err != nil {
			return errorResult(fmt.Errorf("failed to delete service: %w", err)), nil
		}

		output, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		return textResult(fmt.Sprintf("Service deletion initiated\nProcess ID: %s", 
			string(output.Id))), nil
	})
}

func registerServiceEnableSubdomain(server *mcp.Server, client *sdk.Handler) {
	type EnableSubdomainArgs struct {
		ServiceID string `json:"service_id" mcp:"Service ID (22-char string from service_list, NOT the service name)"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "service_enable_subdomain",
		Description: "Enable public subdomain URL for a service (requires service ID, not name)",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[EnableSubdomainArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		servicePath := path.ServiceStackId{
			Id: uuid.ServiceStackId(args.ServiceID),
		}

		// Enable subdomain access
		resp, err := client.PutServiceStackEnableSubdomainAccess(ctx, servicePath)
		if err != nil {
			return errorResult(fmt.Errorf("failed to enable subdomain: %w", err)), nil
		}

		output, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		// Get service info to show the new subdomain
		serviceResp, err := client.GetServiceStack(ctx, servicePath)
		if err != nil {
			return textResult(fmt.Sprintf("Subdomain enabled\nProcess ID: %s\n\nNote: Use 'service_info' to see the subdomain URL once ready.", 
				string(output.Id))), nil
		}

		serviceOutput, err := serviceResp.Output()
		if err != nil {
			return textResult(fmt.Sprintf("Subdomain enabled\nProcess ID: %s", string(output.Id))), nil
		}

		message := fmt.Sprintf("Subdomain access enabled\n\n")
		message += fmt.Sprintf("Service: %s\n", serviceOutput.Name.Native())
		message += fmt.Sprintf("Process ID: %s\n", string(output.Id))
		
		// Show the subdomain URL
		if serviceOutput.SubdomainAccess.Native() || true { // Will be enabled after this operation
			// Construct subdomain URL based on Zerops pattern: {service}-{project}.prg1.zerops.app
			subdomainURL := fmt.Sprintf("https://%s-%s.prg1.zerops.app", 
				serviceOutput.Name.Native(), 
				serviceOutput.Project.Name.Native())
			message += fmt.Sprintf("\nPublic URL: %s\n", subdomainURL)
			message += "\nNote: It may take a moment for the subdomain to become active."
		}

		return textResult(message), nil
	})
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
	
	// Extract ID - handle UUID type properly
	idStr := ""
	if idStringer, ok := id.Interface().(fmt.Stringer); ok {
		idStr = idStringer.String()
	} else {
		// Try to call String() method directly
		if strMethod := id.MethodByName("String"); strMethod.IsValid() {
			results := strMethod.Call(nil)
			if len(results) > 0 {
				idStr = results[0].String()
			}
		} else {
			idStr = fmt.Sprintf("%v", id.Interface())
		}
	}
	
	// Extract status
	statusStr := fmt.Sprintf("%v", status.Interface())
	
	// Extract type name
	typeStr := "Unknown"
	if serviceStackTypeInfo.IsValid() {
		if typeNameField := serviceStackTypeInfo.FieldByName("ServiceStackTypeName"); typeNameField.IsValid() {
			if nativeMethod := typeNameField.MethodByName("Native"); nativeMethod.IsValid() {
				results := nativeMethod.Call(nil)
				if len(results) > 0 {
					typeStr = results[0].String()
				}
			}
		}
	}
	
	// Extract created date
	createdStr := ""
	if formatMethod := created.MethodByName("Format"); formatMethod.IsValid() {
		results := formatMethod.Call([]reflect.Value{reflect.ValueOf("2006-01-02 15:04:05")})
		if len(results) > 0 {
			createdStr = results[0].String()
		}
	}
	
	return fmt.Sprintf("%d. SERVICE: %s\n"+
		"   SERVICE_ID: %s  <-- Use this ID for deploy_push\n"+
		"   Type: %s\n"+
		"   Status: %s\n"+
		"   Created: %s\n\n",
		index,
		nameStr,
		idStr,
		typeStr,
		statusStr,
		createdStr)
}