package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zerops-go/sdk"
)

// RegisterDeploy registers deployment-related tools
func RegisterDeploy(server *mcp.Server, client *sdk.Handler) {
	registerDeployValidate(server)
	registerDeployPush(server, client)
}

func registerDeployValidate(server *mcp.Server) {
	type ValidateArgs struct {
		WorkingDir *string `json:"working_dir,omitempty" mcp:"Working directory path (default: current dir)"`
		ConfigPath *string `json:"config_path,omitempty" mcp:"Path to zerops.yml (default: ./zerops.yml)"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deploy_validate",
		Description: "Validate deployment prerequisites",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[ValidateArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		// Determine working directory
		workDir := "."
		if args.WorkingDir != nil && *args.WorkingDir != "" {
			workDir = *args.WorkingDir
		}

		// Check if directory exists
		if _, err := os.Stat(workDir); os.IsNotExist(err) {
			return textResult(fmt.Sprintf("Error: Directory not found: %s", workDir)), nil
		}

		// Check if git is initialized
		gitDir := filepath.Join(workDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return textResult(fmt.Sprintf("Error: Git not initialized in %s\n\n"+
				"Run these commands to initialize git:\n"+
				"  cd %s\n"+
				"  git init\n"+
				"  git add .\n"+
				"  git commit -m \"Initial commit\"\n\n"+
				"Note: Zerops requires at least one commit to deploy.", workDir, workDir)), nil
		}

		// Determine config path
		configPath := filepath.Join(workDir, "zerops.yml")
		if args.ConfigPath != nil && *args.ConfigPath != "" {
			configPath = *args.ConfigPath
			if !filepath.IsAbs(configPath) {
				configPath = filepath.Join(workDir, configPath)
			}
		}

		// Check if config exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return textResult(fmt.Sprintf("Error: Config file not found: %s\n\n"+
				"Create a zerops.yml file with your deployment configuration.", configPath)), nil
		}

		// Check if zcli is installed
		zcliPath, err := exec.LookPath("zcli")
		if err != nil {
			return textResult("Error: zcli not found\n\n" +
				"Install zcli: https://docs.zerops.io/references/cli"), nil
		}

		// Check zcli version
		cmd := exec.CommandContext(ctx, "zcli", "version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return textResult(fmt.Sprintf("Error: Failed to check zcli version: %v", err)), nil
		}

		message := "Deployment validation successful\n\n"
		message += fmt.Sprintf("✓ Working directory: %s\n", workDir)
		message += fmt.Sprintf("✓ Git initialized: %s\n", gitDir)
		message += fmt.Sprintf("✓ Config file: %s\n", configPath)
		message += fmt.Sprintf("✓ zcli path: %s\n", zcliPath)
		message += fmt.Sprintf("✓ zcli version: %s", strings.TrimSpace(string(output)))

		return textResult(message), nil
	})
}

func registerDeployPush(server *mcp.Server, client *sdk.Handler) {
	type PushArgs struct {
		ServiceID  string  `json:"service_id" mcp:"Service ID (22-char string from service_list, NOT the service name like 'kbapi')"`
		ProjectID  string  `json:"project_id" mcp:"Project ID (22-char string from project_list)"`
		WorkingDir *string `json:"working_dir,omitempty" mcp:"Working directory with your code (default: current dir)"`
		ConfigPath *string `json:"config_path,omitempty" mcp:"Path to zerops.yml config file (default: ./zerops.yml)"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deploy_push",
		Description: "Deploy application to Zerops - requires zcli installed and both project ID and service ID",
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[PushArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		args := params.Arguments

		// Validate service ID format (should be 22 characters)
		if len(args.ServiceID) < 20 || len(args.ServiceID) > 24 {
			return textResult(fmt.Sprintf("Error: Invalid service ID format: '%s'\n\n"+
				"Expected: 22-character ID like 'ePbuhAuFRTWx2tE3VCGBgQ'\n"+
				"Got: %d characters\n\n"+
				"Use 'service_list' to get the correct service ID.\n"+
				"Note: Service ID is NOT the service name (like 'kbapi' or 'api').", 
				args.ServiceID, len(args.ServiceID))), nil
		}

		// Validate project ID format 
		if len(args.ProjectID) < 20 || len(args.ProjectID) > 24 {
			return textResult(fmt.Sprintf("Error: Invalid project ID format: '%s'\n\n"+
				"Expected: 22-character ID\n"+
				"Got: %d characters\n\n"+
				"Use 'project_list' to get the correct project ID.", 
				args.ProjectID, len(args.ProjectID))), nil
		}

		// Validate zcli is available
		if _, err := exec.LookPath("zcli"); err != nil {
			return textResult("Error: zcli not found. Install from: https://docs.zerops.io/references/cli"), nil
		}

		// Determine working directory
		workDir := "."
		if args.WorkingDir != nil && *args.WorkingDir != "" {
			workDir = *args.WorkingDir
		}

		// Check if git is initialized (required for zcli push)
		gitDir := filepath.Join(workDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return textResult(fmt.Sprintf("Error: Git not initialized in %s\n\n"+
				"Run these commands in your project directory:\n"+
				"  cd %s\n"+
				"  git init\n"+
				"  git add .\n"+
				"  git commit -m \"Initial commit\"\n\n"+
				"Note: Zerops requires at least one commit to track files for deployment.", workDir, workDir)), nil
		}

		// First, ensure zcli is logged in
		apiKey := os.Getenv("ZEROPS_API_KEY")
		if apiKey == "" {
			return textResult("Error: ZEROPS_API_KEY environment variable not set"), nil
		}

		// Login to zcli if needed
		loginCmd := exec.CommandContext(ctx, "zcli", "login", "--token", apiKey)
		_, _ = loginCmd.CombinedOutput()
		
		// Build zcli command with both projectId and serviceId to avoid interactive prompts
		cmdArgs := []string{"push"}
		
		// Add BOTH project ID and service ID to avoid prompts
		cmdArgs = append(cmdArgs, "--projectId", args.ProjectID)
		cmdArgs = append(cmdArgs, "--serviceId", args.ServiceID)

		// Add config path if specified
		if args.ConfigPath != nil && *args.ConfigPath != "" {
			cmdArgs = append(cmdArgs, "--zeropsYamlPath", *args.ConfigPath)
		}
		
		// Add working directory if not current
		if workDir != "." && workDir != "" {
			cmdArgs = append(cmdArgs, "--workingDir", workDir)
		}

		// Execute zcli push
		cmd := exec.CommandContext(ctx, "zcli", cmdArgs...)
		cmd.Dir = workDir

		// Set environment with token
		cmd.Env = append(os.Environ(), 
			fmt.Sprintf("ZEROPS_TOKEN=%s", apiKey))

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if err != nil {
			return textResult(fmt.Sprintf("Deployment failed\n\n"+
				"Command: zcli %s\n"+
				"Working dir: %s\n\n"+
				"Output:\n%s\n\n"+
				"Error: %v",
				strings.Join(cmdArgs, " "), workDir, outputStr, err)), nil
		}

		// Parse output for success indicators
		if strings.Contains(outputStr, "Deploy finished") || 
		   strings.Contains(outputStr, "successfully") ||
		   strings.Contains(outputStr, "Build completed") {
			return textResult(fmt.Sprintf("Deployment successful\n\n"+
				"Service ID: %s\n\n"+
				"Output:\n%s", args.ServiceID, outputStr)), nil
		}

		return textResult(fmt.Sprintf("⚠️ Deployment completed with output:\n\n%s", outputStr)), nil
	})
}
