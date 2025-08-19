# Environment Variable Configuration Guide

This guide demonstrates how to use environment variables with the Zerops MCP server configuration, following Claude Code's `.mcp.json` specification.

## Environment Variable Expansion

The `.mcp.json` file supports two syntaxes for environment variable expansion:
- `${VAR}` - Expands to the value of environment variable VAR
- `${VAR:-default}` - Expands to VAR if set, otherwise uses default value

## Quick Setup Examples

### 1. Basic Setup (Stdio Mode)

Create a `.env` file in your project root:
```bash
# .env
ZEROPS_API_KEY=your-actual-api-key-here
ZEROPS_MCP_PATH=/usr/local/bin/zerops-mcp
```

Load environment variables and run:
```bash
source .env
claude mcp add zerops --config .mcp.json --transport stdio
```

### 2. HTTP Mode with Production Server

```bash
# Set your API key
export ZEROPS_API_KEY="your-api-key"

# Use default production URL (https://mcp-16cb-8080.prg1.zerops.app/)
claude mcp add zerops-cloud --config .mcp.json --transport http
```

### 3. Local Development with Custom Port

```bash
# Set environment variables
export ZEROPS_API_KEY="your-dev-key"
export MCP_PORT=9999
export SKIP_VALIDATION=true  # For local testing only

# Start local server
./zerops-mcp --transport http --port 9999 --skip-validation

# Connect Claude
claude mcp add zerops-local --config .mcp.json --transport http-local
```

## Team Configuration Sharing

### For Development Teams

1. **Share the `.mcp.json` file** in your repository
2. **Create `.env.example`** with required variables:

```bash
# .env.example
ZEROPS_API_KEY=your-api-key-here
ZEROPS_MCP_PATH=/path/to/zerops-mcp
ZEROPS_MCP_URL=https://your-server.com/
LOG_LEVEL=info
```

3. **Each team member** copies and fills their own `.env`:
```bash
cp .env.example .env
# Edit .env with your personal API key
```

### For Different Environments

Use environment-specific variables:

```bash
# Development
export ZEROPS_DEV_API_KEY="dev-key"
export LOG_LEVEL="debug"
export ZEROPS_DEBUG="true"

# Staging
export ZEROPS_STAGING_API_KEY="staging-key"
export ZEROPS_STAGING_URL="https://staging-mcp.zerops.app/"

# Production
export ZEROPS_PROD_API_KEY="prod-key"
export LOG_LEVEL="error"
export MCP_CACHE_ENABLED="true"
```

## Platform-Specific Paths

The `.mcp.json` uses platform-specific default paths:

### macOS
```bash
export ZEROPS_MCP_PATH="${HOME}/.local/bin/zerops-mcp"
export MCP_LOG_DIR="${HOME}/.local/share/zerops-mcp/logs"
export MCP_CACHE_DIR="${HOME}/.cache/zerops-mcp"
```

### Linux
```bash
export ZEROPS_MCP_PATH="${HOME}/.local/bin/zerops-mcp"
export MCP_LOG_DIR="${HOME}/.local/share/zerops-mcp/logs"
export MCP_CACHE_DIR="${HOME}/.cache/zerops-mcp"
```

### Windows
```powershell
$env:ZEROPS_MCP_PATH = "$env:USERPROFILE\.local\bin\zerops-mcp.exe"
$env:MCP_LOG_DIR = "$env:APPDATA\zerops-mcp\logs"
$env:MCP_CACHE_DIR = "$env:LOCALAPPDATA\zerops-mcp\cache"
```

## Security Best Practices

### 1. Never Commit Sensitive Data
```bash
# .gitignore
.env
.env.local
*.key
*_API_KEY
```

### 2. Use Secrets Management
For production deployments, use proper secrets management:

```bash
# AWS Secrets Manager
export ZEROPS_API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id zerops-api-key --query SecretString --output text)

# HashiCorp Vault
export ZEROPS_API_KEY=$(vault kv get -field=api_key secret/zerops)

# Azure Key Vault
export ZEROPS_API_KEY=$(az keyvault secret show \
  --vault-name MyVault --name zerops-api-key --query value -o tsv)
```

### 3. Validate Required Variables
Create a validation script:

