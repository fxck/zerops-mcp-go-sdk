# Zerops MCP Server

A Model Context Protocol (MCP) server implementation that enables AI assistants to interact with the Zerops platform. This server provides tools for managing projects, services, and deployments through a standardized protocol.

## What is MCP?

The Model Context Protocol is an open standard that enables seamless integration between AI assistants and external systems. This implementation allows AI models to directly interact with Zerops infrastructure through well-defined tools.

## Features

- **Project Management**: Create, list, search, and delete Zerops projects
- **Service Operations**: Manage services including creation, deletion, and configuration
- **Deployment Tools**: Deploy applications using integrated zcli functionality
- **Knowledge Base Access**: Query 159+ deployment recipes and configurations
- **Multi-organization Support**: Work across different Zerops organizations
- **Built-in Instructions**: Provides workflow guidance to connected AI clients

## Installation

### Quick Install

#### macOS/Linux
```bash
curl -sSL https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.sh | sh
```

#### Windows
```powershell
irm https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.ps1 | iex
```

### Build from Source

Requirements:
- Go 1.21 or higher
- Git

```bash
git clone https://github.com/krls2020/zerops-mcp-go-sdk
cd zerops-mcp-go-sdk
make clean && make all  # Build for all platforms
```

### Pre-built Binaries

The latest release includes pre-built binaries for:
- **Windows AMD64**: `zerops-mcp-win-x64.exe`
- **Linux AMD64**: `zerops-mcp-linux-amd64`
- **Linux i386**: `zerops-mcp-linux-i386`
- **macOS Intel**: `zerops-mcp-darwin-amd64`
- **macOS ARM64**: `zerops-mcp-darwin-arm64`

