# OpenAPI

Guild uses two complementary contract layers:

- `spec/*.schema.json` defines the portable Guild objects.
- `openapi/guild.v1alpha1.yaml` defines the bootstrap control-plane HTTP API.

Why both exist:

- JSON Schema is the stable object standard for orchestrators, adapters, and conformance fixtures.
- OpenAPI is the API contract for SDK generation, docs, request/response envelopes, operation names, and error shapes.

Validate the OpenAPI contract:

```bash
make lint-openapi
```

SDK generation direction:

- Any OpenAPI SDK generator should be able to consume `openapi/guild.v1alpha1.yaml`.
- Current hand-written SDKs are bootstrap clients until generated SDKs can replace them in a reproducible, contributor-friendly way.
- Object types in the current TypeScript SDK are generated from JSON Schema as an interim drift guard.
- Speakeasy is an optional supported path, not a project requirement. See `sdk/GENERATION.md`.
