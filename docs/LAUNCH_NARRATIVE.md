# Guild Launch Narrative

## Category

Guild is an agent work-contract layer.

It is not another orchestrator, agent OS, kanban board, or shared-memory product.
It is the small operational contract every agent should consume before it starts and produce evidence against before it stops.

## Core Story

Every agent run starts with a mandate and ends with proof.

Today, agents are often dropped into work with:

- vague instructions
- too much repo context
- unclear ownership
- no local claim or lock
- ad hoc approvals
- proof scattered across chat, shell output, and PR comments

Guild turns that into a boring, inspectable lifecycle:

```text
task source -> mandate -> claim -> context pack -> preflight -> work -> proof -> verify -> replay
```

The first wedge is immediate pain relief for teams already using autonomous agents.
Humans can keep creating tasks where they already work, especially GitHub Issues.
Agents can consume those tasks directly through CLI or MCP.

## The One-Liner

Every agent run starts with a mandate and ends with proof.

## The 3-Step Pitch

1. Give the agent one clear mandate with allowed scope and success criteria.
2. Make the agent claim the mandate, compile bounded context, and preflight risky actions.
3. Require proof artifacts and replay before the work is considered done.

## What To Avoid In Messaging

- "swarm intelligence"
- "agent control plane"
- "autonomous company"
- "multi-agent orchestration"
- "self-improving agents"
- "communism for agents"
- "freemason-trained agents"

These frames are crowded, vague, or too easy for bigger agent OS projects to absorb.

## What To Emphasize

- mandate
- claim
- bounded context
- preflight guardrails
- human approval only when needed
- proof artifacts
- verification
- replay
- GitHub-native task intake
- MCP-native agent consumption

## Launch Assets

### GitHub README

Must communicate:

- the one-line contract
- the AgentDesk local workflow
- the MCP server
- GitHub Issues ingestion
- GitHub Actions verification
- why this is not another orchestrator

### Demo Video

Show:

- a GitHub issue labeled `agent:ready`
- `guild-agentdesk next --source github`
- `guild-agentdesk claim --id ...`
- context compilation and preflight
- proof artifacts added by the agent
- `guild-agentdesk verify --github-report`
- replay export

### Architecture Post

Title idea:

"Agents need work contracts, not bigger chat histories"

### Spec Post

Title idea:

"Introducing Taskpack: a mandate format for agent work"

### Visual Explainer

One diagram showing:

- GitHub Issues / local tasks on the left
- Guild AgentDesk in the middle
- Codex, Claude, OpenClaw, OpenFang, LangGraph, and custom agents on the right through CLI/MCP
- proof and replay below

## Audience Order

### 1. Builders already using autonomous coding agents

They feel the pain first: duplicated tasks, missing proof, unclear handoffs, and giant context dumps.

### 2. Agent framework authors

They need a neutral contract their orchestrators can consume instead of inventing local task/proof semantics.

### 3. Engineering teams experimenting with AI workers

They need guardrails without forcing humans to manage a new dashboard all day.

## Proof Points Needed For Credibility

1. `agentdesk` works locally with no server.
2. MCP hosts can fetch, claim, and verify mandates.
3. GitHub Issues labeled `agent:ready` become mandates.
4. GitHub Actions can verify proof and publish a PR report.
5. Replay bundles can be exported and inspected.

## Draft HN Angle

Title:

Show HN: Guild, work contracts for autonomous agents

Body direction:

- every agent run starts with a mandate and ends with proof
- not another orchestrator
- local-first CLI plus executable MCP server
- turns GitHub Issues into agent-consumable mandates
- verifies tests, changed files, handoffs, approvals, and replay in CI

## Draft X Angle

Post:

Agents should not start from a vague chat prompt.

They should start from a mandate:

- objective
- allowed scope
- bounded context
- preflight rules
- required proof
- replay

We built Guild: every agent run starts with a mandate and ends with proof.

## Competitive Story

Do not attack OpenFang, Hermes, Codex, Claude, LangGraph, CrewAI, or OpenAI Agents SDK.

Instead:

- they help agents execute
- Guild gives each run a portable work contract

That avoids a framework war and makes Guild useful to every runtime.
