# MCP Setup

Guild now exposes a TypeScript-first MCP stdio server through `guild-agentdesk mcp serve`.
The old Go binary MCP server remains available as a native fallback.

Use it when you want Codex, Claude, OpenClaw, OpenFang, or another MCP-capable agent host to fetch mandates, claim work, compile context, run preflight checks, request approvals, publish proof, verify completion, and export replay without a hosted service.

## 90-Second Agent Quickstart

```bash
corepack enable
corepack pnpm install
corepack pnpm run build:agentdesk-ts
node packages/agentdesk-cli/dist/index.js init
node packages/agentdesk-cli/dist/index.js mandate create "Update MCP docs" --writable "docs/**,packages/agentdesk-mcp/**"
node packages/agentdesk-cli/dist/index.js doctor
node packages/agentdesk-cli/dist/index.js mcp serve
```

The MCP host connects to `guild-agentdesk mcp serve` over stdio.

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

## Claude Desktop Config

```json
{
  "mcpServers": {
    "guild-agentdesk": {
      "command": "node",
      "args": ["/absolute/path/to/guild/packages/agentdesk-cli/dist/index.js", "mcp", "serve"],
      "env": {
        "GITHUB_REPOSITORY": "lucid-fdn/guild"
      }
    }
  }
}
```

If the agent should pull GitHub Issues labeled `agent:ready`, add a token:

```json
{
  "mcpServers": {
    "guild-agentdesk": {
      "command": "node",
      "args": ["/absolute/path/to/guild/packages/agentdesk-cli/dist/index.js", "mcp", "serve"],
      "env": {
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
command = "node"
args = ["/absolute/path/to/guild/packages/agentdesk-cli/dist/index.js", "mcp", "serve"]

[mcp_servers.guild-agentdesk.env]
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
  "command": "node",
      "args": ["/absolute/path/to/guild/packages/agentdesk-cli/dist/index.js", "mcp", "serve"],
  "env": {
    "GUILD_AGENTDESK_SOURCE": "github",
    "GITHUB_REPOSITORY": "lucid-fdn/guild",
    "GITHUB_TOKEN": "ghp_or_fine_grained_token"
  }
}
```

If your host uses a Claude-style `mcpServers` object, wrap the same `command`, `args`, and `env` under a `guild-agentdesk` key.

## Repo-Local TypeScript Server

The TypeScript MCP package remains useful for adapter development:

```bash
git clone https://github.com/lucid-fdn/guild.git
cd guild
corepack enable
corepack pnpm install
corepack pnpm run build:agentdesk-ts
node packages/agentdesk-cli/dist/index.js mcp serve
```

The public install target is `npx guild-agentdesk mcp serve` once the npm package is published.

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
node packages/agentdesk-cli/dist/index.js bootstrap github --repo lucid-fdn/guild

GITHUB_TOKEN="$(gh auth token)" \
GITHUB_REPOSITORY="lucid-fdn/guild" \
GUILD_AGENTDESK_SOURCE="github" \
node packages/agentdesk-cli/dist/index.js mcp serve
```

The agent can then call `guild_get_next_mandate` and `guild_claim_mandate`.
