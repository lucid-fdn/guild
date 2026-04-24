# Roadmap

Guild is intentionally agent-contract-first. The order matters: local mandates and proof first, then adapters, then shared API and learning.

The current launch wedge is agent-first: every agent run should start with a mandate and end with proof. See [Agent-First Pivot Plan](AGENT_FIRST_PIVOT_PLAN.md).

## Now

- Keep `Taskpack`, `DRI Binding`, `Artifact`, and `Promotion Record` stable enough for early adopters.
- Make the bootstrap server easy to run locally.
- Keep examples, fixtures, runtime validation, and CI aligned.
- Make the CLI the primary non-UI adoption path.
- Make MCP and CLI the primary agent-facing surfaces.
- Ship local claim leases so multiple agents can safely pull from the same desk.
- Ship GitHub Issues ingestion for issues labeled `agent:ready`.
- Ship an executable MCP server around `agentdesk.yaml` and `.agentdesk/`.
- Add a workspace constitution file so humans can define rules once and agents can self-serve.
- Add preflight checks for path, command, tool, and approval decisions.
- Make OpenAPI the SDK generation contract.
- Prove the DRI + artifact-first model with small demos.
- Keep MCP/A2A bridges thin and orchestrator-agnostic.
- Keep LangGraph as the first real orchestrator adapter without making Guild a graph runtime.
- Keep the experience plane useful with or without a live shared API.
- Treat institutional memory and context routing as a core pillar, not a separate generic memory product.

## V1 Alpha

- Add generated TypeScript and Python SDK types from the JSON Schemas.
- Replace bootstrap SDKs with OpenAPI-generated SDKs once generation is reproducible and contributor-friendly.
- Add conformance tests that orchestrator adapters can run.
- Expand the current task detail UI into editable approvals and artifact inspection.
- Add adapter conformance profiles and badges for early ecosystem trust.
- Add OpenTelemetry trace IDs to API responses and smoke tests.
- Expand replay bundle export from single-task bundles to recursive task trees.
- Add a Postgres storage implementation behind the current service interfaces.
- Add memory registry metadata for artifact/record tags: agent, office, model, orchestrator, mandate, visibility, and promotion status.
- Add a first context compiler endpoint that emits role-specific context packs from taskpacks, artifacts, policies, and commons entries.

## V1 Beta

- Harden the MCP adapter against a live MCP SDK host.
- Harden the A2A adapter against a live A2A transport.
- Add the second real orchestrator adapter, likely OpenAI Agents SDK.
- Add replay bundles: taskpack, inputs, artifacts, decisions, and trace metadata.
- Add policy hooks for approval-required actions.
- Add benchmark runner worker for promotion candidates.
- Add pluggable retrieval backends for vector stores and customer memory systems without replacing the institutional registry.

## V1

- Publish the spec as the primary product artifact.
- Publish adapter conformance badge.
- Ship a replay UI that makes institutional learning visible.
- Ship promotion gates that compare candidate learnings against replay suites.
- Keep Guild orchestrator-agnostic: adapters integrate, but the core never becomes a runtime.
- Publish the institutional memory/context routing profile as part of adapter conformance.

## Later

- Add IPFS/Filecoin/Arweave adapters for public commons snapshots and tamper-evident institutional records.
- Add signed provenance chains for cross-organization shared learnings.
- Add decentralized reputation only after local registry, replay, and promotion gates are proven.
