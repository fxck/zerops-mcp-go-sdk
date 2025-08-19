package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zeropsio/zerops-go/sdk"
)

// RegisterAuth registers authentication-related tools
func RegisterAuth(server *mcp.Server, client *sdk.Handler) {
	// auth_validate - Validate API key and check authentication
	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_validate",
		Description: "Validate Zerops API key and check authentication",
	}, func(ctx context.Context, _ *mcp.ServerSession, _ *mcp.CallToolParamsFor[EmptyArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		resp, err := client.GetUserInfo(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("authentication failed: %w", err)), nil
		}

		output, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		message := fmt.Sprintf("Authentication successful\n\n"+
			"User: %s\n"+
			"Email: %s\n"+
			"Status: %s\n\n"+
			"Organizations:\n",
			output.FullName.Native(),
			output.Email.Native(),
			string(output.Status))

		for _, client := range output.ClientUserList {
			message += fmt.Sprintf("  • %s (Role: %s)\n",
				client.Client.AccountName.Native(),
				string(client.RoleCode))
		}

		return textResult(message), nil
	})

	// region_list - List available regions
	mcp.AddTool(server, &mcp.Tool{
		Name:        "region_list",
		Description: "List all available Zerops regions",
	}, func(ctx context.Context, _ *mcp.ServerSession, _ *mcp.CallToolParamsFor[EmptyArgs]) (*mcp.CallToolResultFor[struct{}], error) {
		resp, err := client.GetRegion(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("failed to list regions: %w", err)), nil
		}

		output, err := resp.Output()
		if err != nil {
			return errorResult(fmt.Errorf("failed to parse response: %w", err)), nil
		}

		message := fmt.Sprintf("Available regions (%d):\n\n", len(output.Items))
		for _, region := range output.Items {
			defaultStr := ""
			if region.IsDefault.Native() {
				defaultStr = " [DEFAULT]"
			}
			message += fmt.Sprintf("• %s - %s%s\n",
				region.Name.Native(),
				region.Address.Native(),
				defaultStr)
		}

		return textResult(message), nil
	})
}
