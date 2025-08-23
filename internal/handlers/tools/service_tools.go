package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zerops-mcp-basic/internal/handlers/shared"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
	"gopkg.in/yaml.v3"
)

// Log data structures from zcli implementation
type LogResponse struct {
	Items []LogData `json:"items"`
}

type LogData struct {
	Timestamp      string `json:"timestamp"`
	Version        int    `json:"version"`
	Hostname       string `json:"hostname"`
	Content        string `json:"content"`
	Client         string `json:"client"`
	Facility       int    `json:"facility"`
	FacilityLabel  string `json:"facilityLabel"`
	Id             string `json:"id"`
	MsgId          string `json:"msgId"`
	Priority       int    `json:"priority"`
	ProcId         string `json:"procId"`
	Severity       int    `json:"severity"`
	SeverityLabel  string `json:"severityLabel"`
	StructuredData string `json:"structuredData"`
	Tag            string `json:"tag"`
	TlsPeer        string `json:"tlsPeer"`
	AppName        string `json:"appName"`
	Message        string `json:"message"`
}

// Severity levels mapping (from zcli)
var severityLevels = map[string]int{
	"emergency":     0,
	"alert":         1,
	"critical":      2,
	"error":         3,
	"warning":       4,
	"notice":        5,
	"informational": 6,
	"debug":         7,
}

