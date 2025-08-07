#!/usr/bin/env pwsh
# Copyright (c) 2024 Zerops MCP SDK. MIT license.

$ErrorActionPreference = 'Stop'

if ($v) {
  $Version = "v${v}"
}
if ($Args.Length -eq 1) {
  $Version = $Args.Get(0)
}

$McpInstall = $env:ZEROPS_MCP_INSTALL
$BinDir = if ($McpInstall) {
  "${McpInstall}\bin"
} else {
  "${Home}\.zerops\mcp\bin"
}

$McpExe = "$BinDir\zerops-mcp.exe"
$Target = 'win-x64'

$DownloadUrl = if (!$Version) {
  "https://github.com/krls2020/zerops-mcp-go-sdk/releases/latest/download/zerops-mcp-${Target}.exe"
} else {
  "https://github.com/krls2020/zerops-mcp-go-sdk/releases/download/${Version}/zerops-mcp-${Target}.exe"
}

if (!(Test-Path $BinDir)) {
  New-Item $BinDir -ItemType Directory | Out-Null
}

curl.exe -Lo $McpExe $DownloadUrl

$User = [System.EnvironmentVariableTarget]::User
$Path = [System.Environment]::GetEnvironmentVariable('Path', $User)
if (!(";${Path};".ToLower() -like "*;${BinDir};*".ToLower())) {
  [System.Environment]::SetEnvironmentVariable('Path', "${Path};${BinDir}", $User)
  $Env:Path += ";${BinDir}"
}

Write-Output ""
Write-Output "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
Write-Output "🎉 Installation complete! Zerops MCP Server installed to:"
Write-Output "   ${McpExe}"
Write-Output "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
Write-Output ""
Write-Output "📋 Configure Claude Desktop:"
Write-Output ""
Write-Output "1. First, set your Zerops API key as an environment variable:"
Write-Output "   `$env:ZEROPS_API_KEY = `"your-api-key-here`""
Write-Output ""
Write-Output "2. Then add the MCP server to Claude Desktop:"
Write-Output "   claude mcp add zerops -s user `"${McpExe}`""
Write-Output ""
Write-Output "Or manually edit your Claude Desktop config:"
Write-Output "   %APPDATA%\Claude\claude_desktop_config.json"
Write-Output ""
Write-Output "Add this configuration:"
Write-Output "  {"
Write-Output "    `"mcpServers`": {"
Write-Output "      `"zerops`": {"
Write-Output "        `"command`": `"${McpExe}`".Replace('\', '\\')"
Write-Output "        `"env`": {"
Write-Output "          `"ZEROPS_API_KEY`": `"your-api-key-here`""
Write-Output "        }"
Write-Output "      }"
Write-Output "    }"
Write-Output "  }"
Write-Output ""
Write-Output "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
Write-Output "📚 Resources:"
Write-Output "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
Write-Output "• Get API key: https://app.zerops.io/settings/token-management"
Write-Output "• Documentation: https://github.com/krls2020/zerops-mcp-go-sdk"
Write-Output "• Zerops Discord: https://discord.com/invite/WDvCZ54"