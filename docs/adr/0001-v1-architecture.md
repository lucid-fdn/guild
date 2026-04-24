# ADR 0001: Guild v1 Architecture

## Status

Accepted

## Date

2026-04-24

## Context

Guild needs to be:

- agent-first: every run starts with a mandate and ends with proof
- orchestrator-agnostic
- scalable under bounded multi-agent workloads
- performant on control-plane operations
- understandable by open-source contributors
- practical to self-host

The project must avoid becoming a new agent runtime.
It must sit above runtimes and standardize work-contract primitives first, then institutional/commons primitives after the agent workflow is useful.

## Decision

Guild v1 will be built as:

- a monorepo
- a Go control plane
- a TypeScript UI and primary SDK
- a Python SDK for ecosystem reach
- a single deployable backend for v1, with modular internal boundaries
- a local AgentDesk workflow and executable MCP server as the primary adoption path

## Why

### Monorepo

We need tight iteration between:

- spec
- server
- UI
- SDKs
- adapters

A monorepo keeps schema changes, docs, SDK types, and reference implementations aligned.

### Go control plane

Go is selected for:

- strong concurrency for scheduling and event handling
- simple deployment
- operational familiarity
- readable codebase for infra-minded contributors

Rejected alternatives:

- Python for the full control plane
  - too much pressure on async correctness and long-term maintainability
- Rust for v1
  - excellent performance, but slower contributor onboarding and higher iteration cost

### TypeScript UI and SDK

TypeScript is mandatory for:

- frontend
- schema-driven developer tooling
- modern OSS integrations

### Python SDK

Python is mandatory because the AI ecosystem expects it.
Even if the control plane is not Python, the adoption layer must be.

## Architecture

Guild v1 has four planes:

### 1. Experience Plane

- task views
- DRI views
- artifact explorer
- approval inbox
- replay timeline

### 2. Control Plane

- task registry
- DRI binding service
- policy engine
- approval service
- scheduler
- audit service

### 3. Execution Plane

- orchestrator adapters
- MCP bridge
- A2A bridge
- context compilation
- artifact IO
- institutional memory routing

### 4. Learning Plane

- candidate extraction
- benchmark runner
- promotion workflow
- commons metadata

## Storage Decisions

### Postgres

Used for:

- task metadata
- DRI bindings
- approvals
- policies
- commons metadata
- audit logs

Reason:

- transactional integrity
- predictable querying
- broad operational familiarity

### Object storage

Used for:

- artifact bodies
- taskpack archives
- replay bundles
- benchmark corpora

Reason:

- cheap durable storage
- better fit than large blobs in Postgres

### NATS JetStream

Used for:

- asynchronous task events
- approval notifications
- replay jobs
- promotion jobs

Reason:

- simple operations
- strong fit for control-plane eventing
- easier to reason about than building everything from polling tables

### Redis

Used for:

- hot caches
- locks and leases
- rate limits
- ephemeral coordination state

## Core Object Model

Guild standardizes exactly four top-level objects in v1:

- `Taskpack`
- `DRI Binding`
- `Artifact`
- `Promotion Record`

All other backend data structures are implementation details unless explicitly promoted into the public spec.

## Request Flow

### Task submission

1. Client submits `Taskpack`
2. Server validates schema
3. Server persists metadata and emits `guild.task.created`
4. DRI binding is created or attached
5. Scheduler dispatches to adapter or waits for approval

### Artifact emission

1. Runtime or adapter uploads artifact body to object storage
2. Runtime registers artifact metadata with Guild
3. Guild persists artifact metadata and lineage
4. Guild emits `guild.artifact.created`

### Promotion flow

1. Candidate artifact is marked as promotable
2. Evaluation worker runs benchmark suite
3. Evidence is attached to `Promotion Record`
4. Human or policy decision accepts or rejects the candidate

## Context Strategy

Guild will not treat full message history as the primary coordination format.

Instead:

- task context is compiled from structured inputs
- artifacts are referenced before content is hydrated
- role-specific payloads are generated per task
- institutional memory is routed by mandate, office, visibility, tags, and provenance
- agent-local scratchpads remain outside the canonical institutional record unless explicitly published as artifacts

This is the only path that scales to large agent populations.

Guild therefore owns institutional memory, not generic agent memory. The control plane keeps the registry of mandates, offices, records, approvals, replay evidence, commons entries, and promotion decisions. Vector stores, customer data stores, and decentralized archives remain pluggable backends.

The default v1 memory stack is Postgres plus object storage. DePIN-style storage can be added later for public commons snapshots, tamper-evident records, and community-owned learnings, but it must not sit on the hot path for private task execution.

## Security Decisions

- approvals are first-class objects, not side effects
- artifact checksums are stored
- adapters operate with least privilege
- MCP tool descriptions are treated as untrusted
- audit logs are append-only from the application perspective

## Consequences

### Positive

- cleaner adoption story
- strong performance profile for the control plane
- easier self-hosting
- good contributor ergonomics

### Negative

- more cross-language coordination
- more initial schema discipline required
- learning plane may feel delayed compared with hypey demos

## Rejected Paths

### Build a new orchestrator

Rejected because it collapses the differentiation.

### Full microservice split in v1

Rejected because it adds operational and cognitive overhead too early.

### Chat-first memory model

Rejected because it does not scale and makes replay brittle.

### DePIN-first memory model

Rejected for v1 because hot-path agent context needs low latency, simple local development, and private-by-default storage. Decentralized storage remains a future backend for public commons and verifiable provenance.

## Follow-Up ADRs

Future ADRs should cover:

- auth model
- multitenancy model
- policy engine implementation
- replay execution semantics
- benchmark suite packaging