```bash
#!/bin/bash
# validate-env.sh

required_vars=(
  "ZEROPS_API_KEY"
  "ZEROPS_MCP_PATH"
)

for var in "${required_vars[@]}"; do
  if [ -z "${!var}" ]; then
    echo "Error: $var is not set"
    exit 1
  fi
done

echo "✅ All required environment variables are set"
```

## Advanced Configuration

### Custom API Endpoints
```bash
# Use custom Zerops API endpoint
export ZEROPS_API_URL="https://api.custom.zerops.io"

# Use custom knowledge base
export KNOWLEDGE_API_URL="https://kb.custom.zerops.io"
```

### Performance Tuning
```bash
# Adjust timeouts and retries
export MCP_TIMEOUT=60000        # 60 seconds
export MCP_MAX_RETRIES=5        # Retry failed requests 5 times

# Configure caching
export MCP_CACHE_ENABLED=true
export MCP_CACHE_TTL=600        # Cache for 10 minutes
```

### Debug Mode
```bash
# Enable debug logging
export ZEROPS_DEBUG=true
export LOG_LEVEL=debug
export MCP_LOG_DIR="./logs"

# Run with verbose output
./zerops-mcp --transport stdio 2>&1 | tee debug.log
```

## Docker Configuration

For containerized deployments:

```dockerfile
# Dockerfile
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY zerops-mcp /usr/local/bin/
ENV MCP_TRANSPORT=http
ENV MCP_HTTP_HOST=0.0.0.0
ENV MCP_HTTP_PORT=8080
EXPOSE 8080
CMD ["zerops-mcp"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  zerops-mcp:
    build: .
    ports:
      - "${MCP_PORT:-8080}:8080"
    environment:
      - MCP_TRANSPORT=http
      - MCP_HTTP_HOST=0.0.0.0
      - MCP_HTTP_PORT=8080
    env_file:
      - .env
```

## Troubleshooting

### Check Expanded Values
```bash
# Print expanded configuration
envsubst < .mcp.json | jq .

# Test specific variable expansion
echo "URL will be: ${ZEROPS_MCP_URL:-https://mcp-16cb-8080.prg1.zerops.app/}"
```

### Common Issues

1. **Variable not expanding**: Ensure variable is exported
   ```bash
   export ZEROPS_API_KEY="your-key"  # ✅ Correct
   ZEROPS_API_KEY="your-key"         # ❌ Not exported
   ```

2. **Default value not working**: Check syntax
   ```json
   "${VAR:-default}"   // ✅ Correct
   "${VAR:default}"    // ❌ Missing dash
   ```

3. **Path issues on Windows**: Use forward slashes or escape backslashes
   ```json
   "${USERPROFILE}/.local/bin/zerops-mcp.exe"     // ✅ Works
   "${USERPROFILE}\\.local\\bin\\zerops-mcp.exe"  // ✅ Also works
   "${USERPROFILE}\.local\bin\zerops-mcp.exe"     // ❌ Invalid
   ```

## Example Scripts

### setup.sh - Complete Setup Script
```bash
#!/bin/bash
# setup.sh - Initialize Zerops MCP with environment variables

# Load environment
if [ -f .env ]; then
  export $(cat .env | xargs)
fi

# Validate required variables
if [ -z "$ZEROPS_API_KEY" ]; then
  echo "Error: ZEROPS_API_KEY not set"
  echo "Please set: export ZEROPS_API_KEY='your-key'"
  exit 1
fi

# Detect transport mode
TRANSPORT="${MCP_TRANSPORT:-stdio}"

# Install binary if needed
if [ ! -f "${ZEROPS_MCP_PATH:-./zerops-mcp}" ]; then
  echo "Installing Zerops MCP..."
  curl -sSL https://raw.githubusercontent.com/krls2020/zerops-mcp-go-sdk/main/install.sh | sh
fi

# Configure Claude Desktop
echo "Configuring Claude Desktop for $TRANSPORT mode..."
claude mcp add zerops --config .mcp.json --transport $TRANSPORT

echo "✅ Setup complete!"
echo "Restart Claude Desktop to use Zerops MCP"
```

## Further Reading

- [Claude Code MCP Documentation](https://docs.anthropic.com/en/docs/claude-code/mcp)
- [Zerops Documentation](https://docs.zerops.io)
- [Model Context Protocol Specification](https://modelcontextprotocol.io)