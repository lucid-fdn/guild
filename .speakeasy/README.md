# Optional Speakeasy Scaffold

Guild is vendor-neutral. OpenAPI is the SDK generation contract; Speakeasy is one optional generator path.

The checked-in `.speakeasy/gen.yaml` is a convenience scaffold for teams that want to try Speakeasy SDK generation. It must not become required for local development, CI, or contribution.

Current repo behavior:

- CI validates the OpenAPI document with Redocly.
- Speakeasy generation is intentionally not run in CI.
- `make verify` does not require a Speakeasy account, token, or CLI.
- The existing SDK clients remain bootstrap-only until generated SDK output can be reproduced in a contributor-friendly way.

Manual optional path:

```bash
speakeasy quickstart
```

Use `openapi/guild.v1alpha1.yaml` as the source document.

For the vendor-neutral SDK policy, see `sdk/GENERATION.md`.
