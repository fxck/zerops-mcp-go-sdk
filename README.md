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

The Zerops MCP SDK provides comprehensive tools for managing Zerops projects, services, and deployments through AI assistants like Claude.

### Quick Reference

#### 🔍 Discovery & Information

**`discovery`** - Get project overview and service details
- **Required**: `project_id`
- **Optional**: `service_id`, `service_name` (filter to single service)

<details>
<summary>Example Output</summary>

```json
{
  "count": 7,
  "project": {
    "env_keys": [
      "envIsolation", "ZEROPS_API_KEY", "staticCdnUrl",
      "sshIsolation", "zeropsSubdomainHost", "storageCdnUrl",
      "ZEROPS_TOKEN", "zeropsSubdomainString", "apiCdnUrl"
    ],
    "id": "eOc4woejQjC5KhohvkVKPQ",
    "name": "zagent3"
  },
  "services": [
    {
      "env_keys": ["zeropsSubdomain", "hostname", "VSCODE_PASSWORD"],
      "hostname": "zagent",
      "id": "WAlvwg9GQ3qBQAi37Gts5A",
      "process_count": 0,
      "status": "ACTIVE",
      "type": "nodejs_v22",
      "active_version": {
        "created": "2025-08-22 22:36:10 +0000 UTC",
        "id": "epNDMsLRQSqeTl4lXgEwpw",
        "status": "ACTIVE",
        "updated": "2025-08-22 22:38:56 +0000 UTC"
      }
    },
    {
      "env_keys": ["portTls", "serviceId", "projectId", "superUser", "user"],
      "hostname": "db",
      "id": "l17tXjvhSzKYcAERRXuqFw",
      "process_count": 0,
      "status": "ACTIVE",
      "type": "postgresql_v17"
    }
  ]
}
```
</details>

**`get_service_types`** - List all available service types

<details>
<summary>Example Output</summary>

```json
{
  "count": 67,
  "note": "Use knowledge_base tool for detailed configuration examples",
  "service_types": [
    "PostgreSQL@postgresql@17",
    "PostgreSQL@postgresql@16",
    "Node.js@nodejs@22",
    "Node.js@nodejs@20",
    "Python@python@3.12",
    "Go@go@1.23",
    "PHP@php@8.3",
    "Static@static",
    "Docker@docker@26.1",
    "MariaDB@mariadb@10.6",
    "KeyDB@keydb@6",
    "Object storage@object-storage",
    "...and 55 more types"
  ]
}
```
</details>

#### 🚀 Service Management

**`import_services`** - Create new services from YAML
- **Required**: `project_id`, `yaml`

<details>
<summary>Example Output</summary>

```json
{
  "status": "import_completed",
  "project_id": "eOc4woejQjC5KhohvkVKPQ",
  "project_name": "my-project",
  "services": [
    {
      "id": "WAlvwg9GQ3qBQAi37Gts5A",
      "hostname": "nodejs-app",
      "process_count": 1,
      "import_process_id": "abc123def456"
    },
    {
      "id": "XBmwxh0HQ4rCRAj48Hut6B",
      "hostname": "postgres-db",
      "process_count": 1,
      "import_process_id": "def789ghi012"
    }
  ],
  "count": 2,
  "message": "Services imported successfully. Use 'discovery' tool to get full details."
}
```
</details>

**`restart_service`** - Restart a service
- **Required**: `service_id`

<details>
<summary>Example Output</summary>

```json
{
  "process_id": "Rxdb0wOBTk6xsqgEVoux8A",
  "service_id": "WAlvwg9GQ3qBQAi37Gts5A",
  "service_name": "zagent",
  "status": "PENDING",
  "action_name": "stack.start",
  "created": "2025-08-23T08:53:22.996Z",
  "stop_process_id": "rGYRmLQaRzmlpSGlBBE8Xg",
  "start_process_id": "Rxdb0wOBTk6xsqgEVoux8A",
  "message": "Service restart initiated (stop + start). Use 'get_process_status' to monitor progress."
}
```
</details>

