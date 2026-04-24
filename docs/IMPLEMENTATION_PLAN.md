# Guild Implementation Plan

## Purpose

This document turns the Guild thesis into a concrete build plan for a v1 product that is:

- orchestrator-agnostic
- scalable
- performant
- security-aware
- easy to adopt
- easy to extend

## Product Scope

Guild v1 is the institution layer above orchestrators.

It standardizes and operationalizes:

- `Taskpack`
- `DRI Binding`
- `Artifact`
- `Promotion Record`

Guild v1 ships a reference control plane and UI that make those objects useful in practice.

## Product Goals

1. Make handoffs portable.
2. Make ownership explicit.
3. Make outputs durable and replayable.
4. Make review and approval natural.
5. Make adoption possible without replacing existing orchestrators.

## Non-Goals For v1

1. Build a new orchestrator.
2. Build a public agent marketplace.
3. Build full identity and settlement rails.
4. Allow autonomous self-promotion into the commons without human review.
5. Solve every multi-agent UX problem on day one.

## Product Pillars

### 1. Portable Work

Any orchestrator should be able to create and consume a `Taskpack`.

### 2. Clear Accountability

Every task has exactly one DRI.
Reviewers and specialists may contribute, but they do not own the final outcome.

### 3. Artifact-First Execution

Artifacts are the durable record.
Messages coordinate work; artifacts carry outputs.

### 4. Replayable Institutions

Runs must be inspectable, replayable, and evaluable.

### 5. Promotion-Gated Learning

Institutions do not learn by accident.
They learn by proving that a new behavior improves benchmarks.

## User Roles

### Platform Engineer

Needs:

- adapters to existing orchestrators
- policy control
- observability
- replay
- deployment simplicity

### Agent Builder

Needs:

- easy `Taskpack` creation
- local examples
- strong SDKs
- clear schemas
- simple integration points

### Team Lead / Operator

Needs:

- visibility into task ownership
- approval and escalation controls
- review flow
- evidence for promoted learnings

## High-Level Roadmap

### Phase 0: Spec Freeze

Duration:

- 2 weeks

Deliverables:

- `Taskpack` schema
- `DRI Binding` schema
- `Artifact` schema
- `Promotion Record` schema
- example objects
- spec README
- compatibility and versioning rules

Exit criteria:

- schemas validate
- examples validate
- no unresolved naming ambiguity in core fields

### Phase 1: Single-Node Control Plane

Duration:

- 4 weeks

Deliverables:

- task registry service
- DRI assignment service
- artifact metadata service
- approval service
- basic replay trace ingestion
- Postgres migrations
- object storage integration

Exit criteria:

- can create a task, bind a DRI, produce artifacts, and view the result in the UI

Bootstrap note:

- the first OSS implementation may use file-backed persistence for speed and portability
- the public object model and API shape should remain compatible with the planned Postgres-backed control plane

### Phase 2: Experience Plane

Duration:

- 3 weeks

Deliverables:

- task list and task detail views
- artifact viewer
- DRI ownership graph
- run trace viewer
- approval inbox

Exit criteria:

- a user can understand who owns a task, what happened, and what artifacts were produced

### Phase 3: Adapter Layer

Duration:

- 4 weeks

Deliverables:

- TypeScript SDK
- Python SDK
- MCP adapter
- A2A adapter
- one orchestrator reference adapter

Suggested first orchestrator adapter:

- OpenAI Agents SDK or LangGraph

Exit criteria:

- at least two distinct runtimes can create valid `Taskpack`s and submit artifacts into Guild

### Phase 4: Replay And Evaluation

Duration:

- 3 weeks

Deliverables:

- run replay pipeline
- benchmark suite runner
- candidate learning artifact type
- human-reviewed promotion workflow

Exit criteria:

- one run can produce a candidate learning
- benchmark evidence can be attached
- candidate can be manually accepted or rejected

## Workstreams

### Workstream A: Spec

Owner:

- product + platform

Tasks:

- finalize core nouns
- write schema docs
- define compatibility rules
- add example fixtures
- define canonical validation commands

### Workstream B: Control Plane

Owner:

- backend

Tasks:

- task state machine
- DRI assignment logic
- approval endpoints
- artifact metadata endpoints
- run registry
- policy hooks

### Workstream C: Experience Plane

Owner:

- frontend

Tasks:

- task dashboard
- task detail page
- artifact detail page
- replay timeline
- approvals
- commons browser shell

### Workstream D: Adapters

Owner:

- SDK and platform

Tasks:

- TypeScript client
- Python client
- A2A bridge
- MCP bridge
- orchestrator adapters

### Workstream E: Learning Plane

Owner:

- platform + evaluation

Tasks:

- candidate extraction job
- benchmark runner
- promotion workflow
- commons metadata
- attribution views

## Detailed v1 Feature Set

### Feature 1: Task Creation

Functional requirements:

- accept a valid `Taskpack`
- validate schema at ingest
- assign internal task id and status
- link the task to a project or institution

