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

These objects are intentionally small.
The goal is not to standardize an entire agent runtime.
The goal is to standardize the institutional layer above orchestrators.

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
