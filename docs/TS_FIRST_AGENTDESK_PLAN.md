# TypeScript-First AgentDesk Plan

Guild's public surface is moving to TypeScript-first because the agent ecosystem expects npm-native CLIs, MCP servers, GitHub tooling, and SDKs.

The goal is not to rewrite the whole historical platform. The goal is to port and sharpen the wedge:

```text
GitHub issue -> mandate -> claim -> context -> preflight -> proof -> verify -> replay -> MCP
```

## Package Layout

- `packages/agentdesk-core`: local file-backed AgentDesk behavior.
- `packages/agentdesk-github`: GitHub issue intake, issue creation, PR/Actions reporting.
- `packages/agentdesk-mcp`: stdio MCP server and tool contract.
- `packages/agentdesk-cli`: the `guild-agentdesk` npx front door.

## What Is In Scope

- Local `agentdesk.yaml` and `.agentdesk/` files.
- Mandate creation and loading.
- Local claim locks.
- Context pack compilation.
- Preflight policy checks.
- Approval request/resolve records.
- Proof artifacts.
- Handoffs.
- Verification reports.
- Replay bundle export.
- GitHub issue creation and GitHub issue-to-mandate sync.
- MCP tools for agent hosts.

## What Is Deliberately Out Of Scope

- Re-porting the old server/control plane before user demand exists.
- Rebuilding the UI as the main product.
- Postgres/object storage/event bus runtime as the default path.
- Institution/commons automation beyond the proof/replay foundation.

## Go Status

The Go implementation remains available as a native fallback and behavior reference.
It should not be the primary README path once the npm package is published.
It is excluded from GitHub language stats via `.gitattributes` because the public surface is now TypeScript-first.

## Definition Of Done For The Migration

From a clean checkout:

```bash
corepack pnpm install
corepack pnpm run build:agentdesk-ts
node packages/agentdesk-cli/dist/index.js init
node packages/agentdesk-cli/dist/index.js mandate create "Fix docs" --writable "docs/**"
node packages/agentdesk-cli/dist/index.js next
node packages/agentdesk-cli/dist/index.js claim --id <mandate-id>
node packages/agentdesk-cli/dist/index.js proof add --id <mandate-id> --kind test_report --path test-results.xml
node packages/agentdesk-cli/dist/index.js proof add --id <mandate-id> --kind changed_files --path changed-files.json
node packages/agentdesk-cli/dist/index.js handoff create --id <mandate-id> --to reviewer --summary "Ready"
node packages/agentdesk-cli/dist/index.js verify --id <mandate-id>
node packages/agentdesk-cli/dist/index.js replay export --id <mandate-id>
node packages/agentdesk-cli/dist/index.js mcp serve
```
