# MCP Setup

Guild ships `guild-agentdesk-mcp`, an executable MCP stdio server that wraps the current repo's `agentdesk.yaml` and `.agentdesk/` directory.

Use it when you want Codex, Claude, OpenClaw, OpenFang, or another MCP-capable agent host to fetch mandates, claim work, compile context, run preflight checks, request approvals, publish proof, verify completion, and export replay without a hosted service.

## 90-Second Agent Quickstart

From this repo:

```bash
corepack enable
corepack pnpm install --frozen-lockfile
go build -o bin/guild ./cli/cmd/guild
./bin/guild agentdesk init
./bin/guild agentdesk mandate create "Update MCP docs" --writable "docs/**,adapters/mcp/**"
GUILD_CLI="$(pwd)/bin/guild" corepack pnpm --dir adapters/mcp exec guild-agentdesk-mcp
```

The MCP host connects to that last command over stdio.

After connection, the agent flow is:

```text
guild_get_next_mandate
guild_claim_mandate
guild_compile_context
guild_check_preflight
guild_publish_artifact
guild_create_handoff
guild_verify_mandate
guild_export_replay_bundle
```

## Local Binary Path

For local development, build the CLI once and point the MCP server at the absolute binary path:

```bash
go build -o bin/guild ./cli/cmd/guild
export GUILD_CLI="$(pwd)/bin/guild"
corepack pnpm --dir adapters/mcp exec guild-agentdesk-mcp
```

After the repo is available on GitHub, users can also install the CLI with Go:

```bash
go install github.com/lucid-fdn/guild/cli/cmd/guild@latest
export GUILD_CLI="guild"
corepack pnpm --dir adapters/mcp exec guild-agentdesk-mcp
```

Until the MCP package is published to npm, the supported package-style path is the workspace executable:

```bash
corepack pnpm --dir adapters/mcp exec guild-agentdesk-mcp
```

## Claude Desktop Config

Use the local binary path you built above:

```json
{
  "mcpServers": {
    "guild-agentdesk": {
      "command": "corepack",
      "args": ["pnpm", "--dir", "/absolute/path/to/guild/adapters/mcp", "exec", "guild-agentdesk-mcp"],
      "env": {
        "GUILD_CLI": "/absolute/path/to/guild/bin/guild"
      }
    }
  }
}
```

If the agent should pull GitHub Issues labeled `agent:ready`, add:

```json
{
  "mcpServers": {
    "guild-agentdesk": {
      "command": "corepack",
      "args": ["pnpm", "--dir", "/absolute/path/to/guild/adapters/mcp", "exec", "guild-agentdesk-mcp"],
      "env": {
        "GUILD_CLI": "/absolute/path/to/guild/bin/guild",
        "GUILD_AGENTDESK_SOURCE": "github",
        "GITHUB_REPOSITORY": "lucid-fdn/guild",
        "GITHUB_TOKEN": "ghp_or_fine_grained_token"
      }
    }
  }
}
```

## Codex Config

For Codex-style MCP configuration, use a stdio server entry:

```toml
[mcp_servers.guild-agentdesk]
command = "corepack"
args = ["pnpm", "--dir", "/absolute/path/to/guild/adapters/mcp", "exec", "guild-agentdesk-mcp"]

[mcp_servers.guild-agentdesk.env]
GUILD_CLI = "/absolute/path/to/guild/bin/guild"
GUILD_AGENTDESK_SOURCE = "github"
GITHUB_REPOSITORY = "lucid-fdn/guild"
GITHUB_TOKEN = "ghp_or_fine_grained_token"
```

## OpenFang / OpenClaw / Generic MCP Config

Any host that accepts MCP stdio server definitions can use the same command shape:

```json
{
  "name": "guild-agentdesk",
  "transport": "stdio",
  "command": "corepack",
  "args": ["pnpm", "--dir", "/absolute/path/to/guild/adapters/mcp", "exec", "guild-agentdesk-mcp"],
  "env": {
    "GUILD_CLI": "/absolute/path/to/guild/bin/guild",
    "GUILD_AGENTDESK_SOURCE": "github",
    "GITHUB_REPOSITORY": "lucid-fdn/guild",
    "GITHUB_TOKEN": "ghp_or_fine_grained_token"
  }
}
```

If your host uses a Claude-style `mcpServers` object, wrap the same `command`, `args`, and `env` under a `guild-agentdesk` key.

## Tool Contract

The executable server exposes:

- `guild_get_next_mandate`: returns the next open mandate from local files or GitHub source
- `guild_claim_mandate`: creates a local lease so another agent does not take the same task
- `guild_compile_context`: emits bounded role-specific context
- `guild_check_preflight`: checks write/run/network/secret actions against workspace rules
- `guild_request_approval`: records a human approval request
- `guild_publish_artifact`: records durable proof
- `guild_create_handoff`: creates a handoff proof artifact
- `guild_verify_mandate`: checks proof, handoff, and approvals
- `guild_close_mandate`: closes completed work
- `guild_export_replay_bundle`: exports the portable replay record

## GitHub Issue Intake

Create issues with the `agent:ready` label, then run:

```bash
GITHUB_TOKEN="$(gh auth token)" \
GITHUB_REPOSITORY="lucid-fdn/guild" \
GUILD_AGENTDESK_SOURCE="github" \
GUILD_CLI="$(pwd)/bin/guild" \
corepack pnpm --dir adapters/mcp exec guild-agentdesk-mcp
```

The agent can then call `guild_get_next_mandate` and `guild_claim_mandate`.