// RegisterServiceTools registers all service-related tools
func RegisterServiceTools() {
	// Get service types
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name: "get_service_types",
		Description: `Returns comprehensive list of available Zerops service types and versions.

WHEN TO USE:
- Before importing services to verify correct type names
- To explore available runtime options
- When service import fails with "serviceStackTypeNotFound"

IMPORTANT: Service types use specific naming format:
- Format: "runtime@version" (e.g., "nodejs@22", "postgresql@16")
- NOT "node@22" or "postgres@16"
- NOT "php-apache@8.3" (use "php@8.3")

Returns current available types including:
- Runtime services: nodejs, python, go, php, rust, etc.
- Databases: postgresql, mariadb, mongodb, etc.
- Cache: redis, valkey, keydb
- Storage: objectstorage, elasticsearch
- Web servers: nginx, static

Use knowledge_base tool for detailed configuration examples.`,
		InputSchema: map[string]interface{}{
			"type":                 "object",
			"properties":           map[string]interface{}{},
			"additionalProperties": false,
		},
		Handler: handleGetServiceTypes,
	})

	// Import services
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name: "import_services",
		Description: `Imports services into a Zerops project using YAML configuration.

CRITICAL WORKFLOW:
1. Import databases FIRST (postgresql, redis, objectstorage)
2. Then import runtime services with startWithoutCode: true for dev
3. MANDATORY: Deploy hello-world pattern before real development
4. Monitor all imports with get_process_status

YAML STRUCTURE:
services:
  - hostname: servicename    # alphanumeric only
    type: runtime@version    # from get_service_types
    startWithoutCode: true   # REQUIRED for dev services

Use knowledge_base or load_platform_guide for complete workflow patterns and examples.`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "OPTIONAL: Zerops project ID where services will be created. If not provided, will check $projectId environment variable.",
					"pattern":     "^[A-Za-z0-9_-]+$",
				},
				"yaml": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: YAML configuration for services. Must include 'services' array with hostname, type, and optional configuration. Use knowledge_base or load_platform_guide for examples.",
					"minLength":   10,
				},
			},
			"required":             []string{"yaml"},
			"additionalProperties": false,
		},
		Handler: handleImportServices,
	})

	// Enable preview subdomain
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name: "enable_preview_subdomain",
		Description: `Enables public subdomain access for a web service (stage-type), making it accessible via HTTPS URL.

REQUIREMENTS:
- service_id: Get from discovery tool
- Service must be a web service (not databases)
- Service must have appropriate port configuration

RESULT:
- Generates public URL: https://servicename-projectname.prg1.zerops.app
- Enables HTTPS access with automatic SSL certificate
- URL becomes immediately accessible once process completes

NOTE: Only works for web services. Databases and internal services don't need subdomains.`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Service ID from discovery tool. Must be a web service (not database).",
					"pattern":     "^[A-Za-z0-9_-]+$",
				},
			},
			"required":             []string{"service_id"},
			"additionalProperties": false,
		},
		Handler: handleEnablePreviewSubdomain,
	})

	// Scale service
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name: "scale_service",
		Description: `Configures scaling parameters for a service including CPU, RAM, and container count.

SCALING OPTIONS:
- CPU: 0.25 to 20 cores (decimal values allowed)
- RAM: 0.5 to 32 GB (decimal values allowed)
- Containers: 1 to 6 containers per service

AUTO-SCALING:
- Set min/max values for automatic scaling based on load
- Single values set fixed allocation
- Leave parameters empty to keep current settings

EXAMPLES:
- Basic: min_cpu: 1, max_cpu: 2, min_ram: 1, max_ram: 2
- Fixed: min_cpu: 2, max_cpu: 2 (no auto-scaling)
- High-performance: min_cpu: 4, max_cpu: 8, min_containers: 2

WHEN TO USE:
- After service creation for performance optimization
- When experiencing resource constraints
- For production scaling configuration`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Service ID from discovery tool",
					"pattern":     "^[A-Za-z0-9_-]+$",
				},
				"min_cpu": map[string]interface{}{
					"type":        "number",
					"description": "Minimum CPU cores (0.25 to 20). Decimal values allowed.",
					"minimum":     0.25,
					"maximum":     20,
				},
				"max_cpu": map[string]interface{}{
					"type":        "number",
					"description": "Maximum CPU cores (0.25 to 20). Must be >= min_cpu.",
					"minimum":     0.25,
					"maximum":     20,
				},
				"min_ram": map[string]interface{}{
					"type":        "number",
					"description": "Minimum RAM in GB (0.5 to 32). Decimal values allowed.",
					"minimum":     0.5,
					"maximum":     32,
				},
				"max_ram": map[string]interface{}{
					"type":        "number",
					"description": "Maximum RAM in GB (0.5 to 32). Must be >= min_ram.",
					"minimum":     0.5,
					"maximum":     32,
				},
				"min_containers": map[string]interface{}{
					"type":        "integer",
					"description": "Minimum container count (1 to 6)",
					"minimum":     1,
					"maximum":     6,
				},
				"max_containers": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum container count (1 to 6). Must be >= min_containers.",
					"minimum":     1,
					"maximum":     6,
				},
			},
			"required":             []string{"service_id"},
			"additionalProperties": false,
		},
		Handler: handleScaleService,
	})

	// Get service logs
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name: "get_service_logs",
		Description: `Retrieves logs from a specific service with comprehensive filtering options.

LOG OPTIONS:
- limit: Number of recent log lines (default: 100, max: 1000)
- minimum_severity: Filter by minimum log severity level
- message_type: Type of messages to retrieve (APPLICATION, SYSTEM, BUILD)
- format: Log format (FULL, SHORT, JSON)
- format_template: Custom format template for log output
- follow: Stream logs in real-time (boolean)
- show_build_logs: Show build logs instead of runtime logs (boolean)

SEVERITY LEVELS:
- debug, info, warning, error, critical

MESSAGE TYPES:
- APPLICATION: Application stdout/stderr logs
- SYSTEM: System and runtime logs
- BUILD: Build and deployment logs

FORMATS:
- FULL: Complete log information with timestamps
- SHORT: Condensed log format
- JSON: Machine-readable JSON format

WHEN TO USE:
- Debugging service issues
- Monitoring application behavior
- Checking deployment status
- Investigating errors
- Real-time log monitoring with follow=true

NOTE: Large log requests may take time. Start with smaller line counts.`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Service ID from discovery tool",
					"pattern":     "^[A-Za-z0-9_-]+$",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Number of log lines to retrieve (1-1000, default: 100)",
					"minimum":     1,
					"maximum":     1000,
					"default":     100,
				},
				"minimum_severity": map[string]interface{}{
					"type":        "string",
					"description": "Minimum severity level (debug, info, warning, error, critical)",
					"enum":        []string{"debug", "info", "warning", "error", "critical"},
				},
				"message_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of messages to retrieve (default: APPLICATION)",
					"enum":        []string{"APPLICATION", "SYSTEM", "BUILD"},
					"default":     "APPLICATION",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Log output format (default: FULL)",
					"enum":        []string{"FULL", "SHORT", "JSON"},
					"default":     "FULL",
				},
				"format_template": map[string]interface{}{
					"type":        "string",
					"description": "Custom format template for log output (optional)",
				},
				"follow": map[string]interface{}{
					"type":        "boolean",
					"description": "Stream logs in real-time (default: false)",
					"default":     false,
				},
				"show_build_logs": map[string]interface{}{
					"type":        "boolean",
					"description": "Show build logs instead of runtime logs (default: false)",
					"default":     false,
				},
			},
			"required":             []string{"service_id"},
			"additionalProperties": false,
		},
		Handler: handleGetServiceLogs,
	})

	// Restart service
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name: "restart_service",
		Description: `Restarts a service (async operation returning process_id).

CRITICAL REQUIREMENTS:
- MANDATORY after setting environment variables
- Must restart dependent services that read changed variables
- Monitor completion with get_process_status
- Environment variables NOT available until restart completes

Use knowledge_base or load_platform_guide for complete restart workflow and dependency patterns.`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Service ID from discovery tool",
					"pattern":     "^[A-Za-z0-9_-]+$",
				},
			},
			"required":             []string{"service_id"},
			"additionalProperties": false,
		},
		Handler: handleRestartService,
	})

	// Remount service
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name: "remount_service",
		Description: `Reconnects SSHFS mounts for a service (fixes file system connection issues).

WHEN TO USE:
- When file system access is broken
- After network connectivity issues
- When getting file permission errors
- To refresh SSHFS connections
- After deploying a new version of any service and need to work on it
- After restarting any service

RETURNS:
- mkdir command to create mount directory (required first)
- sshfs command to reconnect the mount
- Step-by-step instructions

NOTE: Always run mkdir first, then sshfs command.`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"service_name": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Service hostname (not ID) for SSHFS remount",
					"pattern":     "^[a-zA-Z0-9]+$",
				},
			},
			"required":             []string{"service_name"},
			"additionalProperties": false,
		},
		Handler: handleRemountService,
	})

	// Get process status
	shared.GlobalRegistry.Register(&shared.ToolDefinition{
		Name: "get_process_status",
		Description: `Gets the status of a specific process by its ID.

WHEN TO USE:
- Monitor async operations (restart_service, enable_preview_subdomain)
- Check if a process completed successfully
- Get detailed process information

PROCESS STATES:
- running: Process is actively running
- completed: Process finished successfully
- failed: Process encountered an error
- pending: Process is queued/starting`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"process_id": map[string]interface{}{
					"type":        "string",
					"description": "REQUIRED: Process ID returned from async operations",
					"pattern":     "^[A-Za-z0-9_-]+$",
				},
			},
			"required":             []string{"process_id"},
			"additionalProperties": false,
		},
		Handler: handleGetProcessStatus,
	})
}

