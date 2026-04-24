# TypeScript SDK

The TypeScript SDK is the first-class client library for:

- schema-safe object creation
- control-plane API access
- browser and server integrations
- adapter development

Planned packages:

- `@guild/spec`
- `@guild/client`
- `@guild/adapter-core`

The SDK should generate or derive types directly from the canonical JSON Schemas.

Guild's SDK generation posture is vendor-neutral. See `../GENERATION.md`.

Current implementation:

- `src/spec.ts` is generated from `spec/*.schema.json`
- `src/index.ts` re-exports the generated spec types and provides the HTTP client
- `make check-typescript-spec` fails when generated spec types are stale
- the HTTP client supports status, list, get, create, and replay-export operations

Regenerate types from the repository root:

```bash
make generate-typescript-spec
```

Check generated types without mutating files:

```bash
make check-typescript-spec
```

Check the SDK:

```bash
make check-sdk
```

The SDK check runs both TypeScript type-checking and the client behavior tests.
