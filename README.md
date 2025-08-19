# Zerops MCP Server

MCP server for managing Zerops infrastructure through AI assistants like Claude.

## Features

- Manage Zerops projects and services
- Deploy applications with 159+ recipes
- Local (stdio) or remote (HTTP) modes

## Installation

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/krls2020/zerops-mcp-go-sdk/releases)

### Build from Source

```bash
git clone https://github.com/krls2020/zerops-mcp-go-sdk
cd zerops-mcp-go-sdk
make all
```

### Quick Install

```bash
# macOS/Linux
curl -sSL https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.sh | sh

# Windows PowerShell
irm https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.ps1 | iex
```


## Prerequisites

- Zerops API key: [app.zerops.io/settings/token-management](https://app.zerops.io/settings/token-management)
- zcli (for deployments in stdio mode)

## Claude Code Setup

```bash
# Quick setup
export ZEROPS_API_KEY="your-api-key"
claude mcp add zerops -s user ~/.local/bin/zerops-mcp
```

Or add to config manually (`~/Library/Application Support/Claude/claude_desktop_config.json`):

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

## HTTP Server Setup

```bash
# Start server
./zerops-mcp --transport http --port 8080
```

### Connect with Claude Code

```bash
# Using hosted server
claude mcp add --transport http zerops https://your-server.zerops.app/mcp \
  --header "Authorization: Bearer your-api-key"

# Or using environment variable
export ZEROPS_API_KEY="your-api-key"
claude mcp add --transport http zerops https://your-server.zerops.app/mcp \
  --header "Authorization: Bearer $ZEROPS_API_KEY"
```

### Testing the Server

```bash
curl -X POST http://localhost:8080/ \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```


## Tools

**Authentication**: auth_validate, auth_show  
**Projects**: project_list, project_create, project_delete, project_search, project_import  
**Services**: service_list, service_info, service_delete, service_enable_subdomain, service_disable_subdomain, service_start, service_stop  
**Deployment**: deploy_validate, deploy_push  
**Knowledge**: knowledge_search, knowledge_get

## Development

```bash
make test        # Run tests
make lint        # Run linter
make all         # Build all platforms
```




## License

MIT License

## Links

[Zerops Docs](https://docs.zerops.io) | [MCP Spec](https://modelcontextprotocol.io) | [Issues](https://github.com/krls2020/zerops-mcp-go-sdk/issues) | [Discord](https://discord.com/invite/WDvCZ54)