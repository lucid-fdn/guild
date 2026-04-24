# Guild Conformance

Conformance is how an orchestrator, adapter, or control plane proves it respects the Guild institutional contract.

The first conformance profile is intentionally small:

- expose `/healthz`
- expose `/api/v1/status`
- list at least one valid `Taskpack`
- accept a valid `Artifact`
- reject a `DRI Binding` whose `taskpack_id` does not exist
- reject unknown JSON fields on public spec objects
- reject malformed path IDs
- reject unsupported item methods
- export a replay bundle with task ownership and artifacts

Run against a local server:

```bash
go run ./cli/cmd/guild conformance --base-url http://localhost:8080
```

Run the full local bootstrap smoke test:

```bash
make smoke
```

Run the full adopter e2e path:

```bash
make e2e
```

This is not the final adapter conformance suite. It is the first executable boundary: any implementation claiming Guild compatibility should preserve strict object validation and institutional referential integrity.

## Adapter Profiles

Adapter profiles describe what an adapter supports and which local checks back its compatibility claim.

Current profile:

- `adapter-alpha`

Current adapter profile files:

- `profiles/mcp.v1alpha1.json`
- `profiles/a2a.v1alpha1.json`
- `profiles/langgraph.v1alpha1.json`

Reusable badge:

![Guild adapter-alpha](badges/guild-adapter-alpha.svg)

Validate profiles:

```bash
make lint-adapter-profiles
```

The CLI uses the public Go spec packages:

- `github.com/lucid-fdn/guild/pkg/spec`
- `github.com/lucid-fdn/guild/pkg/spec/validate`
