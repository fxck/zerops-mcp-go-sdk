# Zerops MCP Tools Documentation

This document provides comprehensive documentation for all available Zerops MCP (Model Context Protocol) tools.

## Overview

The Zerops MCP SDK provides 14 tools for managing Zerops projects, services, and deployments through AI assistants like Claude.

## Quick Reference

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

## Detailed Tool Documentation

### 1. discovery

**Purpose:** ESSENTIAL FIRST STEP: Discovers all services in a project with their IDs, hostnames, service types, and environment variable availability. CRITICAL: Requires a project ID. To get the project ID, the agent can run 'echo $projectId' in the container environment.

**Parameters:**
- `project_id` (required): Project UUID. Get it by running 'echo $projectId' in the container.

**Returns (Condensed):**
```json
{
  "project": {
    "id": "project-uuid",
    "name": "my-project",
    "env_keys": ["DATABASE_URL", "API_KEY"]
  },
  "services": [
    {
      "id": "service-uuid",
      "hostname": "app",
      "type": "nodejs@22",
      "env_keys": ["PORT", "NODE_ENV"],
      "process_count": 2
    }
  ],
  "count": 1
}
```

**Usage:**
```
Always run this first to understand the project structure and get service IDs for other operations.
```

---

### 2. get_service_types

**Purpose:** Returns all available Zerops service types and versions.

**Parameters:** None

**Returns:**
```json
{
  "service_types": [
    "nodejs@22", "nodejs@20", "nodejs@18",
    "python@3.12", "python@3.11", 
    "go@1.22", "go@1.21",
    "php@8.3", "php@8.2",
    "postgresql@16", "postgresql@15",
    "mariadb@11", "mariadb@10.6",
    "mongodb@7", "mongodb@6",
    "valkey@7.2", "redis@7",
    "nginx@1.24", "static@1"
  ],
  "count": 19,
  "note": "Use knowledge_base tool for detailed configuration examples"
}
```

**Usage:**
```
Use before import_services to verify correct service type names.
```

---

### 3. import_services

**Purpose:** Creates services in a project using YAML configuration.

**Parameters:**
- `project_id` (optional): Project UUID. If not provided, uses `$projectId` environment variable.
- `yaml` (required): YAML configuration for services

**YAML Format:**
```yaml
services:
  - hostname: app
    type: nodejs@22
    enableSubdomainAccess: true
    minContainers: 1
  - hostname: db
    type: postgresql@16
    mode: NON_HA
```

**Returns:**
```json
{
  "status": "import_initiated",
  "message": "Services are being created. Check status with 'discovery' tool."
}
```

**Usage:**
```
1. Run get_service_types to verify service type names
2. Use knowledge_base for runtime-specific examples
3. Validate hostnames are alphanumeric (no special characters)
```

---

### 4. enable_preview_subdomain

**Purpose:** Enables public subdomain access for a web service (async operation).

**Parameters:**
- `service_id` (required): Service UUID from discovery

**Returns:**
```json
{
  "process_id": "process-uuid",
  "status": "process_started",
  "message": "Subdomain enablement started. Use 'get_running_processes' with this service_id to check progress. Once completed, use 'discovery' to see the actual subdomain URL."
}
```

**Usage:**
```
1. Only works for web services (nodejs, php, python, go, etc.)
2. Returns process_id for async operation
3. Monitor with get_running_processes
4. Check discovery for final subdomain URL
```

---

### 5. scale_service

**Purpose:** Configures service scaling parameters (CPU, RAM, containers).

**Parameters:**
- `service_id` (required): Service UUID from discovery
- `min_cpu` (optional): Minimum CPU cores (0.25-20)
- `max_cpu` (optional): Maximum CPU cores (0.25-20)
- `min_ram` (optional): Minimum RAM in GB (0.5-32)
- `max_ram` (optional): Maximum RAM in GB (0.5-32)
- `min_containers` (optional): Minimum containers (1-6)
- `max_containers` (optional): Maximum containers (1-6)

**Returns:**
```json
{
  "status": "scaling_configured",
  "service_id": "service-uuid",
  "parameters": {
    "min_cpu": 1,
    "max_cpu": 2,
    "min_ram": 1,
    "max_ram": 2
  },
  "message": "Service scaling parameters have been configured"
}
```

**Usage:**
```
- Set min/max values for auto-scaling
- Set equal values for fixed allocation
- Leave parameters empty to keep current settings
```

---

### 6. get_service_logs

**Purpose:** Retrieves logs from a specific service with filtering options.

