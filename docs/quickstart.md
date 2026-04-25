# Quickstart

This guide gets Guild running locally and walks through the core story:

Every agent run starts with a mandate and ends with proof.

## Prerequisites

- Node.js 22+
- Corepack-enabled pnpm

## Install

TypeScript-first repo alpha:

```bash
corepack enable
corepack pnpm install
corepack pnpm run build:agentdesk-ts
node packages/agentdesk-cli/dist/index.js init
node packages/agentdesk-cli/dist/index.js doctor
```

The public npm target is:

```bash
npx guild-agentdesk init
```

The legacy Go CLI remains available as a native fallback, not the primary agent-facing surface.

## Run AgentDesk Locally

AgentDesk is the fastest path because it does not require a server.
It stores the work contract in `agentdesk.yaml` and `.agentdesk/`.

```bash
node packages/agentdesk-cli/dist/index.js init
node packages/agentdesk-cli/dist/index.js mandate create "Fix failing auth tests" --writable "src/auth/**,tests/auth/**"
node packages/agentdesk-cli/dist/index.js next
node packages/agentdesk-cli/dist/index.js claim --id <mandate-id> --agent codex
node packages/agentdesk-cli/dist/index.js context compile --id <mandate-id> --role coder
node packages/agentdesk-cli/dist/index.js preflight --id <mandate-id> --action write --path src/auth/login.ts
node packages/agentdesk-cli/dist/index.js proof add --id <mandate-id> --kind test_report --path test-results.xml
node packages/agentdesk-cli/dist/index.js proof add --id <mandate-id> --kind changed_files --path changed-files.json
node packages/agentdesk-cli/dist/index.js handoff create --id <mandate-id> --to reviewer --summary "Ready for review"
node packages/agentdesk-cli/dist/index.js doctor --id <mandate-id>
node packages/agentdesk-cli/dist/index.js verify --id <mandate-id>
node packages/agentdesk-cli/dist/index.js replay export --id <mandate-id>
```

## Bootstrap GitHub Intake

In an existing GitHub repo:

```bash
GITHUB_TOKEN="$(gh auth token)" \
node packages/agentdesk-cli/dist/index.js bootstrap github --repo lucid-fdn/your-repo

GITHUB_TOKEN="$(gh auth token)" \
node packages/agentdesk-cli/dist/index.js issue create "Fix docs typo" \
  --repo lucid-fdn/your-repo \
  --scope "docs/**" \
  --acceptance "Docs are updated and proof is attached."
```

This creates:

- `agentdesk.yaml` with local and GitHub issue task sources
- `.agentdesk/` working directories
- an `agent:ready` issue template
- portable GitHub Actions workflows that install the pinned Guild CLI
- `agent:ready` and priority labels when `GITHUB_TOKEN` is set

## Copy-Paste Demo

After bootstrap, create one GitHub issue from the CLI, then run:

```bash
GITHUB_TOKEN="$(gh auth token)" \
node packages/agentdesk-cli/dist/index.js issue create "Fix docs typo" \
  --repo lucid-fdn/your-repo \
  --scope "docs/**" \
  --acceptance "Docs are updated and proof is attached."

GITHUB_TOKEN="$(gh auth token)" \
node packages/agentdesk-cli/dist/index.js next --source github --repo lucid-fdn/your-repo

node packages/agentdesk-cli/dist/index.js claim --id <mandate-id> --agent codex
node packages/agentdesk-cli/dist/index.js context compile --id <mandate-id> --role coder
node packages/agentdesk-cli/dist/index.js preflight --id <mandate-id> --action write --path docs/example.md
node packages/agentdesk-cli/dist/index.js proof add --id <mandate-id> --kind test_report --path test-results.xml
node packages/agentdesk-cli/dist/index.js proof add --id <mandate-id> --kind changed_files --path changed-files.json
node packages/agentdesk-cli/dist/index.js handoff create --id <mandate-id> --to reviewer --summary "Ready for review"
node packages/agentdesk-cli/dist/index.js verify --id <mandate-id>
node packages/agentdesk-cli/dist/index.js replay export --id <mandate-id>
```

Open a PR with `.agentdesk/**` committed and the generated Agent Work Contract workflow will post the verification report.

## Connect An MCP Host

Use the single binary from any initialized workspace:

```bash
node packages/agentdesk-cli/dist/index.js mcp serve
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