**`scale_service`** - Configure service resources
- **Required**: `service_id`
- **Optional**: `min_cpu`, `max_cpu`, `min_ram`, `max_ram`, `min_replicas`, `max_replicas`

<details>
<summary>Example Output</summary>

```json
{
  "status": "success",
  "service_name": "webapp",
  "scaling_config": {
    "min_cpu": 1,
    "max_cpu": 4,
    "min_ram": 512,
    "max_ram": 2048
  }
}
```
</details>

#### 🌐 Network & Access

**`enable_preview_subdomain`** - Enable public web access
- **Required**: `service_id`

<details>
<summary>Example Output</summary>

```json
{
  "status": "success",
  "process_id": "ghi789jkl012",
  "service_name": "webapp",
  "message": "Subdomain enablement started. Use 'get_running_processes' to check progress."
}
```
</details>

**`remount_service`** - Fix SSHFS mount issues
- **Required**: `service_name`

<details>
<summary>Example Output</summary>

```json
{
  "status": "success",
  "service_name": "webapp",
  "mount_path": "/var/www/webapp",
  "commands": {
    "check_mount": "mount | grep \"/var/www/webapp\"",
    "combined": "\n# Check if already mounted and unmount if necessary\nif mount | grep -q \"/var/www/webapp\"; then\n    echo \"Unmounting existing mount at /var/www/webapp\"\n    fusermount -u \"/var/www/webapp\" 2>/dev/null || umount \"/var/www/webapp\" 2>/dev/null || true\nfi\n\n# Create mount directory if it doesn't exist\nmkdir -p \"/var/www/webapp\"\n\n# Mount SSHFS\nsshfs -o StrictHostKeyChecking=no,reconnect,ServerAliveInterval=15,ServerAliveCountMax=3,auto_cache,kernel_cache \"webapp:/var/www\" \"/var/www/webapp\"\n",
    "mkdir": "mkdir -p \"/var/www/webapp\"",
    "sshfs": "sshfs -o StrictHostKeyChecking=no,reconnect,ServerAliveInterval=15,ServerAliveCountMax=3,auto_cache,kernel_cache \"webapp:/var/www\" \"/var/www/webapp\"",
    "unmount": "fusermount -u \"/var/www/webapp\" 2>/dev/null || umount \"/var/www/webapp\" 2>/dev/null || true"
  },
  "instructions": [
    "Option 1: Run the combined command that handles everything:",
    "Option 2: Run commands step by step:",
    "1. Check if already mounted: mount | grep \"/var/www/webapp\"",
    "2. Unmount if needed: fusermount -u \"/var/www/webapp\" 2>/dev/null || umount \"/var/www/webapp\" 2>/dev/null || true"
  ],
  "message": "Commands to remount SSHFS for service 'webapp':"
}
```
</details>

#### ⚙️ Environment Variables

**`set_project_env`** - Set project-wide environment variable
- **Required**: `project_id`, `key`, `value`

<details>
<summary>Example Output</summary>

```json
{
  "key": "TEST_VAR",
  "message": "Project environment variable 'TEST_VAR' has been set",
  "process_id": "1L5GLsraTYaW2fdADWhTyw",
  "status": "env_var_set"
}
```
</details>

**`set_service_env`** - Set service-specific environment variable
- **Required**: `service_id`, `key`, `value`

<details>
<summary>Example Output</summary>

```json
{
  "status": "success",
  "service_name": "webapp",
  "key": "PORT",
  "value": "3000",
  "message": "Service environment variable set successfully."
}
```
</details>

#### 📊 Monitoring & Logs

**`get_service_logs`** - Retrieve service logs
- **Required**: `service_id`
- **Optional**: `limit`, `minimum_severity`, `message_type`, `format`, `show_build_logs`

<details>
<summary>Example Output</summary>

