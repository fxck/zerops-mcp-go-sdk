package http

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
	"gopkg.in/yaml.v3"
)

// GetToolDefinitions returns all available tool definitions for discovery
func GetToolDefinitions() []interface{} {
	return []interface{}{
		// Project tools
		map[string]interface{}{
			"name":        "project_list",
			"description": "List all projects across all your organizations",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		map[string]interface{}{
			"name":        "project_create",
			"description": "Create a new Zerops project",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Project name",
					},
					"region": map[string]interface{}{
						"type":        "string",
						"description": "Region ID (optional)",
					},
				},
				"required": []string{"name"},
			},
		},
		map[string]interface{}{
			"name":        "project_delete",
			"description": "Delete a Zerops project",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"projectId": map[string]interface{}{
						"type":        "string",
						"description": "Project ID to delete",
					},
				},
				"required": []string{"projectId"},
			},
		},
		map[string]interface{}{
			"name":        "project_search",
			"description": "Search projects by name",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []string{"query"},
			},
		},
		map[string]interface{}{
			"name":        "project_import",
			"description": "Import services from YAML configuration",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"projectId": map[string]interface{}{
						"type":        "string",
						"description": "Project ID",
					},
					"yaml": map[string]interface{}{
						"type":        "string",
						"description": "YAML configuration content",
					},
				},
				"required": []string{"projectId", "yaml"},
			},
		},

		// Service tools
		map[string]interface{}{
			"name":        "service_list",
			"description": "List services in a project",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"projectId": map[string]interface{}{
						"type":        "string",
						"description": "Project ID",
					},
				},
				"required": []string{"projectId"},
			},
		},
		map[string]interface{}{
			"name":        "service_info",
			"description": "Get detailed information about a service",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"serviceId": map[string]interface{}{
						"type":        "string",
						"description": "Service ID",
					},
				},
				"required": []string{"serviceId"},
			},
		},
		map[string]interface{}{
			"name":        "service_delete",
			"description": "Delete a service",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"serviceId": map[string]interface{}{
						"type":        "string",
						"description": "Service ID to delete",
					},
				},
				"required": []string{"serviceId"},
			},
		},
		map[string]interface{}{
			"name":        "service_enable_subdomain",
			"description": "Enable public subdomain access for a service",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"serviceId": map[string]interface{}{
						"type":        "string",
						"description": "Service ID",
					},
				},
				"required": []string{"serviceId"},
			},
		},
		map[string]interface{}{
			"name":        "service_disable_subdomain",
			"description": "Disable public subdomain access for a service",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"serviceId": map[string]interface{}{
						"type":        "string",
						"description": "Service ID",
					},
				},
				"required": []string{"serviceId"},
			},
		},
		map[string]interface{}{
			"name":        "service_start",
			"description": "Start a stopped service",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"serviceId": map[string]interface{}{
						"type":        "string",
						"description": "Service ID",
					},
				},
				"required": []string{"serviceId"},
			},
		},
		map[string]interface{}{
			"name":        "service_stop",
			"description": "Stop a running service",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"serviceId": map[string]interface{}{
						"type":        "string",
						"description": "Service ID",
					},
				},
				"required": []string{"serviceId"},
			},
		},

		// Region tools
		map[string]interface{}{
			"name":        "region_list",
			"description": "List available Zerops regions",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},

		// Auth tools
		map[string]interface{}{
			"name":        "auth_validate",
			"description": "Validate your API key and show account information",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},

		// Deployment tools (instructions only)
		map[string]interface{}{
			"name":        "deploy_validate",
			"description": "Get instructions for validating deployment configuration",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"projectId": map[string]interface{}{
						"type":        "string",
						"description": "Project ID",
					},
					"serviceId": map[string]interface{}{
						"type":        "string",
						"description": "Service ID",
					},
				},
			},
		},
		map[string]interface{}{
			"name":        "deploy_push",
			"description": "Get instructions for deploying code to a service",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"projectId": map[string]interface{}{
						"type":        "string",
						"description": "Project ID",
					},
					"serviceId": map[string]interface{}{
						"type":        "string",
						"description": "Service ID",
					},
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Source path (optional, defaults to current directory)",
					},
				},
				"required": []string{"projectId", "serviceId"},
			},
		},
	}
}

