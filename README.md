# Zerops MCP Server

MCP server for managing Zerops infrastructure through AI assistants like Claude.

## Live Testing Server

Try immediately without installation:

```bash
claude mcp add --transport http zerops https://mcp-16cb-8080.prg1.zerops.app/mcp \
  --header "Authorization: Bearer your-zerops-api-key"
```

Get your API key: [app.zerops.io/settings/token-management](https://app.zerops.io/settings/token-management)

## Local Mode (stdio)

Run the server locally on your machine.

### Installation

#### Quick Install Script

```bash
# macOS/Linux
curl -sSL https://raw.githubusercontent.com/fxck/zerops-mcp-go-sdk/main/install.sh | sh

# Windows PowerShell
irm https://raw.githubusercontent.com/fxck/zerops-mcp-go-sdk/main/install.ps1 | iex
```

#### Pre-built Binaries

Download from [GitHub Releases](https://github.com/fxck/zerops-mcp-go-sdk/releases):

- **Windows**: `zerops-mcp-win-x64.exe`
- **macOS Intel**: `zerops-mcp-darwin-amd64`
- **macOS Apple Silicon**: `zerops-mcp-darwin-arm64`
- **Linux AMD64**: `zerops-mcp-linux-amd64`
- **Linux 386**: `zerops-mcp-linux-i386`

### Add to Claude Code

```bash
# Quick setup (after install script)
export ZEROPS_API_KEY="your-api-key"

# macOS/Linux - uses install script path
claude mcp add zerops -s user ~/.local/bin/zerops-mcp

# Windows - uses install script path  
claude mcp add zerops -s user ~\.zerops\mcp\bin\zerops-mcp.exe
```

Or manual config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "zerops": {
      "command": "/path/to/zerops-mcp",
      "env": {
        "ZEROPS_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Remote Mode (HTTP)

Host your own MCP server.

### Build

```bash
# Clone and build
git clone https://github.com/fxck/zerops-mcp-go-sdk
cd zerops-mcp-go-sdk
make all

# Or build for current platform only
go build -o zerops-mcp cmd/mcp-server/main.go
```

### Deploy

```bash
# Start server
./zerops-mcp --transport http --port 8080

# Test the server
curl -X POST http://localhost:8080/ \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

### Add to Claude Code

```bash
claude mcp add --transport http zerops https://your-server.com \
  --header "Authorization: Bearer your-api-key"
```

## Available Tools

The Zerops MCP SDK provides 14 tools for managing Zerops projects, services, and deployments through AI assistants like Claude.

### Quick Reference

| Tool | Current Description | Required Parameters |
|------|-------------------|-------------------|
| `discovery` | ESSENTIAL FIRST STEP: Discovers all services in a project with their IDs, hostnames, service types, and environment variable availability. CRITICAL: Requires a project ID. To get the project ID, the agent can run 'echo $projectId' in the container environment. Returns condensed data about all services with their unique IDs, service hostnames and types, available environment variables at project and service level, and current project configuration. Always use this tool first to understand the project structure before performing other operations. | `project_id` |
| `get_service_types` | Provides a comprehensive list of all available Zerops service types and their versions. Returns all service types available in Zerops platform with their specific versions (e.g., nodejs@22, python@3.12, postgresql@16, etc.). Use this before import_services to ensure you're using valid service type names. Essential for understanding what services you can create in your project. | None |
| `import_services` | Creates services in a Zerops project using YAML configuration. This tool initiates the service creation process and returns immediately with a process status. The actual service creation is asynchronous, so use the 'discovery' tool to check when services are fully created. Supports all Zerops service types and configurations. | `project_id`, `yaml` |
| `enable_preview_subdomain` | Enables public subdomain access for a specific service (HTTP routing). This is an asynchronous operation that sets up public web access for your service. Use with web services (nodejs, python, php, go, etc.) to make them accessible from the internet. Returns a process ID to monitor the operation status. | `service_id` |
| `scale_service` | Configures service scaling parameters including CPU, RAM, and container counts. Allows you to set minimum and maximum resource limits for auto-scaling or fixed resource allocation. Essential for production workload management and cost optimization. | `service_id` |
| `get_service_logs` | Retrieves logs from a specific service with optional filtering by time period and line count. Useful for debugging applications, monitoring service health, and troubleshooting deployment issues. Supports time-based filtering (1h, 24h, 7d) and line limiting. | `service_id` |
| `set_project_env` | Sets project-level environment variables that are available to all services within the project. Use this for shared configuration like database URLs, API keys, or other global settings that multiple services need access to. Changes trigger service restarts automatically. | `project_id`, `key`, `value` |
| `set_service_env` | Sets service-specific environment variables for individual services. These variables are only available to the specific service and override project-level variables with the same name. Good for service-specific configuration like ports or service-unique settings. | `service_id`, `key`, `value` |
| `get_running_processes` | Monitors running processes across the project or for a specific service. Shows process status, creation time, and service association. Essential for monitoring deployment progress, checking async operations status, and debugging service issues. Can be filtered by service for focused monitoring. | None |
| `restart_service` | Initiates a service restart operation (asynchronous). Useful after environment variable changes, configuration updates, or when a service becomes unresponsive. Returns a process ID to monitor the restart progress using get_process_status or get_running_processes. | `service_id` |
| `remount_service` | Fixes SSHFS mount connection issues by providing the correct remount command. When file system access is broken or after network connectivity issues, this tool generates the proper SSHFS command to reconnect the file system. Copy and execute the returned command in your terminal. | `service_name` |
| `get_process_status` | Retrieves the status of a specific process by its ID. Essential for monitoring asynchronous operations like service restarts, subdomain enablement, or service imports. Helps track the progress and completion of background processes. | `process_id` |
| `load_platform_guide` | Loads path-specific workflow guides for different project scenarios (fetched from GitHub with 10min cache). Provides step-by-step workflows for: fresh_project (starting new projects), existing_service (working with existing services), add_services (adding services to existing projects). | `path_type` |
| `knowledge_base` | Provides comprehensive YAML examples and deployment patterns for specific runtimes. Returns runtime-specific zerops.yml examples, service configuration patterns, and deployment best practices. For unsupported runtimes, returns Node.js pattern as reference with guidance to adapt. | `runtime` |

## Common Workflows

### 1. Initial Project Setup
```
1. discovery() - Understand current project state
2. get_service_types() - See available service types
3. knowledge_base("nodejs") - Get YAML examples
4. import_services(yaml: "...") - Create services
5. discovery() - Verify services were created
```

### 2. Enable Public Access
```
1. discovery() - Get service IDs
2. enable_preview_subdomain(service_id: "...") - Start enablement
3. get_running_processes(service_id: "...") - Monitor progress
4. discovery() - Get final subdomain URL
```

### 3. Environment Configuration
```
1. discovery() - See current environment variables
2. set_project_env(key: "DATABASE_URL", value: "...") - Shared config
3. set_service_env(service_id: "...", key: "PORT", value: "3000") - Service-specific
```

### 4. Monitoring and Scaling
```
1. get_service_logs(service_id: "...", lines: 50) - Check logs
2. scale_service(service_id: "...", min_cpu: 2, max_cpu: 4) - Scale up
3. get_running_processes() - Monitor operations
```

## Error Handling

### Common Errors:
- **"Project ID is required. Run 'echo $projectId' in the container to get it."**: Agent must run 'echo $projectId' to get the project ID
- **"serviceStackTypeNotFound"**: Use `get_service_types` to verify correct type names
- **"Invalid hostname"**: Use alphanumeric characters only, no special characters
- **"Response exceeds maximum tokens"**: Use `limit` parameter in `get_running_processes`

### Best Practices:
1. Always start with `discovery` to understand current state
2. Use `get_service_types` before `import_services`
3. Use `knowledge_base` for runtime-specific examples
4. Monitor async operations with `get_running_processes`
5. Set up environment variables after service creation
6. Use appropriate limits to prevent large responses

## Environment Variables

- `$projectId`: Project UUID available in the container environment. Agents can run 'echo $projectId' to get the current project ID and pass it to tools that require project_id parameter.

## Prerequisites

- **API Key**: Get from [app.zerops.io/settings/token-management](https://app.zerops.io/settings/token-management)
- **Go 1.21+**: For building from source

## Development

```bash
make test        # Run tests
make lint        # Run linter
make all         # Build all platforms
make help        # Show all targets
```

## Notes

- All UUIDs follow the pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
- Many operations are asynchronous and return process IDs for monitoring
- Use `discovery` to get final state after async operations complete
- Response sizes are limited to prevent token overflow (use pagination/filtering)

## Links

[Zerops Docs](https://docs.zerops.io) | [MCP Spec](https://modelcontextprotocol.io) | [Issues](https://github.com/fxck/zerops-mcp-go-sdk/issues) | [Discord](https://discord.com/invite/WDvCZ54)

## License

MIT