Download from [Releases](https://github.com/krls2020/zerops-mcp-go-sdk/releases)

## Transport Modes

The Zerops MCP server supports two transport modes:

### 1. Stdio Mode (Default)
Traditional stdio-based communication for local installations with Claude Desktop and other MCP clients.

### 2. HTTP Mode (NEW)
Streamable HTTP transport with SSE (Server-Sent Events) for cloud deployments and remote access.

## Configuration

### Prerequisites

- Zerops API key from [Dashboard](https://app.zerops.io/settings/token-management)
- zcli for deployment operations (for stdio mode)

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

The HTTP mode allows the server to be deployed as a web service and accessed remotely.

#### Starting the Server

```bash
# Using command-line flags
ZEROPS_API_KEY="your-api-key" ./zerops-mcp -transport http -host 0.0.0.0 -port 8080

# Using environment variables
export ZEROPS_API_KEY="your-api-key"
export MCP_TRANSPORT="http"
export MCP_HTTP_HOST="0.0.0.0"
export MCP_HTTP_PORT="8080"
./zerops-mcp
```

#### Configuration Options

| Option | Flag | Environment Variable | Default | Description |
|--------|------|---------------------|---------|-------------|
| Transport Mode | `-transport` | `MCP_TRANSPORT` | `stdio` | Transport protocol (`stdio` or `http`) |
| HTTP Host | `-host` | `MCP_HTTP_HOST` | `0.0.0.0` | HTTP server bind address |
| HTTP Port | `-port` | `MCP_HTTP_PORT` | `8080` | HTTP server port |

#### Authentication

HTTP mode uses Bearer token authentication with your ZEROPS_API_KEY:

```bash
Authorization: Bearer your-zerops-api-key
```

#### Endpoints

- `POST /mcp` - Main MCP endpoint for JSON-RPC requests
- `GET /health` - Health check endpoint (no auth required)

#### Claude Desktop Configuration for HTTP Mode

```bash
claude mcp add --transport http zerops https://your-server.com/mcp \
  --header "Authorization: Bearer your-zerops-api-key"
```

Or manually in config:

```json
{
  "mcpServers": {
    "zerops-http": {
      "transport": "http",
      "url": "https://your-server.com/mcp",
      "headers": {
        "Authorization": "Bearer your-zerops-api-key"
      }
    }
  }
}
```

#### Example Deployment on Zerops

The server can be deployed on Zerops platform itself:

```bash
# Deploy to https://mcp-16cb-8080.prg1.zerops.app/
ZEROPS_API_KEY="your-api-key" MCP_TRANSPORT="http" MCP_HTTP_PORT="8080" ./zerops-mcp
```

#### Testing HTTP Mode

```bash
# Health check
curl https://your-server.com/health

# MCP request with authentication
curl -X POST https://your-server.com/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-zerops-api-key" \
  -H "Accept: application/json, text/event-stream" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
```

## Architecture

### Project Structure

```
zerops-mcp-go-sdk/
├── cmd/mcp-server/         # Main entry point
│   └── main.go            # Server initialization (stdio/HTTP)
├── internal/
│   ├── handlers/          # MCP protocol handlers
│   │   ├── register.go    # Tool registration
│   │   └── tools/         # Tool implementations
│   │       ├── auth.go        # Authentication tools
│   │       ├── projects.go    # Project management
│   │       ├── services.go    # Service operations
│   │       ├── deploy.go      # Deployment tools
│   │       └── knowledge.go   # Knowledge base client
│   ├── instructions/      # Workflow instructions
│   │   └── workflow.go    # Built-in guidance
│   └── transport/         # Transport implementations
│       └── http.go        # HTTP/SSE transport
├── tools/                 # Build scripts
│   └── build.sh          # Cross-platform builds
├── install.sh            # Unix installation
├── install.ps1           # Windows installation
└── Makefile              # Build automation
```

### Core Components

**MCP Server Implementation**
- Uses the `github.com/modelcontextprotocol/go-sdk` library
- Implements JSON-RPC 2.0 protocol over stdio and HTTP/SSE
- Provides tool discovery and execution
- Supports both local (stdio) and remote (HTTP) transport modes

**Tool Categories**
- **Authentication**: API key validation
- **Projects**: CRUD operations for Zerops projects
- **Services**: Service lifecycle management
- **Deployment**: Application deployment via zcli
- **Knowledge**: Recipe and configuration queries

## Available Tools

### Project Tools
- `project_list` - List all projects
- `project_create` - Create new project
- `project_delete` - Delete project
- `project_search` - Search projects by name
- `project_import` - Import services from YAML

### Service Tools
- `service_list` - List project services
- `service_info` - Get service details
- `service_delete` - Delete service
- `service_enable_subdomain` - Enable public access

### Deployment Tools
- `deploy_validate` - Validate deployment configuration
- `deploy_push` - Deploy application code

### Knowledge Tools
- `knowledge_search` - Search deployment recipes
- `knowledge_get` - Retrieve specific recipe

### System Tools
- `auth_validate` - Validate API credentials
- `region_list` - List available regions

## Development

### Building

The project uses a Makefile with the following targets:

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

#### Build output:
All binaries are generated in the `bin/` directory with debug symbols included. The build script embeds version information including git branch, tags, and author details.

Example build sizes:
- Windows/Linux AMD64: ~12MB
- macOS Intel: ~12MB
- Linux i386/macOS ARM64: ~11MB

### Testing

The server can be tested directly via stdio:

```bash
export ZEROPS_API_KEY="your-key"
./zerops-mcp
```

Send JSON-RPC requests to stdin and receive responses on stdout.

## Technical Details

### MCP Protocol

The server implements the Model Context Protocol specification:
- Tool discovery via `tools/list`
- Tool execution via `tools/call`
- Resource management
- Progress notifications
- Error handling

### API Integration

- Direct integration with Zerops API
- External knowledge base API for recipes
- zcli wrapper for deployments

### Error Handling

- Structured error responses
- Validation for all inputs
- Graceful fallbacks

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Submit a pull request

## License

MIT License

## Support

- [Zerops Documentation](https://docs.zerops.io)
- [MCP Specification](https://modelcontextprotocol.io)
- [GitHub Issues](https://github.com/krls2020/zerops-mcp-go-sdk/issues)
- [Discord Community](https://discord.com/invite/WDvCZ54)