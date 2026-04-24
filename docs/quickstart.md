# Quickstart

This guide gets Guild running locally and walks through the core story:

One task. One DRI. Durable artifacts. Replay. Human-approved learning. Commons.

## Prerequisites

- Go 1.23+
- Node.js 22+
- Corepack-enabled pnpm

## Install

```bash
corepack enable
make install
```

## Run The Control Plane

```bash
make run-server
```

In another terminal:

```bash
curl -fsS http://localhost:8080/healthz
curl -fsS http://localhost:8080/api/v1/status
```

The bootstrap server starts with seeded data for the canonical Guild story.

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

The fastest way to verify the full institutional loop is:

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

