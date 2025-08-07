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

## Configuration

### Prerequisites

- Zerops API key from [Dashboard](https://app.zerops.io/settings/token-management)
- zcli for deployment operations

### Claude Desktop Setup

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

## Architecture

### Project Structure

```
zerops-mcp-go-sdk/
├── cmd/mcp-server/         # Main entry point
│   └── main.go            # Server initialization
├── internal/
│   ├── handlers/          # MCP protocol handlers
│   │   ├── register.go    # Tool registration
│   │   └── tools/         # Tool implementations
│   │       ├── auth.go        # Authentication tools
│   │       ├── projects.go    # Project management
│   │       ├── services.go    # Service operations
│   │       ├── deploy.go      # Deployment tools
│   │       └── knowledge.go   # Knowledge base client
│   └── instructions/      # Workflow instructions
│       └── workflow.go    # Built-in guidance
├── tools/                 # Build scripts
│   └── build.sh          # Cross-platform builds
├── install.sh            # Unix installation
├── install.ps1           # Windows installation
└── Makefile              # Build automation
```

### Core Components

**MCP Server Implementation**
- Uses the `github.com/mark3labs/mcp-go` library
- Implements JSON-RPC 2.0 protocol over stdio
- Provides tool discovery and execution

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