**Parameters:**
- `service_id` (required): Service UUID from discovery
- `lines` (optional): Number of log lines (1-1000, default: 100)
- `since` (optional): Time period ("1h", "30m", "24h", "7d")

**Returns:**
```json
{
  "service_id": "service-uuid",
  "service_name": "app",
  "logs": [
    {
      "timestamp": "2024-01-01T12:00:00Z",
      "level": "info",
      "message": "Application started"
    }
  ],
  "lines": 100,
  "note": "Log retrieval requires proper API endpoint implementation"
}
```

**Usage:**
```
- Start with small line counts for large logs
- Use since parameter for time-based filtering
- Good for debugging and monitoring
```

---

### 7. set_project_env

**Purpose:** Sets project-level environment variables (available to all services).

**Parameters:**
- `project_id` (optional): Project UUID. If not provided, uses `$projectId` environment variable.
- `key` (required): Environment variable name (UPPERCASE recommended)
- `value` (required): Environment variable value

**Returns:**
```json
{
  "process_id": "process-uuid",
  "status": "env_var_set",
  "key": "DATABASE_URL",
  "message": "Project environment variable 'DATABASE_URL' has been set"
}
```

**Usage:**
```
- Use for shared configuration (database URLs, API keys)
- Available to ALL services in the project
- Override service-level variables with same name
- Use UPPERCASE naming convention
```

---

### 8. set_service_env

**Purpose:** Sets service-specific environment variables.

**Parameters:**
- `service_id` (required): Service UUID from discovery
- `key` (required): Environment variable name (UPPERCASE recommended)
- `value` (required): Environment variable value

**Returns:**
```json
{
  "status": "env_var_configured",
  "service_id": "service-uuid",
  "key": "PORT",
  "message": "Service environment variable 'PORT' has been configured",
  "note": "Service environment variables are managed as UserData in Zerops"
}
```

**Usage:**
```
- Use for service-specific configuration
- Overrides project-level variables with same name
- Good for ports, service-specific settings
```

---

### 9. get_running_processes

**Purpose:** Monitors running processes, optionally filtered by service.

**Parameters:**
- `service_id` (optional): Service UUID to filter processes
- `limit` (optional): Maximum processes to return (1-100, default: 20)

**Returns (Condensed):**
```json
{
  "service": "app",
  "processes": [
    {
      "id": "process-uuid",
      "status": "running",
      "created": "12:00:05"
    }
  ],
  "count": 1
}
```

**Usage:**
```
- Monitor deployment progress
- Check async operation status (like enable_preview_subdomain)
- Use service_id filter for specific service
- Use limit to control response size
```

---

### 10. restart_service

**Purpose:** Restarts a service (useful after environment variable changes or configuration updates).

**Parameters:**
- `service_id` (required): Service UUID from discovery

**Returns:**
```json
{
  "process_id": "restart-abc123-1234567890",
  "status": "process_started",
  "service_id": "service-uuid",
  "service_name": "app",
  "message": "Service restart initiated. Use 'get_process_status' to monitor progress."
}
```

**Usage:**
```
- After setting environment variables
- After configuration changes
- When service is not responding properly
- Monitor progress with get_process_status
```

---

### 11. remount_service

**Purpose:** Reconnects SSHFS mounts for a service (fixes file system connection issues).

**Parameters:**
- `service_name` (required): Service hostname (not ID) for SSHFS remount

**Returns:**
```json
{
  "status": "success",
  "service_name": "app",
  "command": "sshfs -o StrictHostKeyChecking=no,reconnect,ServerAliveInterval=15,ServerAliveCountMax=3,auto_cache,kernel_cache \"app:/var/www\" \"/var/www/app\"",
  "message": "Run this command to remount SSHFS for service 'app':",
  "instructions": "Copy and run the command above in your terminal to reconnect the SSHFS mount."
}
```

**Usage:**
```
- When file system access is broken
- After network connectivity issues
- When getting file permission errors
- Copy and run the returned SSHFS command
```

---

### 12. get_process_status

**Purpose:** Gets the status of a specific process by its ID (for monitoring async operations).

**Parameters:**
- `process_id` (required): Process UUID returned from async operations

**Returns:**
```json
{
  "process_id": "process-uuid",
  "status": "running",
  "created": "2024-01-01 12:00:00"
}
```

**Usage:**
```
- Monitor restart_service operations
- Check enable_preview_subdomain progress
- Track import_services completion
- Debug failed async operations
```

---

### 13. load_platform_guide