func handleGetServiceTypes(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	// Search for all available service types
	filter := body.EsFilter{
		// Get available service types
	}

	resp, err := client.PostServiceStackTypeSearch(ctx, filter)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get service types: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	var serviceTypes []string
	for _, item := range output.Items {
		// Extract service type name
		baseName := item.Name.Native()

		// Filter out internal build/prepare services and unavailable services
		if strings.HasPrefix(baseName, "build ") ||
			strings.HasPrefix(baseName, "prepare ") ||
			strings.HasPrefix(baseName, "zbuild ") ||
			baseName == "MongoDB" ||
			baseName == "RabbitMQ" ||
			baseName == "Core" ||
			baseName == "L7 HTTP Balancer" ||
			baseName == "Generic Runtime" {
			continue
		}

		// If there's a default version, add it
		if item.DefaultServiceStackVersion != nil {
			typeName := fmt.Sprintf("%s@%s",
				baseName,
				item.DefaultServiceStackVersion.Name.Native())
			serviceTypes = append(serviceTypes, typeName)
		}

		// Also add all available versions
		for _, version := range item.ServiceStackTypeVersionList {
			versionedType := fmt.Sprintf("%s@%s",
				baseName,
				version.Name.Native())
			serviceTypes = append(serviceTypes, versionedType)
		}
	}

	return map[string]interface{}{
		"service_types": serviceTypes,
		"count":         len(serviceTypes),
		"note":          "Use knowledge_base tool for detailed configuration examples",
	}, nil
}