// CallTool handles tool execution for HTTP transport
func CallTool(ctx context.Context, toolName string, args map[string]interface{}) interface{} {
	// Get the Zerops client from context
	client, _ := ctx.Value("zeropsClient").(*sdk.Handler)

	// For deployment tools, we provide instructions instead of executing
	if strings.HasPrefix(toolName, "deploy_") {
		return handleDeploymentTool(toolName, args)
	}

	// For other tools, we need a valid client
	if client == nil {
		return errorResponse("No API key provided. Please provide your Zerops API key in the Authorization header.")
	}

	switch toolName {
	// Project tools
	case "project_list":
		return handleProjectList(ctx, client)
	case "project_create":
		return handleProjectCreate(ctx, client, args)
	case "project_delete":
		return handleProjectDelete(ctx, client, args)
	case "project_search":
		return handleProjectSearch(ctx, client, args)
	case "project_import":
		return handleProjectImport(ctx, client, args)

	// Service tools
	case "service_list":
		return handleServiceList(ctx, client, args)
	case "service_info":
		return handleServiceInfo(ctx, client, args)
	case "service_delete":
		return handleServiceDelete(ctx, client, args)
	case "service_enable_subdomain":
		return handleServiceEnableSubdomain(ctx, client, args)
	case "service_disable_subdomain":
		return handleServiceDisableSubdomain(ctx, client, args)
	case "service_start":
		return handleServiceStart(ctx, client, args)
	case "service_stop":
		return handleServiceStop(ctx, client, args)

	// Region tools
	case "region_list":
		return handleRegionList(ctx, client)

	// Auth tools
	case "auth_validate":
		return handleAuthValidate(ctx, client)

	// Knowledge tools
	case "knowledge_search":
		return handleKnowledgeSearch(ctx, args)
	case "knowledge_get":
		return handleKnowledgeGet(ctx, args)

	default:
		return errorResponse(fmt.Sprintf("Unknown tool: %s", toolName))
	}
}

// Project handlers
func handleProjectList(ctx context.Context, client *sdk.Handler) interface{} {
	// Get user info to list all organizations
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to get user info: %v", err))
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse user info: %v", err))
	}

	var allProjects []projectInfo

	// Get projects for each client/organization
	for _, clientUser := range userOutput.ClientUserList {
		projects := getProjectsForClient(ctx, client, clientUser)
		allProjects = append(allProjects, projects...)
	}

	if len(allProjects) == 0 {
		return textResponse("No projects found.\n\nCreate your first project with 'project_create'")
	}

	// Format as text response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d project(s):\n\n", len(allProjects)))

	for i, p := range allProjects {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, p.Project.Name.Native()))
		result.WriteString(fmt.Sprintf("   ID: %s\n", string(p.Project.Id)))
		result.WriteString(fmt.Sprintf("   Organization: %s\n", p.OrgName))
		result.WriteString(fmt.Sprintf("   Status: %s\n", string(p.Project.Status)))

		if desc, ok := p.Project.Description.Get(); ok {
			result.WriteString(fmt.Sprintf("   Description: %s\n", desc.Native()))
		}

		result.WriteString(fmt.Sprintf("   Created: %s\n\n", p.Project.Created.Format("2006-01-02 15:04:05")))
	}

	return textResponse(result.String())
}

func handleProjectCreate(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return errorResponse("Project name is required")
	}

	// Get user info to find ClientId
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to get user info: %v", err))
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse user info: %v", err))
	}

	if len(userOutput.ClientUserList) == 0 {
		return errorResponse("No organizations found for this user")
	}

	// Use the first available client/organization
	clientId := userOutput.ClientUserList[0].ClientId

	// Create project request
	req := body.PostProject{
		Name:             types.NewString(name),
		ClientId:         clientId,
		TagList:          types.NewStringArray([]string{}),
		EnvVariables:     body.PostProjectEnvVariables{},
		PublicIpV4Shared: types.NewBool(true),
		Location:         types.StringNull{}, // Leave empty for default region
	}

	// Note: Location field should remain empty (StringNull{})
	// The API automatically uses the default region when Location is not set

	resp, err := client.PostProject(ctx, req)
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to create project: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	return textResponse(fmt.Sprintf("✅ Project '%s' created successfully!\n\nProject ID: %s\n\nNext steps:\n• Use 'service_create' to add services\n• Use 'project_import' to import from YAML",
		name, string(output.Id)))
}

