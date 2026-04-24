# Landing Page Copy

## Hero

Headline:

Every agent run starts with a mandate and ends with proof.

Subheadline:

Guild is an agent-first work contract for autonomous agents: mandates, claims, bounded context, preflight guardrails, approvals, proof artifacts, and replay.

Primary CTA:

Run AgentDesk

Secondary CTA:

Read the spec

Support line:

Local-first CLI. Executable MCP server. GitHub Issues in, verified proof out.

## Hero Proof Bar

- One mandate per run
- Local claim locks
- Context before action
- Proof before done
- Replay attached

## Section 1

Heading:

Agents should not start from a vague chat prompt.

Body:

They need the same operational basics a strong human teammate expects:

- What is the mandate?
- What can I change?
- What context is relevant?
- What needs approval?
- What proof makes this complete?
- How does another agent replay the work?

Guild makes those answers machine-readable.

## Section 2

Heading:

The agent work lifecycle.

Card 1:

Title:

Mandate

Copy:

A portable task contract with objective, writable scope, context budget, permissions, references, and acceptance criteria.

Card 2:

Title:

Claim

Copy:

A local lease that prevents multiple agents from picking the same task from the same repo.

Card 3:

Title:

Context

Copy:

A bounded context pack compiled for the agent's role, not a full transcript or repo dump.

Card 4:

Title:

Preflight

Copy:

Allow, deny, or request approval before risky file writes, commands, network calls, dependency installs, or prod actions.

Card 5:

Title:

Proof

Copy:

Typed artifacts for test reports, changed files, screenshots, logs, handoffs, approvals, and reviews.

Card 6:

Title:

Replay

Copy:

A portable record of the mandate, proof, approvals, and handoff trail.

## Section 3

Heading:

Works where your agents already work.

Body:

Guild does not ask you to replace your runtime.
It gives runtimes a common work contract.

Use it with:

- Codex
- Claude
- OpenClaw
- OpenFang
- LangGraph
- CrewAI
- OpenAI Agents SDK
- custom MCP or A2A systems

## Section 4

Heading:

GitHub Issues become agent-ready mandates.

Body:

Humans already create tasks in GitHub.
Guild lets agents consume them directly.

```bash
GITHUB_TOKEN=... guild agentdesk next --source github --repo lucid-fdn/app --query "label:agent:ready state:open"
guild agentdesk claim --id <mandate-id> --agent codex
```

No new kanban ritual required.

## Section 5

Heading:

An MCP server agents can plug into.

Body:

Run `guild mcp serve` in any initialized workspace and MCP hosts can fetch mandates, claim work, compile context, check preflight, request approvals, publish proof, verify completion, and export replay.

```bash
guild mcp serve
```

## Section 6

Heading:

CI can enforce the contract.

Body:

Use `guild agentdesk verify --github-report` as a GitHub Actions check.

The PR report says:

```text
Agent Work Contract: passed
Mandate: ...
Proof: test_report, changed_files, handoff_summary
Approvals: resolved
Replay: attached
```

## Section 7

Heading:

Not another orchestrator.

Body:

Orchestrators decide how agents execute.
Guild defines what work they are allowed to take, how they prove completion, and how the run can be replayed.

That makes Guild useful above many runtimes instead of competing with them.

## Section 8

Heading:

The long game: agent institutions.

Body:

Once every run has mandates, proof, and replay, teams can build higher-order systems:

- accountable roles
- governed approvals
- promotion gates
- commons of accepted learnings
- institutional memory that routes context instead of dumping it

The launch wedge is operational.
The compounding value is institutional.

## FAQ

Question:

Is Guild another agent framework?

Answer:

No. Guild is an agent work-contract layer. It defines mandates, claims, context, preflight, proof, and replay for agents running in other systems.

Question:

Do I need to self-host a server?

Answer:

No for the first workflow. AgentDesk runs locally from `agentdesk.yaml` and `.agentdesk/`. Teams can add the shared API, UI, and production storage later.

Question:

Why not just use GitHub Issues?

Answer:

GitHub Issues are great for humans. Guild turns them into machine-readable mandates with scope, context, approval rules, proof requirements, claims, verification, and replay.

Question:

Why not just use shared memory?

Answer:

Shared memory does not solve ownership, allowed scope, approvals, proof, or replay. Guild routes bounded context and durable artifacts instead of asking every agent to read everything.

Question:

Can Guild work with my existing stack?

Answer:

Yes. That is the point. Use the CLI, MCP server, A2A adapter, LangGraph adapter, or API.

## Footer CTA

Headline:

Give every agent a mandate. Require proof before done.

Primary CTA:

Run AgentDesk

Secondary CTA:

Browse the spec
