# Guild

Guild is the orchestrator-agnostic institution layer for AI teams.

![Guild demo](launch/demo.gif)

It sits above existing runtimes and frameworks and adds the primitives they are missing:

- `Taskpack`: a portable, bounded handoff format for agent work
- `DRI`: one accountable owner per task
- `Artifact`: durable, typed outputs instead of chat as the system of record
- `Commons`: a governed library of promoted team learnings
- `Promotion Gates`: replay and benchmark rules that decide what the institution learns
- `Replay Bundle`: portable evidence for inspecting or evaluating a task run

Guild is not another multi-agent orchestrator.

Guild does not ask teams to replace OpenFang, Hermes, LangGraph, CrewAI, OpenAI Agents SDK, or custom stacks.
Guild gives those stacks ownership, institutional memory, review, replay, and collective learning.

## Start Here

- Read the [Quickstart](docs/quickstart.md) to run Guild locally in minutes.
- Run the canonical [one task, one DRI, commons example](examples/one-task-one-dri-commons/README.md).
- Use `make simulation` to verify the full institution loop end-to-end.
- Review the [launch assets](launch/README.md) for landing copy, demo script, starter issues, and the GIF.

## Why Guild

Humans outperform through institutions, not just communication.

The same pattern applies to AI teams:

- agents need clear ownership
- handoffs need bounded context
- important work needs durable artifacts
- institutions improve by promoting proven behaviors, not by trusting every new idea

The winning architecture is not "1,000 agents thinking together."
It is "many agents coordinating through scoped work, ownership, review, and promoted knowledge."

## Product Thesis

Bring your own orchestrator.
Guild adds structure, ownership, artifacts, and institutional memory.

Core positioning:

- MCP gives agents tools
- A2A gives agents interoperability
- AG-UI gives agents user-facing interactivity
- Guild gives agents institutions

## What Guild Standardizes

Guild aims to make eight portable objects boring, stable, and reusable:

1. `Taskpack`
- objective
- inputs
- artifact references
- constraints
- permissions
- context budget
- acceptance criteria

2. `DRI Binding`
- one accountable owner
- reviewers
- specialists
- approvers
- escalation rules

3. `Artifact`
- typed output
- provenance
- lineage
- evaluation status
- storage location

4. `Promotion Record`
- candidate learning
- evidence
- benchmark result
- acceptance or rejection

5. `Governance Policy`
- rules for risky actions
- approval requirements
- institution-level constraints

6. `Approval Request`
- human approval state
- approver decisions
- policy context

7. `Promotion Gate`
- required replay runs
- required metric deltas
- approval requirement

8. `Commons Entry`
- accepted institutional learning
- artifact reference
- scope and status

Guild also defines `Replay Bundle` as the portable export format that packages a task, ownership, artifacts, and promotion evidence.

## Design Principles

- Orchestrator-agnostic from day one
- Artifact-first, not chat-first
- One task, one owner
- Bounded fan-out
- Learning off the hot path
- Policy and approval built in
- Spec-first before framework-specific convenience
- Replayability as a first-class feature

## Reference Architecture

Guild has four planes:

1. Experience Plane
- approvals
- replay
- trace UI
- task dashboards
- commons browser

2. Control Plane
- task registry
- DRI assignment
- workflow state machine
- policy engine
- scheduler
- approvals

3. Execution Plane
- orchestrator adapters
- context compiler
- model routing integration
- MCP and A2A gateways

4. Learning Plane
- trace ingestion
- candidate extraction
- benchmark runner
- promotion gates
- commons registry

## What Guild Is Not

- not a chatbot framework
- not a new agent runtime
- not a marketplace in v1
- not a dense peer-to-peer swarm
- not a full autonomous self-modifying institution in the hot path

## Day 1 Demo

The first demo should be legible in under 60 seconds:

1. A user opens a task.
2. Guild assigns one DRI and two support roles: reviewer and skeptic.
3. The DRI delegates two narrow subtasks through `Taskpack`s.
4. Each subtask produces typed artifacts.
5. Guild shows the full trace, ownership tree, approvals, and artifacts.
6. The run is replayed.
7. A candidate learning is proposed and benchmarked.
8. The learning is accepted into the commons only if it improves results.

## Standards Posture

Guild is designed to compose with:

- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/specification/draft)
- [Agent2Agent (A2A)](https://a2a-protocol.org/dev/specification/)
- [AG-UI](https://github.com/ag-ui-protocol/ag-ui)
- [OpenTelemetry GenAI semantic conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/)

Guild should avoid inventing new wire protocols when an existing one is good enough.
Its job is to define the institutional layer above them.

## Repository Layout

```text
guild/
  README.md
  .editorconfig
  .gitignore
  Makefile
  package.json
  pnpm-workspace.yaml
  cli/
    cmd/
      guild/
        main.go
  pkg/
    spec/
      models.go
      validate/
  openapi/
    guild.v1alpha1.yaml
    README.md
  spec/
    README.md
    common.schema.json
    taskpack.schema.json
    dri-binding.schema.json
    artifact.schema.json
    promotion-record.schema.json
    replay-bundle.schema.json
    examples/
  docs/
    IMPLEMENTATION_PLAN.md
    LAUNCH_NARRATIVE.md
    LANDING_PAGE_COPY.md
    adr/
      0001-v1-architecture.md
    quickstart.md
  server/
    README.md
  sdk/
    GENERATION.md
    typescript/
      README.md
      src/
        spec.ts
    python/
      README.md
  ui/
    README.md
  adapters/
    typescript/
      README.md
      src/
        index.ts
    mcp/
      README.md
      src/
        index.ts
    a2a/
      README.md
      src/
        index.ts
    langgraph/
      README.md
      src/
        index.ts
  conformance/
    adapter-profile.schema.json
    profiles/
  workers/
    evaluator/
      README.md
  examples/
    README.md
    one-task-one-dri-commons/
    typescript-adapter-core/
  deploy/
    README.md
    docker-compose.local.yml
    otel-collector.yaml
```

## v1 Scope

Guild v1 target scope:

- `Taskpack`, `DRI`, `Artifact`, and `Promotion Record` schemas
- `Replay Bundle` schema for portable replay and evaluation evidence
- a single-node control plane
- Postgres-backed task registry
- object storage-backed artifact system
- approval flow
- replay and trace UI
- MCP and A2A adapters

Guild v1 will not include:

- autonomous commons promotion without human review
- a public marketplace
- federated identity across companies
- settlement, escrow, or billing rails

Current bootstrap implementation:

- JSON Schemas for the four core public objects
- draft 2020-12 spec validation for examples and bootstrap fixtures
- file-backed local control plane for fast OSS iteration
- `GET/POST` endpoints for `Taskpack`, `DRI Binding`, `Artifact`, and `Promotion Record`
- seeded fixture data for local development and UI testing
- runtime validation for schema version, UUIDs, enums, timestamps, URIs, labels, and token budgets
- runtime validation for replay bundle object validity and internal references
- strict JSON decoding that rejects unknown fields
- file-backed referential integrity for institutions, taskpacks, artifacts, DRI bindings, and promotion records
- Postgres storage behind the current service interfaces with runtime migrations
- local object-storage backend for artifact metadata mirroring
- recursive replay bundle export for task trees through the API and CLI
- replay suite runner that emits benchmark-result artifacts, skill-candidate artifacts, and human-review promotion records
- durable evaluation job queue with queued/running/succeeded/failed states
- in-process evaluator worker for replay suites, with deterministic run endpoint for tests
- governance policies, human approval requests, promotion gates, and commons registry entries
- full simulation script for the one task/DRI/artifact/replay/promoted-learning story
- neutral TypeScript adapter core for orchestrator-specific wrappers
- MCP-style bridge package with Guild tool definitions and handlers
- A2A-style bridge package with task/result mappers
- LangGraph adapter package with a node-shaped bridge for real graph integration
- adapter conformance profiles and a reusable `adapter-alpha` badge
- checked TypeScript adapter-core example that submits a run and exports replay
- Next.js experience plane with live/offline dashboard, task detail, DRI graph, artifact graph, replay timeline, approval inbox shell, and commons panel
- Go tests, UI build checks, spec linting, and a local API smoke test

Planned production hardening:

- managed object storage backend for artifact payloads
- Redis and NATS-backed distributed queues
- dead-letter dashboards and multi-worker leasing

## Local Development

Prerequisites:

- Go 1.23+
- Node.js 22+
- pnpm through Corepack

Install dependencies:

```bash
corepack enable
make install
```

Run all checks:

```bash
make verify
```

Run the full pre-release gate:

```bash
make release-check
```

Validate the HTTP API contract:

```bash
make lint-openapi
```

Validate local Markdown links:

```bash
make lint-docs
```

Validate fixture and example references:

```bash
make lint-fixtures
```

SDK generation is vendor-neutral. OpenAPI is the source of truth for the HTTP API; Speakeasy is only an optional scaffold. See `sdk/GENERATION.md`.

Validate one object through the CLI:

```bash
go run ./cli/cmd/guild validate --kind taskpack --file spec/examples/taskpack.example.json
go run ./cli/cmd/guild validate --kind replay-bundle --file spec/examples/replay-bundle.example.json
```

Run the server:

```bash
make run-server
```

Run the UI in another shell:

```bash
make run-ui
```

Run all adapter checks:

```bash
make check-adapters
```

Adapter packages:

- `@guild/adapter-core` provides neutral builders for `Taskpack`, `DRI Binding`, and `Artifact`.
- `@guild/adapter-mcp` exposes MCP-style tool definitions and handlers for taskpacks, DRIs, artifacts, and replay bundles.
- `@guild/adapter-a2a` maps A2A-style task/result envelopes into Guild institutional records.
- `@guild/adapter-langgraph` provides a LangGraph-compatible node that submits Taskpack, DRI, and Artifact records while returning a graph state patch.

Useful endpoints:

```text
GET  http://localhost:8080/healthz
GET  http://localhost:8080/api/v1/status
GET  http://localhost:8080/api/v1/taskpacks
GET  http://localhost:8080/api/v1/replay/taskpacks/{taskpack_id}
POST http://localhost:8080/api/v1/taskpacks
POST http://localhost:8080/api/v1/dri-bindings
POST http://localhost:8080/api/v1/artifacts
POST http://localhost:8080/api/v1/promotion-records
POST http://localhost:8080/api/v1/governance-policies
POST http://localhost:8080/api/v1/approval-requests
POST http://localhost:8080/api/v1/promotion-gates
POST http://localhost:8080/api/v1/commons-entries
```

Run conformance checks against a running Guild-compatible API:

```bash
go run ./cli/cmd/guild conformance --base-url http://localhost:8080
```

Run the full local e2e adopter journey:

```bash
make e2e
```

Export a portable replay bundle:

```bash
go run ./cli/cmd/guild replay-export \
  --base-url http://localhost:8080 \
  --taskpack-id 4e1fe00c-6303-453c-8cb6-2c34f84896e4
```

Run a replay/evaluation suite and open a promotion candidate:

```bash
go run ./cli/cmd/guild replay-suite \
  --base-url http://localhost:8080 \
  --suite examples/replay-suite.example.json
```

Queue a replay/evaluation job through the control plane worker path:

```bash
go run ./cli/cmd/guild eval-submit \
  --base-url http://localhost:8080 \
  --suite examples/replay-suite.example.json \
  --wait
```

Run the full launch simulation:

```bash
make simulation
```

Run the canonical launch example:

```bash
examples/one-task-one-dri-commons/run.sh
```

Launch assets live in [launch](launch/README.md).

Optional local infrastructure:

- Postgres
- Redis
- NATS JetStream
- MinIO
- OpenTelemetry Collector

Use `make dev-up` to bring up the optional infrastructure and `make dev-down` to stop it. The default server is file-backed and does not require this stack; set `GUILD_STORAGE_DRIVER=postgres` and `GUILD_DATABASE_URL` to use Postgres.

## Verification

`make verify` runs:

- frozen pnpm install
- JSON Schema validation for public examples and bootstrap fixtures
- OpenAPI validation for the bootstrap control-plane API
- Markdown local link validation
- fixture and public example reference validation
- adapter conformance profile validation
- generated TypeScript SDK spec type drift check
- CLI validation for public examples
- Go unit tests
- Go server build
- Go CLI build
- TypeScript SDK check
- Python SDK compile check
- TypeScript adapter checks
- TypeScript example check
- TypeScript check
- Next.js production build
- local API smoke test

`make e2e` additionally starts a fresh server and exercises CLI conformance,
CLI replay export, replay-suite promotion candidate creation, durable evaluation
job submission, Python SDK reads, the TypeScript adapter example, and replay
bundle validation.

`make release-check` runs `make verify`, `make e2e`, and release guardrails for
ignored source paths, executable scripts, and generated local artifacts.

## Initial Build Sequence

1. Freeze the core spec objects.
2. Build the single-node control plane.
3. Build the trace and replay UI.
4. Add MCP and A2A adapters.
5. Add learning-plane candidate extraction.
6. Add benchmark-driven promotion gates.

## Category

Guild is "the institution layer for AI teams."

Short version:

One task. One owner. Durable artifacts. Shared learning.
