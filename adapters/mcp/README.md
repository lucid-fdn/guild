# MCP Adapter

[![Guild adapter-alpha](../../conformance/badges/guild-adapter-alpha.svg)](../../conformance/profiles/mcp.v1alpha1.json)

Package: `@guild/adapter-mcp`

This adapter translates Guild concepts into MCP-aware environments where appropriate.

Responsibilities:

- expose selected Guild operations to MCP-aware hosts
- map MCP tool activity into Guild traces
- keep permission and approval semantics explicit

The adapter should avoid inventing new MCP semantics when existing host behavior is sufficient.

Minimum compatibility target:

```bash
go run ./cli/cmd/guild conformance --base-url http://localhost:8080
```

The adapter is considered useful only when it can create valid `Taskpack`s, publish valid `Artifact`s, and preserve Guild's DRI and approval semantics instead of hiding them inside tool-call transcripts.

## What Ships

The v1 alpha bridge exports:

- `guildMcpTools`: closed JSON-schema tool definitions for MCP-style hosts
- `GuildMcpBridge`: a handler that dispatches tool calls into the Guild control plane
- `createGuildMcpBridge(client)`: convenience factory for host integration

Supported tools:

- `guild_create_taskpack`
- `guild_assign_dri`
- `guild_publish_artifact`
- `guild_export_replay_bundle`

Run checks:

```bash
pnpm --dir adapters/mcp run check
```
