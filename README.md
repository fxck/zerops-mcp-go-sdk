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

## Claude Desktop Setup

```bash
export ZEROPS_API_KEY="your-api-key"
claude mcp add zerops -s user ~/.local/bin/zerops-mcp
```

## HTTP Server Setup

```bash
./zerops-mcp --transport http --port 8080
# Clients authenticate via: Authorization: Bearer <api-key>
```


## Tools

**Authentication**: auth_validate, auth_show  
**Projects**: project_list, project_create, project_delete, project_search, project_import  
**Services**: service_list, service_info, service_delete, service_enable_subdomain, service_disable_subdomain, service_start, service_stop  
**Deployment**: deploy_validate, deploy_push  
**Knowledge**: knowledge_search, knowledge_get

## Recipe Import Requirements

Recipe services require standard hostnames:
- Adminer → `adminer`
- Mailpit → `mailpit`
- S3Browser → `s3browser`

Custom names: import first, rename in GUI later.

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