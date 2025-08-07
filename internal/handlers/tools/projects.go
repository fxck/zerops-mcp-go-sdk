package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
	"gopkg.in/yaml.v3"
)

// RegisterProjects registers project management tools
func RegisterProjects(server *mcp.Server, client *sdk.Handler) {
	registerProjectList(server, client)
	registerProjectSearch(server, client)
	registerProjectCreate(server, client)
	registerProjectDelete(server, client)
	registerProjectImport(server, client)
}

func registerProjectList(server *mcp.Server, client *sdk.Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_list",
		Description: "List all projects across all your organizations",
	}, func(ctx context.Context, _ *mcp.ServerSession, _ *mcp.CallToolParamsFor[EmptyArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		// Get all organizations from user info
		userResp, err := client.GetUserInfo(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("failed to get user info: %w", err)), nil
		}

		userOutput, err := userResp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse user info: %w", err)), nil
		}

		// Collect all projects from all organizations
		var projects []projectInfo
		for _, clientUser := range userOutput.ClientUserList {
			clientProjects := getProjectsForClient(ctx, client, clientUser)
			projects = append(projects, clientProjects...)
		}

		if len(projects) == 0 {
			return textResult("No projects found.\n\nCreate your first project with 'project_create'"), nil
		}

		// Format output
		message := fmt.Sprintf("Found %d project(s):\n\n", len(projects))
		for i, p := range projects {
			message += formatProject(i+1, p)
		}

		return textResult(message), nil
	})
}

func registerProjectSearch(server *mcp.Server, client *sdk.Handler) {
	type SearchArgs struct {
		Name string `json:"name" mcp:"Project name to search for (partial match supported)"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_search",
		Description: "Search for projects by name",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[SearchArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		resp, err := client.GetProjectsByName(ctx, path.GetProjectsByName{
			Name: types.NewString(args.Name),
		})
		if err != nil {
			return errorResult(fmt.Errorf("search failed: %w", err)), nil
		}

		output, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		if len(output.Projects) == 0 {
			return textResult(fmt.Sprintf("No projects found matching '%s'", args.Name)), nil
		}

		message := fmt.Sprintf("Found %d project(s) matching '%s':\n\n", len(output.Projects), args.Name)
		for i, project := range output.Projects {
			message += fmt.Sprintf("%d. %s\n", i+1, project.Name.Native())
			message += fmt.Sprintf("   ID: %s\n", string(project.Id))
			message += fmt.Sprintf("   Status: %s\n\n", string(project.Status))
		}

		return textResult(message), nil
	})
}

func registerProjectCreate(server *mcp.Server, client *sdk.Handler) {
	type CreateArgs struct {
		Name        string  `json:"name" mcp:"Project name (alphanumeric, hyphens allowed)"`
		Region      string  `json:"region" mcp:"Region code: 'prg1' (Prague)"`
		Description *string `json:"description,omitempty" mcp:"Optional project description"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_create",
		Description: "Create a new Zerops project",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		// Get user info to find the first available ClientId
		userResp, err := client.GetUserInfo(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("failed to get user info: %w", err)), nil
		}

		userOutput, err := userResp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse user info: %w", err)), nil
		}

		if len(userOutput.ClientUserList) == 0 {
			return errorResult(fmt.Errorf("no organizations found for this user")), nil
		}

		// Use the first available client/organization
		clientId := userOutput.ClientUserList[0].ClientId

		projectBody := body.PostProject{
			Name:             types.NewString(args.Name),
			ClientId:         clientId,
			TagList:          types.NewStringArray([]string{}), // Empty tag list
			EnvVariables:     body.PostProjectEnvVariables{},    // Empty env variables
			PublicIpV4Shared: types.NewBool(true),              // Enable shared public IP
			Location:         types.StringNull{},               // Leave location empty - API will use default region
		}
		
		// Note: The Location field should remain empty (StringNull{})
		// Setting it to "prg1" or any value causes "[400][locationNotFound] Location not found"
		// The API automatically uses the default region when Location is not set

		if args.Description != nil && *args.Description != "" {
			projectBody.Description = types.NewTextNull(*args.Description)
		}

		resp, err := client.PostProject(ctx, projectBody)
		if err != nil {
			return errorResult(fmt.Errorf("failed to create project: %w", err)), nil
		}

		output, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		message := fmt.Sprintf("Project created successfully\n\n"+
			"Name: %s\n"+
			"ID: %s\n"+
			"Region: %s\n\n"+
			"Next: Use 'project_import' to add services",
			args.Name, string(output.Id), args.Region)

		return textResult(message), nil
	})
}

func registerProjectDelete(server *mcp.Server, client *sdk.Handler) {
	type DeleteArgs struct {
		ProjectID string `json:"project_id" mcp:"Project ID (22-char string like 'ePbuhAuFRTWx2tE3VCGBgQ')"`
		Confirm   bool   `json:"confirm" mcp:"Must be true to confirm deletion"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_delete",
		Description: "Delete a project (WARNING: Deletes all services and data)",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		if !args.Confirm {
			return textResult("Deletion cancelled. Set confirm=true to proceed."), nil
		}

		projectPath := path.ProjectId{
			Id: uuid.ProjectId(args.ProjectID),
		}

		resp, err := client.DeleteProject(ctx, projectPath)
		if err != nil {
			return errorResult(fmt.Errorf("failed to delete project: %w", err)), nil
		}

		output, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		return textResult(fmt.Sprintf("Project deletion initiated\nProcess ID: %s", string(output.Id))), nil
	})
}

func registerProjectImport(server *mcp.Server, client *sdk.Handler) {
	type ImportArgs struct {
		ProjectID string `json:"project_id" mcp:"Project ID (22-char string like 'ePbuhAuFRTWx2tE3VCGBgQ')"`
		YAML      string `json:"yaml" mcp:"Zerops YAML configuration for services"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_import",
		Description: "Import services to a project using YAML configuration",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[ImportArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		// Validate YAML
		var yamlData interface{}
		if err := yaml.Unmarshal([]byte(args.YAML), &yamlData); err != nil {
			return textResult(fmt.Sprintf("Invalid YAML: %v", err)), nil
		}

		importBody := body.ServiceStackImport{
			ProjectId: uuid.ProjectId(args.ProjectID),
			Yaml:      types.NewText(args.YAML),
		}

		resp, err := client.PostServiceStackImport(ctx, importBody)
		if err != nil {
			return errorResult(fmt.Errorf("import failed: %w", err)), nil
		}

		_, err = resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		return textResult("Service import initiated\n\n" +
			"Services are being created. This may take a few moments.\n" +
			"Use 'service_list' to check status."), nil
	})
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

func formatProject(index int, p projectInfo) string {
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