**Purpose:** Loads path-specific guides for different project scenarios (fetched from GitHub with 10min cache).

**Parameters:**
- `path_type` (required): Guide type to load
  - `"fresh_project"`: Complete guide for starting new projects
  - `"existing_service"`: Guide for working with existing services  
  - `"add_services"`: Guide for adding services to existing projects

**Returns:**
```json
{
  "source": "fallback",
  "note": "Should fetch from https://raw.githubusercontent.com/zeropsio/zagent-knowledge/main/guides/fresh_project.json (10min cache)",
  "content": {
    "path_type": "fresh_project",
    "title": "Complete Guide: Starting a Fresh Project",
    "workflow": [
      "1. discovery() - Check current project state (should be empty)",
      "2. get_service_types() - See available service types",
      "3. knowledge_base('nodejs') - Get complete YAML examples",
      "4. import_services(yaml: '...') - Create your services",
      "5. discovery() - Verify services were created",
      "6. enable_preview_subdomain(service_id: '...') - Enable public access",
      "7. get_process_status(process_id: '...') - Monitor subdomain setup",
      "8. discovery() - Get final subdomain URLs",
      "9. set_project_env() / set_service_env() - Configure environment",
      "10. get_service_logs() - Monitor application startup"
    ]
  },
  "todo": "Implement HTTP fetch with caching"
}
```

**Usage:**
```
- Get step-by-step workflow guidance
- Learn best practices for specific scenarios
- Follow structured project setup processes
- Content is fetched from GitHub with 10-minute caching
```

---

### 14. knowledge_base

**Purpose:** Provides comprehensive YAML examples and deployment patterns for specific runtimes. For unsupported runtimes, returns Node.js pattern as reference with guidance to adapt.

**Parameters:**
- `runtime` (required): Runtime type to get examples for

**Supported Runtimes:**
- Web: `nodejs`, `python`, `go`, `php`
- Databases: `postgresql`, `mariadb`, `mongodb`
- Cache: `redis`, `valkey`

**Returns (for nodejs):**
```json
{
  "runtime": "Node.js",
  "examples": {
    "basic": "services:\n  - hostname: app\n    type: nodejs@22\n    enableSubdomainAccess: true\n    minContainers: 1",
    "with_database": "services:\n  - hostname: app\n    type: nodejs@22\n    enableSubdomainAccess: true\n    minContainers: 1\n  - hostname: db\n    type: postgresql@16\n    mode: NON_HA"
  },
  "deployment_yaml": "zerops:\n  - setup: prod\n    build:\n      base: nodejs@22\n      prepareCommands:\n        - npm install -g typescript\n      buildCommands:\n        - npm i\n        - npm run build\n      deployFiles:\n        - ./dist\n        - ./node_modules\n        - ./package.json\n    run:\n      base: nodejs@22\n      ports:\n        - port: 3000\n          httpSupport: true\n      envVariables:\n        NODE_ENV: production\n        DB_NAME: db\n        DB_HOST: db\n        DB_USER: db\n        DB_PASS: ${db_password}\n      start: npm run start:prod\n      healthCheck:\n        httpGet:\n          port: 3000\n          path: /status\n\n  - setup: dev\n    build:\n      base: nodejs@22\n      buildCommands:\n        - npm i\n      deployFiles: ./\n    run:\n      base: nodejs@22\n      envVariables:\n        DB_NAME: db\n        DB_HOST: db\n        DB_USER: db\n        DB_PASS: ${db_password}\n      ports:\n        - port: 3000\n          httpSupport: true\n      start: zsc noop",
  "tips": [
    "Use nodejs@22 for latest LTS version",
    "Enable subdomain for public access",
    "Set minContainers: 2 for high availability",
    "Use environment variables for configuration",
    "Use 'prod' setup for production-like services, 'dev' for development/remote",
    "Reference database password with ${db_password}"
  ]
}
```

**Returns (for unsupported runtime like "rust"):**
```json
{
  "runtime": "rust",
  "message": "Runtime 'rust' not directly supported. Use Node.js pattern as reference.",
  "pattern": { ... nodejs examples ... }
}
```

**Usage:**
```
- Get complete zerops.yml examples for supported runtimes
- Unsupported runtimes get Node.js pattern with guidance to adapt
- Use before import_services for correct YAML format
```

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

## Notes

- All UUIDs follow the pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
- Many operations are asynchronous and return process IDs for monitoring
- Use `discovery` to get final state after async operations complete
- Response sizes are limited to prevent token overflow (use pagination/filtering)