func handleProjectDelete(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	projectId, ok := args["projectId"].(string)
	if !ok || projectId == "" {
		return errorResponse("Project ID is required")
	}

	_, err := client.DeleteProject(ctx, path.ProjectId{
		Id: uuid.ProjectId(projectId),
	})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to delete project: %v", err))
	}

	return textResponse(fmt.Sprintf("✅ Project %s deleted successfully", projectId))
}

func handleProjectSearch(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return errorResponse("Search query is required")
	}

	// Use GetProjectsByName for search
	resp, err := client.GetProjectsByName(ctx, path.GetProjectsByName{
		Name: types.NewString(query),
	})
	if err != nil {
		return errorResponse(fmt.Sprintf("Search failed: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	if len(output.Projects) == 0 {
		return textResponse(fmt.Sprintf("No projects found matching '%s'", query))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d project(s) matching '%s':\n\n", len(output.Projects), query))

	for i, project := range output.Projects {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, project.Name.Native()))
		result.WriteString(fmt.Sprintf("   ID: %s\n", string(project.Id)))
		result.WriteString(fmt.Sprintf("   Status: %s\n\n", string(project.Status)))
	}

	return textResponse(result.String())
}

func handleProjectImport(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	projectId, ok := args["projectId"].(string)
	if !ok || projectId == "" {
		return errorResponse("Project ID is required")
	}

	yamlContent, ok := args["yaml"].(string)
	if !ok || yamlContent == "" {
		return errorResponse("YAML content is required")
	}

	// Validate YAML
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err != nil {
		return errorResponse(fmt.Sprintf("Invalid YAML: %v", err))
	}

	// Get user info to find ClientId
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to get user info: %v", err))
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse user info: %v", err))
	}

	if len(userOutput.ClientUserList) == 0 {
		return errorResponse("No organizations found for this user")
	}

	// Use the first available client/organization
	clientId := userOutput.ClientUserList[0].ClientId

	// Import the services
	resp, err := client.PostProjectImport(ctx, body.ProjectImport{
		ClientId: clientId,
		Yaml:     types.NewText(yamlContent),
	})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to import services: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	return textResponse(fmt.Sprintf("✅ Services imported successfully!\n\nProject ID: %s\n\nThe services are being created. Use 'service_list' to check their status.",
		string(output.ProjectId)))
}

// Service handlers
func handleServiceList(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	projectId, ok := args["projectId"].(string)
	if !ok || projectId == "" {
		return errorResponse("Project ID is required")
	}

	// Get project details first
	projectResp, err := client.GetProject(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to get project details: %v", err))
	}

	projectOutput, err := projectResp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse project response: %v", err))
	}

	// Search for services in this project
	filter := body.EsFilter{
		Search: []body.EsSearchItem{
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    types.String(projectId),
			},
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    projectOutput.ClientId.TypedString(),
			},
		},
	}

	searchResp, err := client.PostServiceStackSearch(ctx, filter)
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to search services: %v", err))
	}

	searchOutput, err := searchResp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse search response: %v", err))
	}

	if len(searchOutput.Items) == 0 {
		return textResponse(fmt.Sprintf("No services found in project %s\n\nUse 'project_import' to add services.", projectOutput.Name.Native()))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Services in project %s (%d):\n\n", projectOutput.Name.Native(), len(searchOutput.Items)))

	for i, service := range searchOutput.Items {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, service.Name.Native()))
		result.WriteString(fmt.Sprintf("   ID: %s\n", string(service.Id)))
		result.WriteString(fmt.Sprintf("   Type: %s\n", string(service.ServiceStackTypeVersionId)))
		result.WriteString(fmt.Sprintf("   Status: %s\n", string(service.Status)))
		result.WriteString("\n")
	}

	return textResponse(result.String())
}

