# Release Checklist

Use this before tagging a public release or making launch noise.

## Required

- `make release-check` passes.
- `make verify` passes from a clean checkout.
- `make e2e` passes against a fresh local server.
- README quickstart works on macOS and Linux.
- `docs/quickstart.md` matches the current commands and endpoints.
- Public examples validate against the JSON Schemas.
- Public examples validate through the Guild CLI.
- `examples/one-task-one-dri-commons/run.sh` passes.
- OpenAPI validates with `make lint-openapi`.
- Markdown local links validate with `make lint-docs`.
- Bootstrap fixture and public example references validate with `make lint-fixtures`.
- Bootstrap fixtures validate against the JSON Schemas.
- API conformance passes against the bootstrap server.
- API rejects unknown fields and invalid enum values.
- API rejects orphaned DRI bindings, artifacts, and promotion records.
- UI production build passes.
- No generated binaries, `.next`, `node_modules`, or local `data` directories are committed.

## Product

- README clearly says Guild is not an orchestrator.
- README clearly says every agent run starts with a mandate and ends with proof.
- AgentDesk local workflow includes `next`, `claim`, `context`, `preflight`, `proof`, `verify`, and `replay`.
- MCP docs show the single-binary `guild mcp serve` server path.
- `guild agentdesk doctor` passes in an initialized workspace.
- README clearly says what is implemented now versus planned.
- Launch copy has one concrete demo path.
- README links to the demo GIF, quickstart, canonical example, and launch assets.
- Roadmap explains the path to adapters, replay, and promotion gates.

## Launch

- Create a short demo GIF or video.
- Publish one canonical example of DRI ownership.
- Publish one canonical example of a promotion record.
- Publish one canonical example of approval-gated commons promotion.
- Open starter issues for MCP adapter, A2A adapter, replay UI, and SDK generation.
- Tag the release only after CI passes on the default branch.