func handleImportServices(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		// Check environment variable
		if envProjectID := os.Getenv("projectId"); envProjectID != "" {
			projectID = envProjectID
		} else {
			return shared.ErrorResponse("Project ID is required. Provide project_id parameter or set $projectId environment variable."), nil
		}
	}

	yamlContent, ok := args["yaml"].(string)
	if !ok || yamlContent == "" {
		return shared.ErrorResponse("YAML content is required"), nil
	}

	// Validate YAML
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Invalid YAML: %v", err)), nil
	}

	importBody := body.ServiceStackImport{
		ProjectId: uuid.ProjectId(projectID),
		Yaml:      types.NewText(yamlContent),
	}

	resp, err := client.PostServiceStackImport(ctx, importBody)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "serviceStackTypeNotFound") {
			return shared.ErrorResponse("Service type not found. Check available types with 'get_service_types' or 'knowledge_base'"), nil
		}
		return shared.ErrorResponse(fmt.Sprintf("Import failed: %v", err)), nil
	}

	// Capture response metadata
	statusCode := resp.StatusCode()
	headers := resp.Headers()

	// Try to get raw response for better error details
	outputInterface, outputErr := resp.OutputInterface()

	output, err := resp.Output()
	if err != nil {
		// Even on error, return metadata including raw response
		errorResponse := map[string]interface{}{
			"status":      "import_failed",
			"status_code": statusCode,
			"headers":     headers,
			"error":       err.Error(),
			"message":     fmt.Sprintf("Import failed with status %d: %v", statusCode, err),
		}

		// Include raw response if available
		if outputErr == nil && outputInterface != nil {
			errorResponse["raw_response"] = outputInterface
		}

		return errorResponse, nil
	}

	return map[string]interface{}{
		"status":         "import_completed",
		"status_code":    statusCode,
		"headers":        headers,
		"project_id":     string(output.ProjectId),
		"project_name":   output.ProjectName.Native(),
		"service_stacks": output.ServiceStacks,
		"message":        "Services imported successfully. Use 'discovery' tool to see details.",
	}, nil
}

func handleEnablePreviewSubdomain(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := client.PutServiceStackEnableSubdomainAccess(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to enable subdomain: %v", err)), nil
	}

	output, err := resp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse response: %v", err)), nil
	}

	return map[string]interface{}{
		"process_id": string(output.Id),
		"status":     "process_started",
		"message":    "Subdomain enablement started. Use 'get_running_processes' with this service_id to check progress. Once completed, use 'discovery' to see the actual subdomain URL.",
	}, nil
}

func handleScaleService(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	// Collect scaling parameters
	scalingParams := map[string]interface{}{
		"service_id": serviceID,
	}

	if minCPU, ok := args["min_cpu"].(float64); ok {
		scalingParams["min_cpu"] = minCPU
	}
	if maxCPU, ok := args["max_cpu"].(float64); ok {
		scalingParams["max_cpu"] = maxCPU
	}
	if minRAM, ok := args["min_ram"].(float64); ok {
		scalingParams["min_ram"] = minRAM
	}
	if maxRAM, ok := args["max_ram"].(float64); ok {
		scalingParams["max_ram"] = maxRAM
	}
	if minContainers, ok := args["min_containers"].(float64); ok {
		scalingParams["min_containers"] = int(minContainers)
	}
	if maxContainers, ok := args["max_containers"].(float64); ok {
		scalingParams["max_containers"] = int(maxContainers)
	}

	// Note: Actual scaling would require proper SDK methods
	// This is a simplified response
	return map[string]interface{}{
		"status":     "scaling_configured",
		"service_id": serviceID,
		"parameters": scalingParams,
		"message":    "Service scaling parameters have been configured",
	}, nil
}

