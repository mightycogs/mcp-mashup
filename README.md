# mcp-mashup

MCP (Model Context Protocol) aggregator that combines multiple MCP servers into a single stdio interface. An MCP client connects to mcp-mashup as if it were a single server, while mcp-mashup spawns and manages multiple backend MCP servers behind the scenes.

This is useful when a client needs access to tools from several MCP servers simultaneously without configuring each one separately. Tool names from backend servers are automatically prefixed with the server name (e.g. `github_create_pr`) and dashes are replaced with underscores to ensure broad client compatibility.

## Build, test, install

Requires Go 1.24+.

```
make build     # build the binary
make test      # run tests
make install   # install to ~/.local/bin
```

Run `make` without arguments to see all available targets.

## Defining mcp-mashup as an MCP server in your client

Configure your MCP client (Cursor, Claude Code, Windsurf, etc.) to use mcp-mashup as a server. The `MCP_CONFIG` environment variable must point to the mcp-mashup configuration file (described in the next section).

```json
{
  "mcpServers": {
    "mashup": {
      "command": "mcp-mashup",
      "env": {
        "MCP_CONFIG": "~/.config/mcp/config.json"
      }
    }
  }
}
```

Optional environment variables:

| Variable | Description | Default |
|---|---|---|
| `MCP_CONFIG` | Path to the configuration file (required) | — |
| `MCP_LOG_LEVEL` | Logging level: `error`, `info`, `debug`, `trace` | `info` |
| `MCP_LOG_FILE` | Path to a log file | stderr |

## Configuring mcp-mashup itself

The configuration file referenced by `MCP_CONFIG` is itself an `mcpServers` definition — the same format your client uses. This is intentional: mcp-mashup acts as a client to these backend servers, so the configuration mirrors what any MCP client expects.

In other words, the structure is nested: your client's `mcpServers` points to mcp-mashup, and mcp-mashup's own config contains another `mcpServers` block pointing to the actual backend servers.

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "your-token"
      }
    },
    "shortcut": {
      "command": "npx",
      "args": ["-y", "@shortcut/mcp"],
      "env": {
        "SHORTCUT_API_TOKEN": "your-token"
      },
      "tools": {
        "allowed": ["search-stories", "get-story", "create-story"]
      }
    }
  }
}
```

The optional `tools.allowed` array restricts which tools are exposed from a given server. If omitted, all tools from that server are available.

## Origin and authorship

This project is a fork of [combine-mcp](https://github.com/nazar256/combine-mcp) by Yurii Nazarenko. The original repository does not specify a license. This fork adapts the software for internal use at [mightycogs](https://github.com/mightycogs) — no additional licensing terms are introduced.
