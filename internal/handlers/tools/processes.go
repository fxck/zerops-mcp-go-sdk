package tools

import (
	"context"
	"fmt"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// RegisterProcesses registers process monitoring tools
func RegisterProcesses() {
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name:        "get_running_processes",
		Description: "Get running processes, optionally filtered by service",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional service ID to filter processes",
				},
			},
		},
		Handler: handleGetRunningProcesses,
	})
}

func handleGetRunningProcesses(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	// Check if service_id is provided
	serviceID, hasServiceID := args["service_id"].(string)

	if hasServiceID && serviceID != "" {
		// Get processes for specific service
		servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
		
		// Get service details
		serviceResp, err := client.GetServiceStack(ctx, servicePath)
		if err != nil {
			return shared.ErrorResponse(fmt.Sprintf("Failed to get service: %v", err)), nil
		}

		serviceOutput, err := serviceResp.Output()
		if err != nil {
			return shared.ErrorResponse(fmt.Sprintf("Failed to parse service: %v", err)), nil
		}

		// Get processes for this service
		processFilter := body.EsFilter{
			Search: []body.EsSearchItem{
				{
					Name:     "serviceStackId",
					Operator: "eq",
					Value:    types.String(serviceID),
				},
			},
		}

		processResp, err := client.PostProcessSearch(ctx, processFilter)
		if err != nil {
			return shared.ErrorResponse(fmt.Sprintf("Failed to get processes: %v", err)), nil
		}

		processOutput, err := processResp.Output()
		if err != nil {
			return shared.ErrorResponse(fmt.Sprintf("Failed to parse processes: %v", err)), nil
		}

		var processes []map[string]interface{}
		for _, process := range processOutput.Items {
			processInfo := map[string]interface{}{
				"id":           string(process.Id),
				"status":       string(process.Status),
				"created":      process.Created.Format("2006-01-02 15:04:05"),
				"service_name": serviceOutput.Name.Native(),
				"service_id":   serviceID,
			}
			processes = append(processes, processInfo)
		}

		return map[string]interface{}{
			"processes": processes,
			"count":     len(processes),
			"service":   serviceOutput.Name.Native(),
		}, nil
	}

	// Get all processes across all services
	userResp, err := client.GetUserInfo(ctx)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get user info: %v", err)), nil
	}

	userOutput, err := userResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse user info: %v", err)), nil
	}

	var allProcesses []map[string]interface{}

	// Get processes for all clients
	for _, clientUser := range userOutput.ClientUserList {
		processFilter := body.EsFilter{
			Search: []body.EsSearchItem{
				{
					Name:     "clientId",
					Operator: "eq",
					Value:    clientUser.ClientId.TypedString(),
				},
			},
		}

		processResp, err := client.PostProcessSearch(ctx, processFilter)
		if err != nil {
			continue
		}

		processOutput, err := processResp.Output()
		if err != nil {
			continue
		}

		for _, process := range processOutput.Items {
			processInfo := map[string]interface{}{
				"id":       string(process.Id),
				"status":   string(process.Status),
				"created":  process.Created.Format("2006-01-02 15:04:05"),
			}
			
			// Add more info if needed

			allProcesses = append(allProcesses, processInfo)
		}
	}

	if len(allProcesses) == 0 {
		return map[string]interface{}{
			"processes": []interface{}{},
			"message":   "No running processes found",
		}, nil
	}

	return map[string]interface{}{
		"processes": allProcesses,
		"count":     len(allProcesses),
	}, nil
}