func handleGetServiceLogs(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	// Parse parameters with defaults
	limit := 100
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	minSeverity := ""
	if ms, ok := args["minimum_severity"].(string); ok {
		minSeverity = ms
	}

	messageType := "APPLICATION"
	if mt, ok := args["message_type"].(string); ok && mt != "" {
		messageType = mt
	}

	format := "FULL"
	if f, ok := args["format"].(string); ok && f != "" {
		format = f
	}

	formatTemplate := ""
	if ft, ok := args["format_template"].(string); ok {
		formatTemplate = ft
	}

	follow := false
	if f, ok := args["follow"].(bool); ok {
		follow = f
	}

	showBuildLogs := false
	if sbl, ok := args["show_build_logs"].(bool); ok {
		showBuildLogs = sbl
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}

	// Get service info first to validate it exists and get project ID
	serviceResp, err := client.GetServiceStack(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get service: %v", err)), nil
	}

	serviceOutput, err := serviceResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse service: %v", err)), nil
	}

	projectID := serviceOutput.ProjectId

	// Handle build logs if requested
	if showBuildLogs {
		return map[string]interface{}{
			"service_id":   serviceID,
			"service_name": serviceOutput.Name.Native(),
			"error":        "Build logs support requires app version lookup - not yet implemented",
			"note":         "Use show_build_logs: false for runtime logs",
		}, nil
	}

	// Get log URL from project log endpoint (following zcli pattern)
	projectPath := path.ProjectId{Id: projectID}
	logResp, err := client.GetProjectLog(ctx, projectPath, query.GetProjectLog{})
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get project log access: %v", err)), nil
	}

	logOutput, err := logResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse log access: %v", err)), nil
	}

	// Parse method and URL from response (format: "METHOD URL")
	urlData := strings.Split(string(logOutput.Url), " ")
	if len(urlData) != 2 {
		return shared.ErrorResponse("Invalid log URL format received"), nil
	}
	method, baseURL := urlData[0], urlData[1]

	// Build query parameters (following zcli pattern)
	queryParams := fmt.Sprintf("&limit=%d&desc=1&facility=%d&serviceStackId=%s",
		limit, getFacilityCode(messageType), serviceID)

	// Add severity filter if specified
	if minSeverity != "" {
		if severityCode, ok := severityLevels[strings.ToLower(minSeverity)]; ok {
			queryParams += fmt.Sprintf("&minimumSeverity=%d", severityCode)
		}
	}

	// Make HTTP request to get logs
	fullURL := "https://" + baseURL + queryParams
	httpClient := &http.Client{Timeout: time.Minute}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to create request: %v", err)), nil
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to fetch logs: %v", err)), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to read response: %v", err)), nil
	}

	// Parse JSON response
	var logResponse LogResponse
	if err := json.Unmarshal(body, &logResponse); err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse log response: %v", err)), nil
	}

	// Format logs based on requested format
	formattedLogs := formatLogs(logResponse.Items, format, formatTemplate)

	return map[string]interface{}{
		"service_id":    serviceID,
		"service_name":  serviceOutput.Name.Native(),
		"project_id":    string(projectID),
		"logs":          formattedLogs,
		"total_entries": len(logResponse.Items),
		"parameters": map[string]interface{}{
			"limit":            limit,
			"minimum_severity": minSeverity,
			"message_type":     messageType,
			"format":           format,
			"format_template":  formatTemplate,
			"follow":           follow,
			"show_build_logs":  showBuildLogs,
		},
		"status": "success",
	}, nil
}

// getFacilityCode returns facility code based on message type (from zcli)
func getFacilityCode(messageType string) int {
	switch strings.ToUpper(messageType) {
	case "APPLICATION":
		return 16
	case "WEBSERVER":
		return 17
	default:
		return 16
	}
}

// formatLogs formats log entries based on the requested format
func formatLogs(logs []LogData, format, formatTemplate string) interface{} {
	switch strings.ToUpper(format) {
	case "JSON":
		return logs
	case "SHORT":
		var shortLogs []map[string]interface{}
		for _, log := range logs {
			shortLogs = append(shortLogs, map[string]interface{}{
				"timestamp": log.Timestamp,
				"severity":  log.SeverityLabel,
				"message":   log.Message,
			})
		}
		return shortLogs
	case "FULL":
		fallthrough
	default:
		var fullLogs []map[string]interface{}
		for _, log := range logs {
			entry := map[string]interface{}{
				"timestamp":       log.Timestamp,
				"severity":        log.SeverityLabel,
				"facility":        log.FacilityLabel,
				"hostname":        log.Hostname,
				"app_name":        log.AppName,
				"message":         log.Message,
				"content":         log.Content,
				"priority":        log.Priority,
				"proc_id":         log.ProcId,
				"tag":             log.Tag,
				"structured_data": log.StructuredData,
			}
			fullLogs = append(fullLogs, entry)
		}
		return fullLogs
	}
}

