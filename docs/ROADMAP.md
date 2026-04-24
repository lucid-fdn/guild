# Roadmap

Guild is intentionally spec-first. The order matters: portable objects first, then adapters, then learning.

## Now

- Keep `Taskpack`, `DRI Binding`, `Artifact`, and `Promotion Record` stable enough for early adopters.
- Make the bootstrap server easy to run locally.
- Keep examples, fixtures, runtime validation, and CI aligned.
- Make the CLI the primary non-UI adoption path.
- Make OpenAPI the SDK generation contract.
- Prove the DRI + artifact-first model with small demos.
- Keep MCP/A2A bridges thin and orchestrator-agnostic.
- Keep LangGraph as the first real orchestrator adapter without making Guild a graph runtime.
- Keep the experience plane useful with or without a live control plane.

## V1 Alpha

- Add generated TypeScript and Python SDK types from the JSON Schemas.
- Replace bootstrap SDKs with OpenAPI-generated SDKs once generation is reproducible and contributor-friendly.
- Add conformance tests that orchestrator adapters can run.
- Expand the current task detail UI into editable approvals and artifact inspection.
- Add adapter conformance profiles and badges for early ecosystem trust.
- Add OpenTelemetry trace IDs to API responses and smoke tests.
- Expand replay bundle export from single-task bundles to recursive task trees.
- Add a Postgres storage implementation behind the current service interfaces.

## V1 Beta

- Harden the MCP adapter against a live MCP SDK host.
- Harden the A2A adapter against a live A2A transport.
- Add the second real orchestrator adapter, likely OpenAI Agents SDK.
- Add replay bundles: taskpack, inputs, artifacts, decisions, and trace metadata.
- Add policy hooks for approval-required actions.
- Add benchmark runner worker for promotion candidates.

## V1

- Publish the spec as the primary product artifact.
- Publish adapter conformance badge.
- Ship a replay UI that makes institutional learning visible.
- Ship promotion gates that compare candidate learnings against replay suites.
- Keep Guild orchestrator-agnostic: adapters integrate, but the core never becomes a runtime.