func handleServiceInfo(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	serviceId, ok := args["serviceId"].(string)
	if !ok || serviceId == "" {
		return errorResponse("Service ID is required")
	}

	resp, err := client.GetServiceStack(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to get service details: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Service: %s\n", output.Name.Native()))
	result.WriteString(fmt.Sprintf("ID: %s\n", string(output.Id)))
	result.WriteString(fmt.Sprintf("Type: %s\n", string(output.ServiceStackTypeVersionId)))
	result.WriteString(fmt.Sprintf("Status: %s\n", string(output.Status)))
	result.WriteString(fmt.Sprintf("Mode: %s\n", string(output.Mode)))

	return textResponse(result.String())
}

func handleServiceDelete(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	serviceId, ok := args["serviceId"].(string)
	if !ok || serviceId == "" {
		return errorResponse("Service ID is required")
	}

	_, err := client.DeleteServiceStack(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to delete service: %v", err))
	}

	return textResponse(fmt.Sprintf("✅ Service %s deleted successfully", serviceId))
}

func handleServiceEnableSubdomain(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	serviceId, ok := args["serviceId"].(string)
	if !ok || serviceId == "" {
		return errorResponse("Service ID is required")
	}

	resp, err := client.PutServiceStackEnableSubdomainAccess(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to enable subdomain: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	return textResponse(fmt.Sprintf("✅ Public subdomain access enabled for service %s\n\nProcess ID: %s",
		serviceId, string(output.Id)))
}

func handleServiceDisableSubdomain(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	serviceId, ok := args["serviceId"].(string)
	if !ok || serviceId == "" {
		return errorResponse("Service ID is required")
	}

	resp, err := client.PutServiceStackDisableSubdomainAccess(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to disable subdomain: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	return textResponse(fmt.Sprintf("✅ Public subdomain access disabled for service %s\n\nProcess ID: %s",
		serviceId, string(output.Id)))
}

func handleServiceStart(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	serviceId, ok := args["serviceId"].(string)
	if !ok || serviceId == "" {
		return errorResponse("Service ID is required")
	}

	resp, err := client.PutServiceStackStart(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to start service: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	return textResponse(fmt.Sprintf("✅ Service %s started successfully\n\nProcess ID: %s",
		serviceId, string(output.Id)))
}

func handleServiceStop(ctx context.Context, client *sdk.Handler, args map[string]interface{}) interface{} {
	serviceId, ok := args["serviceId"].(string)
	if !ok || serviceId == "" {
		return errorResponse("Service ID is required")
	}

	resp, err := client.PutServiceStackStop(ctx, path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)})
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to stop service: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	return textResponse(fmt.Sprintf("✅ Service %s stopped successfully\n\nProcess ID: %s",
		serviceId, string(output.Id)))
}

// Region handlers
func handleRegionList(ctx context.Context, client *sdk.Handler) interface{} {
	resp, err := client.GetRegion(ctx)
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to get regions: %v", err))
	}

	output, err := resp.Output()
	if err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse response: %v", err))
	}

	var result strings.Builder
	result.WriteString("Available Zerops regions:\n\n")

	for _, region := range output.Items {
		result.WriteString(fmt.Sprintf("• %s\n", region.Name.Native()))
		if region.IsDefault.Native() {
			result.WriteString("  (Default)\n")
		}
		result.WriteString(fmt.Sprintf("  Address: %s\n\n", region.Address.Native()))
	}

	return textResponse(result.String())
}

