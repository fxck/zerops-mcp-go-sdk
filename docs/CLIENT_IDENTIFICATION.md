# MCP Client Identification

## Overview

MCP clients identify themselves during the `initialize` request through the `clientInfo` parameter. This contains:
- `name`: The client application name
- `version`: The client version
- `title` (optional): Human-readable display name

## Known Client Signatures

### Claude Desktop / Claude Code
```json
{
  "clientInfo": {
    "name": "claude-desktop",
    "version": "0.7.2"
  }
}
```

### ChatGPT (hypothetical - ChatGPT doesn't support MCP yet)
If ChatGPT were to support MCP, it might send:
```json
{
  "clientInfo": {
    "name": "chatgpt-desktop",
    "version": "1.0.0"
  }
}
```

### Other Known Clients
- **RAGFlow**: `"name": "ragflow-mcp-client", "version": "0.1"`
- **Spring AI**: `"name": "spring-ai-mcp", "version": "1.0.0"`
- **Generic MCP clients**: Often use `"name": "mcp-client", "version": "x.x.x"`

## Implementation in Zerops MCP

The HTTP handler logs client information when available:

```go
// In internal/transport/http_handler.go
if method == "initialize" && params != nil {
    if clientInfo, ok := params["clientInfo"].(map[string]interface{}); ok {
        clientName, _ := clientInfo["name"].(string)
        clientVersion, _ := clientInfo["version"].(string)
        fmt.Fprintf(os.Stderr, "Client connected: %s v%s\n", clientName, clientVersion)
        
        // Store in context for potential use
        ctx = context.WithValue(ctx, "clientName", clientName)
        ctx = context.WithValue(ctx, "clientVersion", clientVersion)
    }
}
```

## Usage

### Logging
Client information is logged to stderr when a client connects:
```
Client connected: claude-desktop v0.7.2
```

### Custom Behavior
You can use client info to customize responses:
```go
clientName := ctx.Value("clientName").(string)
if clientName == "claude-desktop" {
    // Claude-specific behavior
} else if clientName == "chatgpt-desktop" {
    // ChatGPT-specific behavior (future)
}
```

## Testing

Run the test script to see client info:
```bash
go run test_client_info.go
```

## Notes

1. **Not all clients send clientInfo** - It's optional in the MCP spec
2. **Version strings vary** - No standard format is enforced
3. **Privacy consideration** - Don't log sensitive client data
4. **Future compatibility** - Design for unknown clients

## Client Detection Strategy

To reliably detect the calling AI model:
1. Check `clientInfo.name` for known patterns
2. Fall back to User-Agent headers (HTTP mode)
3. Consider request patterns/timing as hints
4. Default to generic handling for unknown clients