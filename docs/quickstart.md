# Quickstart

This guide gets Guild running locally and walks through the core story:

Every agent run starts with a mandate and ends with proof.

## Prerequisites

- Go 1.23+
- Node.js 22+
- Corepack-enabled pnpm

## Install

Public alpha install:

```bash
go install github.com/lucid-fdn/guild/cli/cmd/guild@latest
guild agentdesk init
guild agentdesk doctor
```

Repo development install:

```bash
corepack enable
make install
```

## Run AgentDesk Locally

AgentDesk is the fastest path because it does not require a server.
It stores the work contract in `agentdesk.yaml` and `.agentdesk/`.

```bash
go run ./cli/cmd/guild agentdesk init
go run ./cli/cmd/guild agentdesk mandate create "Fix failing auth tests" --writable "src/auth/**,tests/auth/**"
go run ./cli/cmd/guild agentdesk next
go run ./cli/cmd/guild agentdesk claim --id <mandate-id> --agent codex
go run ./cli/cmd/guild agentdesk context compile --id <mandate-id> --role coder
go run ./cli/cmd/guild agentdesk preflight --id <mandate-id> --action write --path src/auth/login.ts
go run ./cli/cmd/guild agentdesk proof add --id <mandate-id> --kind test_report --path test-results.xml
go run ./cli/cmd/guild agentdesk proof add --id <mandate-id> --kind changed_files --path changed-files.json
go run ./cli/cmd/guild agentdesk handoff create --id <mandate-id> --to reviewer --summary "Ready for review"
go run ./cli/cmd/guild agentdesk doctor --id <mandate-id>
go run ./cli/cmd/guild agentdesk verify --id <mandate-id>
go run ./cli/cmd/guild agentdesk replay export --id <mandate-id>
```

## Connect An MCP Host

Use the single binary from any initialized workspace:

```bash
guild mcp serve
```

Claude Desktop, Codex, OpenFang, OpenClaw, and generic MCP host examples are in [MCP Setup](MCP_SETUP.md).

## Run The Shared API

```bash
make run-server
```

In another terminal:

```bash
curl -fsS http://localhost:8080/healthz
curl -fsS http://localhost:8080/api/v1/status
```

The bootstrap server starts with seeded data for the longer replay, approval, promotion, and commons story.

## Open The Experience Plane

In a second terminal:

```bash
make run-ui
```

Then open:

```text
http://localhost:3000
```

The UI shows task ownership, artifacts, replay state, approvals, promotion gates, and commons entries.

## Run The Real Simulation

The fastest way to verify the full agent work-contract loop is:

```bash
make simulation
```

This starts a fresh server and verifies:

- API conformance
- governance policy creation
- promotion gate creation
- approval request creation
- commons entry creation
- replay/evaluation job execution
- benchmark-result artifact creation
- skill-candidate artifact creation
- replay bundle export

Expected output:

```text
simulation-ok http://127.0.0.1:<port>
```

## Inspect The Replay Bundle

With the server running:

```bash
go run ./cli/cmd/guild replay-export \
  --base-url http://localhost:8080 \
  --taskpack-id 4e1fe00c-6303-453c-8cb6-2c34f84896e4 \
  --file replay.json
```

The bundle contains the taskpack, DRI binding, produced artifacts, and promotion evidence.

## Run The Release Gate

Before publishing a branch or release:

```bash
make release-check
```

That command validates specs, OpenAPI, docs, fixtures, SDKs, adapters, the UI build, smoke, e2e, and the full simulation.
