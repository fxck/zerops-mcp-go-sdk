#!/bin/sh
# Copyright (c) 2024 Zerops MCP SDK. MIT license.

set -e

case $(uname -sm) in
"Darwin x86_64") target="darwin-amd64" ;;
"Darwin arm64") target="darwin-arm64" ;;
"Linux i386") target="linux-i386" ;;
*) target="linux-amd64" ;;
esac

if [ $# -eq 0 ]; then
  mcp_uri="https://github.com/fxck/zerops-mcp-go-sdk/releases/latest/download/zerops-mcp-${target}"
else
  mcp_uri="https://github.com/fxck/zerops-mcp-go-sdk/releases/download/${1}/zerops-mcp-${target}"
fi

bin_dir="$HOME/.local/bin"
bin_path="$bin_dir/zerops-mcp"
bin_dir_existed=1

if [ ! -d "$bin_dir" ]; then
  mkdir -p "$bin_dir"
  bin_dir_existed=0

  # By default `~/.local/bin` isn't included in PATH if it doesn't exist
  # First try `.bash_profile`. It doesn't exist by default, but if it does, `.profile` is ignored by bash
  if [ "$(uname -s)" = "Linux" ]; then
    if [ -f "$HOME/.bash_profile" ]; then
      . "$HOME/.bash_profile"
    elif [ -f "$HOME/.profile" ]; then
      . "$HOME/.profile"
    fi
  fi
fi

curl --fail --location --progress-bar --output "$bin_path" "$mcp_uri"
chmod +x "$bin_path"

echo
echo "Zerops MCP Server was installed successfully to '$bin_path'"

if command -v zerops-mcp >/dev/null; then
  echo "Run 'zerops-mcp --help' to get started"
  if [ "$bin_dir_existed" = 0 ]; then
    echo "ℹ️ You may need to relaunch your shell."
  fi
else
  if [ "$(uname -s)" = "Darwin" ]; then
    echo 'Add following line to the `/etc/paths` file and relaunch your shell.';
    echo "  $HOME/.local/bin"
    echo
    echo 'You can do so by running:'
    echo "sudo sh -c 'echo \"$HOME/.local/bin\" >> /etc/paths'"
  else
    echo "Manually add the directory to your '$HOME/.profile' (or similar) and relaunch your shell."
    echo '  export PATH="$HOME/.local/bin:$PATH"'
  fi
  echo
  echo "Run '$bin_path --help' to get started"
fi

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 Installation complete! Now configure Claude Desktop:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo
echo "1. First, set your Zerops API key as an environment variable:"
echo "   export ZEROPS_API_KEY=\"your-api-key-here\""
echo
echo "2. Then add the MCP server to Claude Desktop:"
echo "   claude mcp add zerops -s user $bin_path"
echo
echo "Or manually edit your Claude Desktop config:"
if [ "$(uname -s)" = "Darwin" ]; then
  echo "   ~/Library/Application Support/Claude/claude_desktop_config.json"
else
  echo "   ~/.config/Claude/claude_desktop_config.json"
fi
echo
echo "Add this configuration:"
echo '  {
    "mcpServers": {
      "zerops": {
        "command": "'$bin_path'",
        "env": {
          "ZEROPS_API_KEY": "your-api-key-here"
        }
      }
    }
  }'
echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📚 Resources:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "• Get API key: https://app.zerops.io/settings/token-management"
echo "• Documentation: https://github.com/fxck/zerops-mcp-go-sdk"
echo "• Zerops Discord: https://discord.com/invite/WDvCZ54"
