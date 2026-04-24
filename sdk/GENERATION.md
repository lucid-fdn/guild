# SDK Generation

Guild is vendor-neutral.

The source of truth for the HTTP API is:

```text
openapi/guild.v1alpha1.yaml
```

The source of truth for portable Guild objects is:

```text
spec/*.schema.json
```

## Current State

- The TypeScript SDK is a bootstrap client.
- TypeScript object types are generated from the canonical JSON Schemas to prevent drift.
- CI checks generated TypeScript object types without mutating files.
- OpenAPI is validated in CI and is ready to feed SDK generators.
- The checked-in SDKs are intentionally simple until generated SDK output is reproducible without a hosted vendor workflow.

## Generator Posture

Guild should work with any reasonable OpenAPI SDK generator, including:

- OpenAPI Generator
- Speakeasy
- Fern
- Stainless
- Kiota
- custom internal tooling

No contributor should need a commercial account, cloud workflow, or vendor-specific CLI to build, test, or contribute to Guild.

## Speakeasy

The `.speakeasy/` directory is an optional scaffold for teams that want to try Speakeasy's SDK generator. It is not required by `make verify`, CI, or local development.

If using Speakeasy manually, use:

```bash
speakeasy quickstart
```

and select:

```text
openapi/guild.v1alpha1.yaml
```

Expected contribution rule:

- commit generated SDK output only if it can be regenerated from a documented local command
- keep `openapi/guild.v1alpha1.yaml` as the API source of truth
- keep `spec/*.schema.json` as the portable object source of truth
- do not make CI depend on hosted generator credentials
- do not replace hand-written bootstrap clients until generated clients are smaller, clearer, or materially easier to maintain

## Suggested Generated SDK Workflow

For a generator such as Speakeasy, Fern, Stainless, Kiota, or OpenAPI Generator:

1. Update the OpenAPI contract.
2. Run `make lint-openapi`.
3. Generate SDK output into a temporary branch.
4. Compare public method names against the bootstrap clients.
5. Add a deterministic generation command to this file.
6. Add tests equivalent to the current SDK tests.
7. Only then replace or publish generated clients.

Guild should be easy for open-source contributors first.
Generator quality matters, but contributor reproducibility matters more.

## Near-Term Direction

The intended path is:

1. Keep OpenAPI generator-ready.
2. Keep JSON Schema as the portable object contract.
3. Replace bootstrap SDK clients with generated SDKs only when the generation process is reproducible and contributor-friendly.
4. Keep all generation optional unless it can run locally without proprietary credentials.

## Local Commands

Regenerate checked-in TypeScript object types:

```bash
make generate-typescript-spec
```

Check that generated TypeScript object types are current:

```bash
make check-typescript-spec
```