```json
{
  "service_id": "WAlvwg9GQ3qBQAi37Gts5A",
  "service_name": "zagent",
  "project_id": "eOc4woejQjC5KhohvkVKPQ",
  "logs": [
    {
      "timestamp": "2025-08-23T06:11:11.023482Z",
      "severity": "Informational",
      "message": "[06:11:11] [10.12.228.4][e6459539][ExtensionHostConnection] <1092> Extension Host Process exited with code: 0, signal: null."
    },
    {
      "timestamp": "2025-08-23T06:06:10.405744Z",
      "severity": "Informational",
      "message": "[06:06:10] [10.12.228.4][3fa281c6][ManagementConnection] Unknown reconnection token (seen before)."
    }
  ],
  "total_entries": 2,
  "parameters": {
    "limit": 2,
    "format": "SHORT",
    "message_type": "APPLICATION",
    "show_build_logs": false
  },
  "status": "success"
}
```
</details>

**`get_running_processes`** - Monitor active processes
- **Optional**: `service_id`, `limit`

<details>
<summary>Example Output</summary>

```json
{
  "message": "No running processes found",
  "processes": []
}
```
</details>

**`get_process_status`** - Check specific process status
- **Required**: `process_id`

<details>
<summary>Example Output</summary>

```json
{
  "process_id": "Rxdb0wOBTk6xsqgEVoux8A",
  "status": "RUNNING",
  "created": "2025-08-23 08:53:22"
}
```
</details>

#### 📚 Knowledge & Guides

**`knowledge_base`** - Get configuration examples for services
- **Required**: `runtime` 
- **Different modes**: `service_import` (service import YAML), `database_patterns` (managed services), `nodejs` (runtime deployment config), etc.

<details>
<summary>Service Import Patterns (`runtime: "service_import"`)</summary>

```json
{
  "title": "Complete Service Import YAML Patterns",
  "patterns": {
    "basic_runtime": "# Basic runtime service\nservices:\n  - hostname: app\n    type: nodejs@22\n    startWithoutCode: true\n    enableSubdomainAccess: true",
    "database_services": "# Database services (import FIRST)\nservices:\n  - hostname: db\n    type: postgresql@17\n    mode: NON_HA\n  - hostname: cache\n    type: valkey@7.2\n    mode: NON_HA"
  },
  "field_reference": {
    "hostname": "REQUIRED: Unique service identifier, alphanumeric only",
    "type": "REQUIRED: Service type and version (from get_service_types)",
    "startWithoutCode": "CRITICAL for dev services - prevents auto-start",
    "enableSubdomainAccess": "Enable public access via Zerops subdomain"
  }
}
```
</details>

<details>
<summary>Database Patterns (`runtime: "database_patterns"`)</summary>

```json
{
  "databases": {
    "postgresql": "# PostgreSQL database\nservices:\n  - hostname: db\n    type: postgresql@17\n    mode: NON_HA",
    "mariadb": "# MariaDB database\nservices:\n  - hostname: db\n    type: mariadb@11\n    mode: NON_HA"
  },
  "environment_variables": {
    "postgresql": ["db_connectionString", "db_hostname", "db_port"],
    "usage": "Access from other services using ${servicename_variablename} syntax"
  }
}
```
</details>

<details>
<summary>Runtime Examples (`runtime: "nodejs"`)</summary>

```json
{
  "runtime": "Node.js",
  "examples": {
    "basic": "services:\n  - hostname: app\n    type: nodejs@22\n    enableSubdomainAccess: true\n    minContainers: 1"
  },
  "deployment_yaml": "zerops:\n  - setup: prod\n    build:\n      base: nodejs@22\n      buildCommands: [\"npm i\", \"npm run build\"]\n    run:\n      base: nodejs@22\n      ports: [{\"port\": 3000, \"httpSupport\": true}]\n      start: \"npm run start:prod\"",
  "tips": ["Use nodejs@22 for latest LTS version", "Enable subdomain for public access"]
}
```
</details>

