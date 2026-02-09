# MCP Mashup

<img src="docs/small_create_with_ai.png" style="float: left; margin: 0 15px 15px 0;" width="150">


MCP (Model Context Protocol) aggregator that combines multiple MCP servers into a single stdio interface. An MCP client connects to mcp-mashup as if it were a single server, while mcp-mashup spawns and manages multiple backend MCP servers behind the scenes. 

_Only the stdio transport is supported. The goal of mcp-mashup is purely logical aggregation — it is not intended to extend available protocols, secure access, or abstract away the complexity of the underlying MCP servers. For those capabilities, consider other solutions._

## Build, test, install

Requires Go 1.24+.

```
make build     # build the binary
make test      # run tests
make install   # install to ~/.local/bin
```

Run `make` without arguments to see all available targets.

## Configuration

### Defining mcp-mashup in your client

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
| `MCP_LOG_FILE` | Path to a log file. Without it, logs go to stderr | stderr |

### Configuring mcp-mashup itself

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

The optional `tools.prefixed` array controls which tools keep the server-name prefix. If omitted, all tools are prefixed (default behavior). If set to an empty array `[]`, no tools are prefixed — they keep their original names (with dashes replaced by underscores). If specific tool names are listed, only those tools are prefixed while the rest are exposed with their original names.

_**NOTE:** `tools.allowed` and `tools.prefixed` are custom constructs specific to mcp-mashup and are not part of the MCP standard_

Tool names from backend servers are automatically prefixed with the server name and dashes are replaced with underscores. For example, a tool `create-pr` from server `github` becomes `github_create_pr` in the client.

## Origin and authorship

This project is a fork of [combine-mcp](https://github.com/nazar256/combine-mcp) by Yurii Nazarenko. The original repository does not specify a license. This fork adapts the software for internal use at [mightycogs](https://github.com/mightycogs) — no additional licensing terms are introduced.
