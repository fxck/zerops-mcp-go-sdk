# Zerops MCP Server

A Model Context Protocol (MCP) server implementation that enables AI assistants to interact with the Zerops platform. This server provides tools for managing projects, services, and deployments through a standardized protocol with support for both local (stdio) and remote (HTTP) transport modes.

## What is MCP?

The Model Context Protocol is an open standard that enables seamless integration between AI assistants and external systems. This implementation allows AI models to directly interact with Zerops infrastructure through well-defined tools.

## Features

- **Dual Transport Support**: Both stdio (local) and HTTP (remote) modes
- **Shared Tool Logic**: DRY architecture with unified tool implementations
- **Project Management**: Create, list, search, and delete Zerops projects
- **Service Operations**: Manage services including creation, deletion, start/stop, and subdomain configuration
- **Deployment Tools**: Deploy applications using integrated zcli functionality (stdio) or instructions (HTTP)
- **Knowledge Base Access**: Query 159+ deployment recipes and configurations
- **Multi-organization Support**: Work across different Zerops organizations
- **Per-Request Authentication**: HTTP mode supports stateless authentication with client-provided API keys
- **Built-in Instructions**: Provides workflow guidance to connected AI clients

## Installation

### Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/krls2020/zerops-mcp-go-sdk/releases):

- **Windows**: `zerops-mcp-win-x64.exe`
- **Linux AMD64**: `zerops-mcp-linux-amd64`
- **Linux 386**: `zerops-mcp-linux-i386`
- **macOS Intel**: `zerops-mcp-darwin-amd64`
- **macOS Apple Silicon**: `zerops-mcp-darwin-arm64`

### Build from Source

Requirements:
- Go 1.21 or higher
- Git

```bash
git clone https://github.com/krls2020/zerops-mcp-go-sdk
cd zerops-mcp-go-sdk
make clean && make all  # Build for all platforms
```

Or build for specific platform:
```bash
go build -o zerops-mcp cmd/mcp-server/main.go
```

### Quick Install Script

#### macOS/Linux
```bash
curl -sSL https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.sh | sh
```

#### Windows
```powershell
irm https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.ps1 | iex
```

## Transport Modes

The Zerops MCP server supports two transport modes with shared tool logic:

### 1. Stdio Mode (Default)
Traditional stdio-based communication for local installations with Claude Desktop and other MCP clients. Requires ZEROPS_API_KEY environment variable at startup.

### 2. HTTP Mode
Stateless HTTP transport for cloud deployments and remote access. Each client provides their own API key via Authorization header.

## Configuration

### Prerequisites

