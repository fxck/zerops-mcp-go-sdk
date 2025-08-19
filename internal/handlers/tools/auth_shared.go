package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/sdk"
)

// RegisterAuthShared registers auth tools in the shared registry
func RegisterAuthShared() {
	// Register auth_validate
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "auth_validate",
		Description: "Validate your API key and show account information",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: handleAuthValidateShared,
	})

	// Register auth_show
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "auth_show",
		Description: "Show current authentication status and available regions",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: handleAuthShowShared,
	})
}

func handleAuthValidateShared(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided. Please set ZEROPS_API_KEY environment variable or provide it in the Authorization header."), nil
	}

	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return shared.ErrorResponse("Failed to parse user info"), nil
	}

	var message strings.Builder
	message.WriteString("✅ Authentication successful!\n\n")
	message.WriteString(fmt.Sprintf("User: %s %s\n", userOutput.FirstName, userOutput.LastName))
	message.WriteString(fmt.Sprintf("Email: %s\n", userOutput.Email))
	message.WriteString(fmt.Sprintf("\nAccess to %d organization(s):\n", len(userOutput.ClientUserList)))

	for _, clientUser := range userOutput.ClientUserList {
		message.WriteString(fmt.Sprintf("• %s\n", clientUser.Client.AccountName.Native()))
	}

	return shared.TextResponse(message.String()), nil
}

func handleAuthShowShared(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.TextResponse("Not authenticated\n\nSet ZEROPS_API_KEY environment variable or provide it in the Authorization header to authenticate."), nil
	}

	// Get user info
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return shared.TextResponse(fmt.Sprintf("Authentication check failed: %v\n\nPlease verify your API key.", err)), nil
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return shared.ErrorResponse("Failed to parse user info"), nil
	}

	// Get regions
	regionResp, err := client.GetRegion(ctx)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get regions: %v", err)), nil
	}

	regionOutput, err := regionResp.Output()
	if err != nil {
		return shared.ErrorResponse("Failed to parse regions"), nil
	}

	var message strings.Builder
	message.WriteString("🔐 Authentication Status\n\n")
	message.WriteString(fmt.Sprintf("User: %s %s\n", userOutput.FirstName, userOutput.LastName))
	message.WriteString(fmt.Sprintf("Email: %s\n", userOutput.Email))
	message.WriteString(fmt.Sprintf("\nOrganizations (%d):\n", len(userOutput.ClientUserList)))

	for _, clientUser := range userOutput.ClientUserList {
		message.WriteString(fmt.Sprintf("• %s\n", clientUser.Client.AccountName.Native()))
	}

	message.WriteString("\n📍 Available Regions:\n")
	for _, region := range regionOutput.Items {
		message.WriteString(fmt.Sprintf("• %s", region.Name.Native()))
		if region.IsDefault.Native() {
			message.WriteString(" (default)")
		}
		message.WriteString(fmt.Sprintf("\n  Address: %s\n", region.Address.Native()))
	}

	return shared.TextResponse(message.String()), nil
}