// Auth handlers
func handleAuthValidate(ctx context.Context, client *sdk.Handler) interface{} {
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return errorResponse(fmt.Sprintf("Authentication failed: %v", err))
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return errorResponse("Failed to parse user info")
	}

	var result strings.Builder
	result.WriteString("✅ Authentication successful!\n\n")
	result.WriteString(fmt.Sprintf("User: %s %s\n", userOutput.FirstName, userOutput.LastName))
	result.WriteString(fmt.Sprintf("Email: %s\n", userOutput.Email))
	result.WriteString(fmt.Sprintf("\nAccess to %d organization(s):\n", len(userOutput.ClientUserList)))

	for _, clientUser := range userOutput.ClientUserList {
		result.WriteString(fmt.Sprintf("• %s\n", clientUser.Client.AccountName.Native()))
	}

	return textResponse(result.String())
}

// Deployment handlers (instructions only for HTTP mode)
func handleDeploymentTool(toolName string, args map[string]interface{}) interface{} {
	switch toolName {
	case "deploy_validate":
		return handleDeployValidate(args)
	case "deploy_push":
		return handleDeployPush(args)
	default:
		return errorResponse(fmt.Sprintf("Unknown deployment tool: %s", toolName))
	}
}

func handleDeployValidate(args map[string]interface{}) interface{} {
	projectId, _ := args["projectId"].(string)
	serviceId, _ := args["serviceId"].(string)

	instructions := fmt.Sprintf(`To validate your deployment configuration:

1. Install zcli:
   npm i -g @zerops/zcli

2. Create a zerops.yml file in your project root with your service configuration

3. Validate the configuration:
   zcli service validate zerops.yml

Project ID: %s
Service ID: %s

For more information: https://docs.zerops.io/references/cli`, projectId, serviceId)

	return textResponse(instructions)
}

func handleDeployPush(args map[string]interface{}) interface{} {
	projectId, _ := args["projectId"].(string)
	serviceId, _ := args["serviceId"].(string)
	sourcePath, _ := args["source"].(string)

	if sourcePath == "" {
		sourcePath = "."
	}

	instructions := fmt.Sprintf(`To deploy your application to Zerops:

1. Install zcli if not already installed:
   npm i -g @zerops/zcli

2. Login to Zerops:
   zcli login

3. Deploy your application:
   zcli push --projectId %s --serviceId %s %s

Alternative using zerops.yml:
   zcli push

The deployment will:
• Build your application
• Deploy to the specified service
• Show real-time logs
• Provide deployment status

For more information: https://docs.zerops.io/references/cli/deploy`,
		projectId, serviceId, sourcePath)

	return textResponse(instructions)
}

// Knowledge handlers
func handleKnowledgeSearch(ctx context.Context, args map[string]interface{}) interface{} {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return errorResponse("Search query is required")
	}

	// This would call the knowledge API
	// For now, return a placeholder
	return textResponse(fmt.Sprintf("Knowledge search for '%s' would return deployment recipes and configurations.\n\nThis feature requires integration with the Zerops knowledge base API.", query))
}

func handleKnowledgeGet(ctx context.Context, args map[string]interface{}) interface{} {
	recipeId, ok := args["recipeId"].(string)
	if !ok || recipeId == "" {
		return errorResponse("Recipe ID is required")
	}

	// This would fetch specific recipe from knowledge API
	return textResponse(fmt.Sprintf("Recipe %s details would be fetched from the knowledge base.", recipeId))
}

// Helper types and functions
type projectInfo struct {
	Project output.EsProject
	OrgName string
	OrgId   string
}

func getProjectsForClient(ctx context.Context, client *sdk.Handler, clientUser output.ClientUserExtra) []projectInfo {
	filter := body.EsFilter{
		Search: []body.EsSearchItem{{
			Name:     "clientId",
			Operator: "eq",
			Value:    clientUser.ClientId.TypedString(),
		}},
	}

	resp, err := client.PostProjectSearch(ctx, filter)
	if err != nil {
		return nil
	}

	searchOutput, err := resp.Output()
	if err != nil {
		return nil
	}

	var projects []projectInfo
	for _, project := range searchOutput.Items {
		projects = append(projects, projectInfo{
			Project: project,
			OrgName: clientUser.Client.AccountName.Native(),
			OrgId:   string(clientUser.ClientId),
		})
	}

	return projects
}

// Helper functions
func textResponse(text string) interface{} {
	return map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": text,
			},
		},
	}
}

func errorResponse(message string) interface{} {
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
