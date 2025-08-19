package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
	"gopkg.in/yaml.v3"
)

// RegisterProjectsShared registers project tools in the shared registry
func RegisterProjectsShared() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "project_list",
		Description: "List all projects across all your organizations",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: handleProjectListShared,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "project_search",
		Description: "Search for projects by name",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Project name to search for (partial match supported)",
				},
			},
			"required": []string{"name"},
		},
		Handler: handleProjectSearchShared,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "project_create",
		Description: "Create a new Zerops project",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Project name (alphanumeric, hyphens allowed)",
				},
				"region": map[string]interface{}{
					"type":        "string",
					"description": "Region code: 'prg1' (Prague)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Optional project description",
				},
			},
			"required": []string{"name"},
		},
		Handler: handleProjectCreateShared,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "project_delete",
		Description: "Delete a project (WARNING: Deletes all services and data)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project ID (22-char string like 'ePbuhAuFRTWx2tE3VCGBgQ')",
				},
				"confirm": map[string]interface{}{
					"type":        "boolean",
					"description": "Must be true to confirm deletion",
				},
			},
			"required": []string{"project_id", "confirm"},
		},
		Handler: handleProjectDeleteShared,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "project_import",
		Description: "Import services to a project using YAML configuration",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project ID (22-char string like 'ePbuhAuFRTWx2tE3VCGBgQ')",
				},
				"yaml": map[string]interface{}{
					"type":        "string",
					"description": "Zerops YAML configuration for services",
				},
			},
			"required": []string{"project_id", "yaml"},
		},
		Handler: handleProjectImportShared,
	})
}

func handleProjectListShared(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	// Get all organizations from user info
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get user info: %v", err)), nil
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse user info: %v", err)), nil
	}

	// Collect all projects from all organizations
	var projects []projectInfo
	for _, clientUser := range userOutput.ClientUserList {
		clientProjects := getProjectsForClientShared(ctx, client, clientUser)
		projects = append(projects, clientProjects...)
	}

	if len(projects) == 0 {
		return shared.TextResponse("No projects found.\n\nCreate your first project with 'project_create'"), nil
	}

	// Format output
	var message strings.Builder
	message.WriteString(fmt.Sprintf("Found %d project(s):\n\n", len(projects)))
	for i, p := range projects {
		message.WriteString(formatProjectShared(i+1, p))
	}

	return shared.TextResponse(message.String()), nil
}

func handleProjectSearchShared(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return shared.ErrorResponse("Project name is required"), nil
	}

	resp, err := client.GetProjectsByName(ctx, path.GetProjectsByName{
		Name: types.NewString(name),
	})
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Search failed: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	if len(output.Projects) == 0 {
		return shared.TextResponse(fmt.Sprintf("No projects found matching '%s'", name)), nil
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("Found %d project(s) matching '%s':\n\n", len(output.Projects), name))
	for i, project := range output.Projects {
		message.WriteString(fmt.Sprintf("%d. %s\n", i+1, project.Name.Native()))
		message.WriteString(fmt.Sprintf("   ID: %s\n", string(project.Id)))
		message.WriteString(fmt.Sprintf("   Status: %s\n\n", string(project.Status)))
	}

	return shared.TextResponse(message.String()), nil
}

func handleProjectCreateShared(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return shared.ErrorResponse("Project name is required"), nil
	}

	description, _ := args["description"].(string)

	// Get user info to find the first available ClientId
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get user info: %v", err)), nil
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse user info: %v", err)), nil
	}

	if len(userOutput.ClientUserList) == 0 {
		return shared.ErrorResponse("No organizations found for this user"), nil
	}

	// Use the first available client/organization
	clientId := userOutput.ClientUserList[0].ClientId

	projectBody := body.PostProject{
		Name:             types.NewString(name),
		ClientId:         clientId,
		TagList:          types.NewStringArray([]string{}),
		EnvVariables:     body.PostProjectEnvVariables{},
		PublicIpV4Shared: types.NewBool(true),
		Location:         types.StringNull{}, // Leave empty for default region
	}

	if description != "" {
		projectBody.Description = types.NewTextNull(description)
	}

	resp, err := client.PostProject(ctx, projectBody)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to create project: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	message := fmt.Sprintf("Project created successfully\n\n"+
		"Name: %s\n"+
		"ID: %s\n\n"+
		"Next: Use 'project_import' to add services",
		name, string(output.Id))

	return shared.TextResponse(message), nil
}

func handleProjectDeleteShared(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return shared.ErrorResponse("Project ID is required"), nil
	}

	confirm, _ := args["confirm"].(bool)
	if !confirm {
		return shared.TextResponse("Deletion cancelled. Set confirm=true to proceed."), nil
	}

	projectPath := path.ProjectId{
		Id: uuid.ProjectId(projectID),
	}

	resp, err := client.DeleteProject(ctx, projectPath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to delete project: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	return shared.TextResponse(fmt.Sprintf("Project deletion initiated\nProcess ID: %s", string(output.Id))), nil
}

func handleProjectImportShared(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
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
		return shared.TextResponse(fmt.Sprintf("Invalid YAML: %v", err)), nil
	}

	importBody := body.ServiceStackImport{
		ProjectId: uuid.ProjectId(projectID),
		Yaml:      types.NewText(yamlContent),
	}

	resp, err := client.PostServiceStackImport(ctx, importBody)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Import failed: %v", err)), nil
	}

	_, err = resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	return shared.TextResponse("Service import initiated\n\n" +
		"Services are being created. This may take a few moments.\n" +
		"Use 'service_list' to check status."), nil
}

// Helper functions
func getProjectsForClientShared(ctx context.Context, client *sdk.Handler, clientUser output.ClientUserExtra) []projectInfo {
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

func formatProjectShared(index int, p projectInfo) string {
	result := fmt.Sprintf("%d. %s\n", index, p.Project.Name.Native())
	result += fmt.Sprintf("   ID: %s\n", string(p.Project.Id))
	result += fmt.Sprintf("   Organization: %s\n", p.OrgName)
	result += fmt.Sprintf("   Status: %s\n", string(p.Project.Status))

	if desc, ok := p.Project.Description.Get(); ok {
		result += fmt.Sprintf("   Description: %s\n", desc.Native())
	}

	result += fmt.Sprintf("   Created: %s\n\n", p.Project.Created.Format("2006-01-02 15:04:05"))
	return result
}
