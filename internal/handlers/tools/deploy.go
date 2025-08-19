package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/sdk"
)

// RegisterDeploy registers deployment tools in the global registry
func RegisterDeploy() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "deploy_validate",
		Description: "Validate deployment configuration (zerops.yml)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project ID (22-char string)",
				},
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (22-char string)",
				},
				"yaml_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to zerops.yml file (optional, defaults to ./zerops.yml)",
				},
			},
		},
		Handler: handleDeployValidate,
	})

	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "deploy_push",
		Description: "Deploy application source code to a service",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project ID (22-char string)",
				},
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (22-char string)",
				},
				"source_path": map[string]interface{}{
					"type":        "string",
					"description": "Source directory path (optional, defaults to current directory)",
				},
			},
			"required": []string{"project_id", "service_id"},
		},
		Handler: handleDeployPush,
	})
}

func handleDeployValidate(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	// Check if we're in HTTP mode by looking for the transport context
	isHTTPMode := ctx.Value("httpMode") != nil

	projectID, _ := args["project_id"].(string)
	serviceID, _ := args["service_id"].(string)
	yamlPath, _ := args["yaml_path"].(string)

	if yamlPath == "" {
		yamlPath = "./zerops.yml"
	}

	// In HTTP mode, provide instructions
	if isHTTPMode {
		return shared.TextResponse(fmt.Sprintf(
			"To validate deployment configuration:\n\n"+
				"1. Install zcli: npm i -g @zerops/zcli\n"+
				"2. Create zerops.yml in your project\n"+
				"3. Run: zcli service validate %s\n\n"+
				"Project: %s\n"+
				"Service: %s\n\n"+
				"For more information: https://docs.zerops.io/references/cli",
			yamlPath, projectID, serviceID)), nil
	}

	// In stdio mode, execute zcli
	return executeZcliValidate(yamlPath, projectID, serviceID)
}

func handleDeployPush(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	// Check if we're in HTTP mode
	isHTTPMode := ctx.Value("httpMode") != nil

	projectID, _ := args["project_id"].(string)
	serviceID, _ := args["service_id"].(string)
	sourcePath, _ := args["source_path"].(string)

	if sourcePath == "" {
		sourcePath = "."
	}

	// In HTTP mode, provide instructions
	if isHTTPMode {
		return shared.TextResponse(fmt.Sprintf(
			"To deploy your application:\n\n"+
				"1. Install zcli: npm i -g @zerops/zcli\n"+
				"2. Login: zcli login\n"+
				"3. Deploy: zcli push --projectId %s --serviceId %s %s\n\n"+
				"Or with zerops.yml: zcli push\n\n"+
				"The deployment will:\n"+
				"• Build your application\n"+
				"• Deploy to the specified service\n"+
				"• Show real-time logs\n"+
				"• Provide deployment status\n\n"+
				"For more information: https://docs.zerops.io/references/cli/deploy",
			projectID, serviceID, sourcePath)), nil
	}

	// In stdio mode, execute zcli
	return executeZcliPush(projectID, serviceID, sourcePath)
}

// executeZcliValidate runs zcli validation (stdio mode only)
func executeZcliValidate(yamlPath, projectID, serviceID string) (interface{}, error) {
	// Check if file exists
	absPath, err := filepath.Abs(yamlPath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return shared.ErrorResponse(fmt.Sprintf("File not found: %s", absPath)), nil
	}

	// Run zcli validate
	cmd := exec.Command("zcli", "service", "validate", absPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return shared.TextResponse(fmt.Sprintf("Validation failed:\n\n%s\n\nError: %v",
			string(output), err)), nil
	}

	return shared.TextResponse(fmt.Sprintf("✅ Validation successful!\n\n%s\n\n"+
		"Your zerops.yml is valid and ready for deployment.\n"+
		"Use 'deploy_push' to deploy.", string(output))), nil
}

// executeZcliPush runs zcli push (stdio mode only)
func executeZcliPush(projectID, serviceID, sourcePath string) (interface{}, error) {
	// Validate source path
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return shared.ErrorResponse(fmt.Sprintf("Directory not found: %s", absPath)), nil
	}

	// Build command
	args := []string{"push"}

	// Add IDs if provided
	if projectID != "" && serviceID != "" {
		args = append(args, "--projectId", projectID, "--serviceId", serviceID)
	}

	// Add source path if not current directory
	if sourcePath != "." && sourcePath != "" {
		args = append(args, sourcePath)
	}

	// Run zcli push
	cmd := exec.Command("zcli", args...)
	cmd.Dir = absPath

	// Start the command
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if zcli is installed
		if strings.Contains(string(output), "command not found") ||
			strings.Contains(err.Error(), "executable file not found") {
			return shared.TextResponse("zcli is not installed.\n\n" +
				"Install it with: npm i -g @zerops/zcli\n" +
				"Then login with: zcli login"), nil
		}

		return shared.TextResponse(fmt.Sprintf("Deployment failed:\n\n%s\n\nError: %v",
			string(output), err)), nil
	}

	return shared.TextResponse(fmt.Sprintf("🚀 Deployment started!\n\n%s", string(output))), nil
}