API shape:

- `POST /v1/taskpacks`
- `GET /v1/taskpacks/{taskpack_id}`
- `GET /v1/taskpacks`

### Feature 2: DRI Binding

Functional requirements:

- assign one owner
- allow reviewers and specialists
- allow reassignment with audit trail
- support escalation rules

API shape:

- `POST /v1/dri-bindings`
- `PATCH /v1/dri-bindings/{dri_binding_id}`
- `GET /v1/dri-bindings/{dri_binding_id}`

### Feature 3: Artifact Registry

Functional requirements:

- register artifact metadata
- support object storage pointers
- support lineage
- support evaluation status

API shape:

- `POST /v1/artifacts`
- `GET /v1/artifacts/{artifact_id}`
- `GET /v1/taskpacks/{taskpack_id}/artifacts`

### Feature 4: Approval Flow

Functional requirements:

- open approval requests
- allow approve / deny / expire
- link approvals to tasks and actions
- keep a full audit log

API shape:

- `POST /v1/approvals`
- `POST /v1/approvals/{approval_id}/approve`
- `POST /v1/approvals/{approval_id}/deny`

### Feature 5: Replay

Functional requirements:

- record run metadata and artifact lineage
- replay deterministic control-plane paths
- re-run evaluation on stored artifacts and fixtures

API shape:

- `POST /v1/replays`
- `GET /v1/replays/{replay_id}`

### Feature 6: Candidate Promotion

Functional requirements:

- register candidate learning artifacts
- attach benchmark evidence
- route through manual decision

API shape:

- `POST /v1/promotion-records`
- `GET /v1/promotion-records/{promotion_record_id}`

## Data Model

### Primary relational tables

- `institutions`
- `projects`
- `taskpacks`
- `dri_bindings`
- `task_assignments`
- `artifacts`
- `artifact_edges`
- `runs`
- `approvals`
- `promotion_records`
- `commons_entries`
- `audit_events`

### Object storage prefixes

- `artifacts/`
- `taskpacks/`
- `replays/`
- `benchmarks/`
- `commons/`

### Event subjects

- `guild.task.created`
- `guild.task.assigned`
- `guild.task.status.changed`
- `guild.artifact.created`
- `guild.approval.requested`
- `guild.approval.resolved`
- `guild.replay.started`
- `guild.replay.finished`
- `guild.promotion.created`
- `guild.promotion.decided`

## Performance Targets

### API targets

- p95 read latency under 150ms for metadata endpoints
- p95 write latency under 250ms for control-plane endpoints
- artifact upload path delegated to object storage presigned URLs

### System targets

- support 10k task records per institution without UI degradation
- support 1k concurrently active tasks in a single-node dev cluster
- support 1k agents only through bounded delegation trees, not free chat meshes

### Queueing targets

- scheduler must support lease-based retries
- approval and replay workloads must be asynchronous
- candidate extraction must never block task completion

## Context Strategy

Guild must ship a context compiler contract even if the first implementation is thin.

Inputs:

- taskpack
- actor role
- artifact refs
- permissions
- relevant commons entries
- project policy

Outputs:

- scoped payload for the runtime
- token budget
- artifact hydration list
- action policy

Rules:

- artifact refs before transcript content
- summaries are hints, not the source of truth
- no implicit full-history dump

## Security Baseline

1. Explicit approval for sensitive capabilities
2. Audit every policy decision
3. Store checksums for artifacts
4. Sign webhook or event ingress where applicable
5. Apply least privilege to adapters
6. Treat MCP tool descriptions as untrusted
7. Keep orchestrator credentials out of artifact payloads

## Developer Experience

### Day-1 setup

- `make dev-up`
- `pnpm run lint:spec`
- one sample app in `examples/`

### Required DX features

- CLI validation for all spec objects
- clear example fixtures
- generated client types from schemas
- copy-paste adapter templates
- local replay from saved fixtures

## Testing Strategy

### Spec tests

- JSON Schema validation
- compatibility snapshots

### Backend tests

- unit tests for state transitions
- integration tests against Postgres, Redis, NATS, and MinIO

### Adapter tests

- conformance fixture suite
- replay fixture suite

### UI tests

- end-to-end task creation
- approval flow
- replay flow

## Open Source Rollout Strategy

### What must be public on day one

- the four spec objects
- one reference server
- one reference UI
- one TypeScript SDK
- one Python SDK
- at least one orchestrator adapter

### What can stay rough

- enterprise auth
- hosted cloud offering
- advanced governance packs

## Team Shape

Lean ideal founding build team:

- 1 product-minded systems architect
- 1 backend engineer
- 1 frontend / full-stack engineer
- 1 SDK / integrations engineer

## Exit Criteria For v1

Guild v1 is done when:

- two independent orchestrators can use Guild without rewrites
- a team can create tasks, assign DRIs, produce artifacts, and replay runs
- approvals work
- candidate learnings can be manually promoted with benchmark evidence
- the system is understandable from the UI in under five minutes
