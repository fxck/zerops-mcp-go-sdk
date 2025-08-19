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
curl -sSL https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.sh | sh

# Windows PowerShell
irm https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.ps1 | iex
```

#### Pre-built Binaries

Download from [GitHub Releases](https://github.com/krls2020/zerops-mcp-go-sdk/releases):

- **Windows**: `zerops-mcp-win-x64.exe`
- **macOS Intel**: `zerops-mcp-darwin-amd64`
- **macOS Apple Silicon**: `zerops-mcp-darwin-arm64`
- **Linux AMD64**: `zerops-mcp-linux-amd64`
- **Linux 386**: `zerops-mcp-linux-i386`

### Add to Claude Code

```bash
# Quick setup
export ZEROPS_API_KEY="your-api-key"
claude mcp add zerops -s user ~/.local/bin/zerops-mcp
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
git clone https://github.com/krls2020/zerops-mcp-go-sdk
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

## Features

- Manage Zerops projects and services
- Deploy applications with 159+ recipes  
- Works with Claude Code AI assistant

## Available Tools

**Projects**: project_list, project_create, project_delete, project_search, project_import  
**Services**: service_list, service_info, service_delete, service_enable_subdomain, service_disable_subdomain, service_start, service_stop  
**Deployment**: deploy_validate, deploy_push  
**Knowledge**: knowledge_search, knowledge_get  
**Auth**: auth_validate, auth_show

## Prerequisites

- **API Key**: Get from [app.zerops.io/settings/token-management](https://app.zerops.io/settings/token-management)
- **zcli**: Required for deployments (stdio mode only)
- **Go 1.21+**: For building from source

## Development

```bash
make test        # Run tests
make lint        # Run linter
make all         # Build all platforms
make help        # Show all targets
```

## Links

[Zerops Docs](https://docs.zerops.io) | [MCP Spec](https://modelcontextprotocol.io) | [Issues](https://github.com/krls2020/zerops-mcp-go-sdk/issues) | [Discord](https://discord.com/invite/WDvCZ54)

## License

MIT