- Zerops API key from [Dashboard](https://app.zerops.io/settings/token-management)
- zcli for deployment operations (stdio mode only)

### Stdio Mode Setup (Claude Desktop)

#### Recommended Method

After installation, configure Claude Desktop using the CLI:

1. Set your API key as environment variable:
```bash
export ZEROPS_API_KEY="your-api-key-here"
```

2. Add the MCP server to Claude Desktop:
```bash
claude mcp add zerops -s user ~/.local/bin/zerops-mcp
```

#### Manual Configuration

Alternatively, manually edit your Claude Desktop config file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Linux**: `~/.config/Claude/claude_desktop_config.json`  
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "zerops": {
      "command": "/path/to/zerops-mcp",
      "env": {
        "ZEROPS_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

### HTTP Mode Setup

The HTTP mode allows the server to be deployed as a web service and accessed remotely. Unlike stdio mode, HTTP mode is **stateless** - each client provides their own API key.

#### Starting the Server

```bash
# Start HTTP server - clients MUST provide API key via Authorization header
./zerops-mcp --transport http --host 0.0.0.0 --port 8080

# Using environment variables
export MCP_TRANSPORT="http"
export MCP_HTTP_HOST="0.0.0.0"
export MCP_HTTP_PORT="8080"
./zerops-mcp
```

#### Configuration Options

| Option | Flag | Environment Variable | Default | Description |
|--------|------|---------------------|---------|-------------|
| Transport Mode | `--transport` | `MCP_TRANSPORT` | `stdio` | Transport protocol (`stdio` or `http`) |
| HTTP Host | `--host` | `MCP_HTTP_HOST` | `0.0.0.0` | HTTP server bind address |
| HTTP Port | `--port` | `MCP_HTTP_PORT` | `8080` | HTTP server port |

#### Authentication

HTTP mode uses **per-request authentication**. Each client MUST provide their own Zerops API key via Bearer token:

```bash
Authorization: Bearer your-zerops-api-key
```

The server creates a new Zerops SDK client for each request using the provided API key.

**Important**: 
- ✅ **Production**: The server always enforces authentication via Bearer token
- ⚠️ **Security**: Clients MUST provide their API key via Authorization header

#### Endpoints

- `POST /` - Main MCP endpoint for JSON-RPC requests
- `GET /health` - Health check endpoint (no auth required)

#### Claude Desktop Configuration for HTTP Mode

```bash
# Configure Claude to use HTTP transport
claude mcp add --transport http zerops https://your-server.com/ \
  --header "Authorization: Bearer your-zerops-api-key"
```

Or manually in config:

```json
{
  "mcpServers": {
    "zerops-http": {
      "transport": "http",
      "url": "https://your-server.com/",
      "headers": {
        "Authorization": "Bearer your-zerops-api-key"
      }
    }
  }
}
```

#### Example Cloud Deployment

For production deployment on Zerops platform:

```bash
# Production deployment (requires authentication)
./zerops-mcp --transport http --host 0.0.0.0 --port 8080

# Example endpoint: https://mcp-16cb-8080.prg1.zerops.app/
# Note: Clients must provide Bearer token with ZEROPS_API_KEY
```

#### Testing HTTP Mode

```bash
# Health check
curl https://mcp-16cb-8080.prg1.zerops.app/health

# List available tools
curl -X POST https://mcp-16cb-8080.prg1.zerops.app/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-zerops-api-key" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# Call a tool
curl -X POST https://mcp-16cb-8080.prg1.zerops.app/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-zerops-api-key" \
  -d '{
    "jsonrpc":"2.0",
    "method":"tools/call",
    "params": {
      "name": "project_list",
      "arguments": {}
    },
    "id":2
  }'
```

## Architecture

### Shared Tool Registry (DRY Design)

The server uses a **shared tool registry** that ensures both stdio and HTTP transports use the same tool implementations:

```
internal/
├── handlers/
│   ├── shared/
│   │   └── registry.go         # Central tool registry
│   ├── tools/
│   │   ├── *_shared.go         # Shared tool implementations
│   │   └── *.go                # MCP wrapper for stdio
│   └── register_shared.go      # Registration logic
└── transport/
    ├── http_shared.go           # HTTP handler using registry
    └── http.go                  # HTTP server setup
```

### Key Components

**Shared Tool Registry** (`internal/handlers/shared/registry.go`)
- Central registry for all tool definitions
- Single source of truth for tool logic
- Used by both stdio and HTTP transports

**Tool Implementations** (`internal/handlers/tools/*_shared.go`)
- `auth_shared.go` - Authentication tools
- `projects_shared.go` - Project management
- `services_shared.go` - Service operations
- `deploy_shared.go` - Deployment tools
- `knowledge_shared.go` - Knowledge base access

**Transport Handlers**
- **Stdio**: Uses MCP SDK directly with registered tools
- **HTTP**: Uses shared registry for stateless operation

### Transport-Specific Behavior

Only deployment tools behave differently between transports:
- **Stdio mode**: Executes `zcli` commands locally
- **HTTP mode**: Returns instructions for manual execution

All other tools use identical logic regardless of transport.

## Available Tools (18 Total)

### Authentication Tools (2)
- `auth_validate` - Validate API credentials and show account info
- `auth_show` - Show authentication status and available regions

### Project Tools (5)
- `project_list` - List all projects across organizations
- `project_create` - Create new project
- `project_delete` - Delete project (requires confirmation)
- `project_search` - Search projects by name
- `project_import` - Import services from YAML configuration

### Service Tools (7)
- `service_list` - List services in a project
- `service_info` - Get detailed service information
- `service_delete` - Delete service (requires confirmation)
- `service_enable_subdomain` - Enable public subdomain access
- `service_disable_subdomain` - Disable public subdomain access
- `service_start` - Start a stopped service
- `service_stop` - Stop a running service

### Deployment Tools (2)
- `deploy_validate` - Validate deployment configuration
- `deploy_push` - Deploy application code

### Knowledge Tools (2)
- `knowledge_search` - Search deployment recipes and patterns
- `knowledge_get` - Retrieve specific recipe by ID

## Development

### Building

```bash
make help        # Show available targets
make test        # Run tests
make lint        # Run linter  
make clean       # Remove build artifacts
make all         # Build for all platforms
```

#### Platform-specific builds:
```bash
make windows-amd  # Windows AMD64
make linux-amd    # Linux AMD64
make linux-i386   # Linux i386
make darwin-amd   # macOS Intel
make darwin-arm   # macOS Apple Silicon
```

### Testing

#### Stdio Mode Testing
```bash
export ZEROPS_API_KEY="your-key"
./zerops-mcp
# Send JSON-RPC to stdin
```

#### HTTP Mode Testing
```bash
# Start server
./zerops-mcp --transport http --port 8080

# Test with curl
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

## Technical Details

### MCP Protocol Implementation

The server implements the Model Context Protocol specification:
- Tool discovery via `tools/list`
- Tool execution via `tools/call`
- Shared tool registry for consistent behavior
- Error handling with structured responses

### API Integration

- **Zerops SDK**: Direct integration with Zerops API
- **Knowledge Base**: External API for deployment recipes
- **zcli Wrapper**: Local command execution (stdio mode only)
- **Per-Request Clients**: Stateless operation in HTTP mode

### Security

- **Stdio Mode**: API key stored in environment variable
- **HTTP Mode**: Per-request authentication with Bearer tokens
- **No Shared State**: Each HTTP request is independent
- **Always Validated**: Authentication is always enforced in production

## Publishing Checklist

### Before Release
- [x] Remove all debug/test code
- [x] Remove `--skip-validation` flag
- [x] Clean up demo configurations
- [x] Format all Go code
- [x] Update dependencies
- [x] Test both transport modes
- [x] Update documentation
- [x] Create CHANGELOG.md

### Release Process
1. Tag the release: `git tag v1.0.0`
2. Run release script: `./release.sh 1.0.0`
3. Upload binaries to GitHub Releases
4. Update installation scripts
5. Publish to MCP registry

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Ensure both transports work correctly
5. Submit a pull request

## License

MIT License

## Support

- [Zerops Documentation](https://docs.zerops.io)
- [MCP Specification](https://modelcontextprotocol.io)
- [GitHub Issues](https://github.com/krls2020/zerops-mcp-go-sdk/issues)
- [Discord Community](https://discord.com/invite/WDvCZ54)