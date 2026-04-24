# Guild Spec

Guild standardizes eight canonical objects:

- `Taskpack`
- `DRI Binding`
- `Artifact`
- `Promotion Record`
- `Governance Policy`
- `Approval Request`
- `Promotion Gate`
- `Commons Entry`

Guild also defines one portable evidence object:

- `Replay Bundle`

The agent-first surface adds three operational objects:

- `Workspace Constitution`
- `Context Pack`
- `Preflight Decision`

These objects are intentionally small.
The goal is not to standardize an entire agent runtime.
The goal is to standardize the institutional layer above orchestrators.

## Institutional Memory Boundary

Guild owns institutional memory, not generic agent memory.

The specs should make it possible to answer:

- who owned the mandate
- which office/role acted
- which artifacts became public records
- which approvals represented consent
- which replay evidence became institutional memory
- which promoted commons entries became shared culture
- which context pack an agent was allowed to see

The specs should not require every agent to share a single memory store or transcript. Agent-local scratchpads, vector databases, customer data stores, and decentralized archives are implementation backends. Guild records the institutional metadata, provenance, visibility, and replayable artifacts that let those systems interoperate.

## Rules

- JSON Schema draft 2020-12
- explicit `schema_version` on every object
- additive changes for minor versions
- breaking changes only on major version bumps
- canonical examples must validate in CI

## Compatibility

Guild objects should be usable from:

- orchestrator-specific runtimes
- A2A bridges
- MCP-aware applications
- control planes
- evaluation systems
- artifact stores

## Versioning

Current version line:

- `v1alpha1`

Recommended release progression:

- `v1alpha1` for early adopters and reference integrations
- `v1beta1` when at least three independent implementations pass conformance
- `v1` when the field set is stable and examples are battle-tested in production

## Canonical Flow

1. A client or orchestrator creates a `Taskpack`.
2. Guild binds the task to a `DRI Binding`.
3. Agents produce one or more `Artifact`s.
4. A `Governance Policy` and `Approval Request` capture human control when risk crosses policy boundaries.
5. A `Promotion Gate` decides which benchmark evidence is required before learning is accepted.
6. A `Promotion Record` captures how an institution learned from the run.
7. A `Commons Entry` publishes the accepted learning to the institution.
8. A `Replay Bundle` exports the task, ownership, artifacts, and promotion evidence for replay or evaluation.

## Agent-First Flow

1. A human commits a `Workspace Constitution` with mission, scope, allowances, approval rules, and task sources.
2. An agent asks for the next mandate from local files, GitHub Issues, CI, or another source.
3. The agent compiles a `Context Pack` for its role and token budget.
4. The agent runs `Preflight Decision` checks before edits, commands, network calls, or risky actions.
5. The agent publishes proof as `Artifact`s.
6. The mandate closes only after required proof exists.
7. A `Replay Bundle` exports the mandate, proof, and decisions.
