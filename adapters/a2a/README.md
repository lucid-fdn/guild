# A2A Adapter

[![Guild adapter-alpha](../../conformance/badges/guild-adapter-alpha.svg)](../../conformance/profiles/a2a.v1alpha1.json)

Package: `@guild/adapter-a2a`

This adapter maps Guild tasks and artifacts onto A2A interactions.

Responsibilities:

- discover remote agent capabilities via agent cards
- send and receive bounded work packets
- translate long-running remote task state into Guild run state
- preserve artifact references and provenance

The adapter should treat A2A as the transport and interoperability layer, not as the institutional layer itself.

Minimum compatibility target:

```bash
go run ./cli/cmd/guild conformance --base-url http://localhost:8080
```

The adapter is considered useful only when A2A task exchange keeps Guild's bounded work packets, DRI ownership, artifact references, and provenance intact.

## What Ships

The v1 alpha bridge exports:

- `actorFromAgentCard`
- `buildTaskpackFromA2ATask`
- `buildDriBindingFromA2ATask`
- `buildArtifactFromA2AResult`
- `toA2AArtifactMessage`
- `submitA2ATask`

The adapter intentionally treats A2A as the transport/interoperability layer and Guild as the institutional record.

Run checks:

```bash
pnpm --dir adapters/a2a run check
```
