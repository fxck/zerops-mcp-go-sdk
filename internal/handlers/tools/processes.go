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
		Description: `Retrieves information about running processes, optionally filtered by service.

PROCESS INFORMATION:
- Process IDs and status
- Creation timestamps
- Associated service information
- Process state and metadata

FILTERING OPTIONS:
- No service_id: Returns all processes across all services (limited to 50)
- With service_id: Returns processes only for specified service
- Use limit parameter to control response size

PROCESS STATES:
- running: Process is actively running
- completed: Process finished successfully
- failed: Process encountered an error
- pending: Process is queued/starting

WHEN TO USE:
- Monitoring service deployments
- Checking process status after operations
- Debugging service issues
- Tracking long-running operations`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "OPTIONAL: Service ID to filter processes. If omitted, returns all processes.",
					"pattern":     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "OPTIONAL: Maximum number of processes to return (1-100, default: 20)",
					"minimum":     1,
					"maximum":     100,
					"default":     20,
				},
			},
			"additionalProperties": false,
		},
		Handler: handleGetRunningProcesses,
	})
}

func handleGetRunningProcesses(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	// Get limit parameter
	limit := 20 // default
	if l, ok := args["limit"].(float64); ok && l > 0 && l <= 100 {
		limit = int(l)
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

		// Get RUNNING processes for this service
		processFilter := body.EsFilter{
			Search: []body.EsSearchItem{
				{
					Name:     "serviceStackId",
					Operator: "eq",
					Value:    types.String(serviceID),
				},
				{
					Name:     "status",
					Operator: "eq",
					Value:    types.String("running"),
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
		for i, process := range processOutput.Items {
			if i >= limit {
				break
			}
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

	// Get RUNNING processes for all clients
	for _, clientUser := range userOutput.ClientUserList {
		if len(allProcesses) >= limit {
			break
		}
		processFilter := body.EsFilter{
			Search: []body.EsSearchItem{
				{
					Name:     "clientId",
					Operator: "eq",
					Value:    clientUser.ClientId.TypedString(),
				},
				{
					Name:     "status",
					Operator: "eq",
					Value:    types.String("running"),
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
			if len(allProcesses) >= limit {
				break
			}
			processInfo := map[string]interface{}{
				"id":      string(process.Id),
				"status":  string(process.Status),
				"created": process.Created.Format("2006-01-02 15:04:05"),
			}
			
			allProcesses = append(allProcesses, processInfo)
		}
	}

	if len(allProcesses) == 0 {
		return map[string]interface{}{
			"processes": []interface{}{},
			"message":   "No running processes found",
		}, nil
	}

	result := map[string]interface{}{
		"processes": allProcesses,
		"count":     len(allProcesses),
		"limit":     limit,
	}
	
	if len(allProcesses) == limit {
		result["note"] = fmt.Sprintf("Results limited to %d processes. Use 'limit' parameter to see more or filter by service_id.", limit)
	}
	
	return result, nil
}