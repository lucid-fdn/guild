# MCP Adapter

[![Guild adapter-alpha](../../conformance/badges/guild-adapter-alpha.svg)](../../conformance/profiles/mcp.v1alpha1.json)

Package: `@guild/adapter-mcp`

This package exposes Guild AgentDesk to MCP-aware environments.
For users, the preferred alpha install path is the single Go binary: `guild mcp serve`.
This TypeScript package remains useful for adapter development, embedding MCP tool definitions, and testing host compatibility.

For copy-paste Claude, Codex, OpenFang, OpenClaw, and generic MCP host configs, see [MCP Setup](../../docs/MCP_SETUP.md).

Responsibilities:

- expose selected AgentDesk operations to MCP-aware hosts
- let agents fetch, claim, verify, close, and replay mandates without a hosted service
- map MCP tool activity into Guild proof and replay records
- keep permission and approval semantics explicit

The adapter should avoid inventing new MCP semantics when existing host behavior is sufficient.

Minimum compatibility target:

```bash
go run ./cli/cmd/guild conformance --base-url http://localhost:8080
```

The adapter is considered useful only when it can create valid mandates, compile bounded context, check preflight decisions, publish proof `Artifact`s, and preserve Guild's DRI and approval semantics instead of hiding them inside tool-call transcripts.

## Executable Local Server

Run from a workspace that has already been initialized with `guild agentdesk init`:

```bash
guild mcp serve
```

Example host configuration:

```json
{
  "mcpServers": {
    "guild-agentdesk": {
      "command": "guild",
      "args": ["mcp", "serve"],
      "env": {
        "GITHUB_REPOSITORY": "lucid-fdn/guild"
      }
    }
  }
}
```

Useful environment variables:

- `GUILD_AGENTDESK_SOURCE`: set to `github` to pull mandates from GitHub Issues
- `GITHUB_TOKEN`: token used by the GitHub source adapter
- `GITHUB_REPOSITORY`: default `owner/repo` for GitHub source ingestion

The server speaks JSON-RPC over MCP stdio frames and implements:

- `initialize`
- `tools/list`
- `tools/call`

## What Ships

The v1 alpha bridge exports:

- `guildMcpTools`: closed JSON-schema tool definitions for MCP-style hosts
- `GuildMcpBridge`: a handler that dispatches tool calls into a Guild-compatible client
- `createGuildMcpBridge(client)`: convenience factory for host integration
- `LocalAgentDeskClient`: a CLI-backed client for `agentdesk.yaml` and `.agentdesk/`
- `guild-agentdesk-mcp`: an executable local MCP server

The TypeScript executable remains available for package-level testing:

```bash
GUILD_CLI="guild" corepack pnpm --dir adapters/mcp exec guild-agentdesk-mcp
```

Supported tools:

- `guild_get_next_mandate`
- `guild_claim_mandate`
- `guild_create_taskpack`
- `guild_assign_dri`
- `guild_publish_artifact`
- `guild_compile_context`
- `guild_check_preflight`
- `guild_request_approval`
- `guild_create_handoff`
- `guild_verify_mandate`
- `guild_close_mandate`
- `guild_export_replay_bundle`

Run checks:

```bash
pnpm --dir adapters/mcp run check
```