**`load_platform_guide`** - Get workflow guides for different scenarios
- **Required**: `path_type` (fresh_project, existing_service, add_services)

<details>
<summary>Example Output</summary>

```json
{
  "path_type": "fresh_project",
  "steps": [
    "1. discovery() - Check current project state",
    "2. get_service_types() - See available services",
    "3. knowledge_base('nodejs') - Get YAML examples",
    "4. import_services() - Create services"
  ]
}
```
</details>

## Common Workflows

### 1. Initial Project Setup
```bash
# Start with discovery to understand current state
discovery(project_id: "eOc4woejQjC5KhohvkVKPQ")

# Get available service types
get_service_types()

# Get runtime-specific examples
knowledge_base(runtime: "nodejs")

# Create services using YAML
import_services(project_id: "eOc4woejQjC5KhohvkVKPQ", yaml: "...")

# Verify services were created
discovery(project_id: "eOc4woejQjC5KhohvkVKPQ")
```

### 2. Enable Public Access
```bash
# Get service ID from discovery
discovery(project_id: "eOc4woejQjC5KhohvkVKPQ")

# Enable public subdomain
enable_preview_subdomain(service_id: "WAlvwg9GQ3qBQAi37Gts5A")

# Monitor progress
get_running_processes(service_id: "WAlvwg9GQ3qBQAi37Gts5A")

# Check final status and get subdomain URL
discovery(service_id: "WAlvwg9GQ3qBQAi37Gts5A")
```

### 3. Environment Configuration
```bash
# See current environment variables
discovery(project_id: "eOc4woejQjC5KhohvkVKPQ")

# Set shared project-wide variables
set_project_env(project_id: "eOc4woejQjC5KhohvkVKPQ", key: "DATABASE_URL", value: "postgresql://...")

# Set service-specific variables
set_service_env(service_id: "WAlvwg9GQ3qBQAi37Gts5A", key: "PORT", value: "3000")
```

### 4. Monitoring and Debugging
```bash
# Check service logs
get_service_logs(service_id: "WAlvwg9GQ3qBQAi37Gts5A", limit: 50, minimum_severity: "error")

# Monitor active processes
get_running_processes()

# Check specific process status
get_process_status(process_id: "abc123def456")

# View service status and deployment info
discovery(service_id: "WAlvwg9GQ3qBQAi37Gts5A")
```

### 5. Service Management
```bash
# Scale service resources
scale_service(service_id: "WAlvwg9GQ3qBQAi37Gts5A", min_cpu: 1, max_cpu: 4, min_ram: 512, max_ram: 2048)

# Restart service
restart_service(service_id: "WAlvwg9GQ3qBQAi37Gts5A")

# Fix mount issues (development)
remount_service(service_name: "webapp")
```

## Error Handling

### Common Errors:
- **"Project ID is required. Run 'echo $projectId' in the container to get it."**: Agent must run `echo $projectId` to get the project ID
- **"No service found with ID/name 'xyz'"**: Service doesn't exist or wrong ID/name provided
- **"serviceStackTypeNotFound"**: Use `get_service_types` to verify correct type names
- **"Invalid hostname"**: Use alphanumeric characters only, no special characters
- **"Response exceeds maximum tokens"**: Use `limit` parameter to reduce response size
- **"Build logs support requires app version lookup - not yet implemented"**: Use `show_build_logs: false` for runtime logs

### Best Practices:
1. **Always start with `discovery`** to understand current project state
2. **Use filtering** - `discovery(service_id: "...")` for specific services
3. **Verify service types** with `get_service_types` before `import_services`
4. **Get examples** with `knowledge_base(runtime: "...")` for correct YAML structure
5. **Monitor async operations** with `get_running_processes` or `get_process_status`
6. **Check deployment status** - `discovery` shows `active_version` for runtime services
7. **Use appropriate limits** to prevent large responses (`limit` parameter)
8. **Handle mount issues** with `remount_service` for development environments

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