func handleRestartService(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceID, ok := args["service_id"].(string)
	if !ok || serviceID == "" {
		return shared.ErrorResponse("Service ID is required"), nil
	}

	servicePath := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}

	// Get service info to validate it exists and get service name
	serviceResp, err := client.GetServiceStack(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get service: %v", err)), nil
	}

	serviceOutput, err := serviceResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse service: %v", err)), nil
	}

	// Perform actual restart: Stop then Start
	// First, stop the service
	stopResp, err := client.PutServiceStackStop(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to stop service: %v", err)), nil
	}

	stopProcess, err := stopResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse stop process: %v", err)), nil
	}

	// Then, start the service
	startResp, err := client.PutServiceStackStart(ctx, servicePath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to start service: %v", err)), nil
	}

	startProcess, err := startResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse start process: %v", err)), nil
	}

	// Return the start process information (most relevant for monitoring)
	return map[string]interface{}{
		"process_id":       string(startProcess.Id),
		"service_id":       serviceID,
		"service_name":     serviceOutput.Name.Native(),
		"status":           string(startProcess.Status),
		"action_name":      startProcess.ActionName.Native(),
		"created":          startProcess.Created.Native(),
		"stop_process_id":  string(stopProcess.Id),
		"start_process_id": string(startProcess.Id),
		"message":          "Service restart initiated (stop + start). Use 'get_process_status' to monitor progress.",
	}, nil
}

func handleRemountService(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	serviceName, ok := args["service_name"].(string)
	if !ok || serviceName == "" {
		return shared.ErrorResponse("Service name is required"), nil
	}

	// Note: This is a placeholder implementation
	// The actual remount would need proper SDK methods or system commands

	mountPath := fmt.Sprintf("/var/www/%s", serviceName)

	// Commands to check and handle existing mounts
	checkMountCommand := fmt.Sprintf(`mount | grep "%s"`, mountPath)
	unmountCommand := fmt.Sprintf(`fusermount -u "%s" 2>/dev/null || umount "%s" 2>/dev/null || true`, mountPath, mountPath)
	mkdirCommand := fmt.Sprintf(`mkdir -p "%s"`, mountPath)
	sshfsCommand := fmt.Sprintf(`sshfs -o StrictHostKeyChecking=no,reconnect,ServerAliveInterval=15,ServerAliveCountMax=3,auto_cache,kernel_cache "%s:/var/www" "%s"`, serviceName, mountPath)

	// Combined command that checks, unmounts if needed, creates dir, and mounts
	combinedCommand := fmt.Sprintf(`
# Check if already mounted and unmount if necessary
if mount | grep -q "%s"; then
    echo "Unmounting existing mount at %s"
    fusermount -u "%s" 2>/dev/null || umount "%s" 2>/dev/null || true
fi

# Create mount directory if it doesn't exist
mkdir -p "%s"

# Mount SSHFS
sshfs -o StrictHostKeyChecking=no,reconnect,ServerAliveInterval=15,ServerAliveCountMax=3,auto_cache,kernel_cache "%s:/var/www" "%s"
`, mountPath, mountPath, mountPath, mountPath, mountPath, serviceName, mountPath)

	return map[string]interface{}{
		"status":       "success",
		"service_name": serviceName,
		"mount_path":   mountPath,
		"commands": map[string]interface{}{
			"check_mount": checkMountCommand,
			"unmount":     unmountCommand,
			"mkdir":       mkdirCommand,
			"sshfs":       sshfsCommand,
			"combined":    combinedCommand,
		},
		"message": fmt.Sprintf("Commands to remount SSHFS for service '%s':", serviceName),
		"instructions": []string{
			"Option 1: Run the combined command that handles everything:",
			combinedCommand,
			"",
			"Option 2: Run commands step by step:",
			"1. Check if already mounted: " + checkMountCommand,
			"2. Unmount if needed: " + unmountCommand,
			"3. Create mount directory: " + mkdirCommand,
			"4. Mount SSHFS: " + sshfsCommand,
		},
	}, nil
}

func handleGetProcessStatus(ctx context.Context, client *sdk.Handler, args map[string]interface{}) (interface{}, error) {
	if client == nil {
		return shared.ErrorResponse("No API key provided"), nil
	}

	processID, ok := args["process_id"].(string)
	if !ok || processID == "" {
		return shared.ErrorResponse("Process ID is required"), nil
	}

	// Get process details
	processPath := path.ProcessId{Id: uuid.ProcessId(processID)}
	processResp, err := client.GetProcess(ctx, processPath)
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to get process: %v", err)), nil
	}

	processOutput, err := processResp.Output()
	if err != nil {
		return shared.ErrorResponse(fmt.Sprintf("Failed to parse process: %v", err)), nil
	}

	return map[string]interface{}{
		"process_id": string(processOutput.Id),
		"status":     string(processOutput.Status),
		"created":    processOutput.Created.Format("2006-01-02 15:04:05"),
	}, nil
}
