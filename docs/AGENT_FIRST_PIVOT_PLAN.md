# Agent-First Pivot Plan

## Verdict

The broad "institution/control plane for agent companies" position is too close to Paperclip and too easy for Agent OS projects to absorb as a feature.

The stronger wedge is narrower:

```text
Every agent run should start with a mandate and end with proof.
```

This project should become the agent-first preflight, permission, context, proof, and handoff layer that works across existing task systems and agent runtimes.

## New Positioning

Working product frame:

```text
The work contract for autonomous agents.
```

Longer version:

```text
A local-first protocol, CLI, and MCP server that lets agents self-serve tasks, rules, context, approvals, proof artifacts, and handoffs from any workspace.
```

Avoid leading with:

- agent company
- org chart
- human dashboard
- control plane
- swarm
- memory
- agent OS

Lead with:

- mandate
- preflight
- permission
- scoped context
- proof
- handoff
- replay

## Immediate Pain

Agents already run inside repos, terminals, CI, IDEs, cloud workspaces, and orchestrators. Their work contract is usually informal.

The pain shows up as:

- the agent starts under-briefed
- the agent loads too much or too little context
- the agent does not know which files/tools are allowed
- the agent edits overlapping scope with another agent
- risky actions require human judgment but are not declared ahead of time
- handoffs are transcript dumps instead of structured packets
- completion is asserted in chat instead of proven with artifacts
- humans cannot later inspect why a task was done, by whom, and with what evidence

## Product Thesis

Humans should define intent and guardrails once. Agents should self-serve inside those boundaries.

```text
Human sets the workspace constitution.
Agent operates the desk.
```

The system should minimize human babysitting, not human responsibility.

Humans provide:

- mission
- task sources
- default allowances
- writable and forbidden scope
- approval rules
- success criteria
- escalation path

Agents consume:

- next mandate
- allowed scope
- compiled context
- preflight decision
- approval request path
- proof requirements
- handoff packet
- replay/export record

## Core User Stories

### Human/admin

As a workspace owner, I want to define guardrails once so agents can work without asking me every five minutes.

As a team lead, I want agents to pull tasks from GitHub/Linear/Jira/Slack/local files without adopting a new kanban.

As a reviewer, I want every agent run to produce proof artifacts before claiming completion.

As a security owner, I want risky actions to be checked against policy before the agent executes them.

### Agent/runtime

As an agent, I want to ask what my current mandate is.

As an agent, I want to know which files, commands, tools, and APIs I may use.

As an agent, I want a bounded context pack for my role and task.

As an agent, I want to request approval when policy blocks me.

As an agent, I want to publish proof and hand off to another agent without dumping my transcript.

### Integrator

As an orchestrator author, I want a neutral work packet and proof format that does not force my users into another Agent OS.

As a platform team, I want to run this locally, file-backed, or against a shared server depending on maturity.

## Non-Goals

- Do not build another kanban.
- Do not build another Agent OS.
- Do not build another vector memory platform.
- Do not require a hosted SaaS for v1.
- Do not require Go knowledge to use the project.
- Do not require humans to use a dashboard before agents can benefit.
- Do not compete head-on with Paperclip's autonomous company control plane.

## Product Shape

The product has five surfaces.

### 1. Workspace constitution

A small file committed to the repo or mounted into the workspace.

Default filename:

```text
agentdesk.yaml
```

Example:

```yaml
version: v1alpha1
workspace: lucid-fdn/app
mission: Ship reliable product changes with accountable AI agents.

defaults:
  max_runtime_minutes: 45
  max_cost_usd: 5
  context_budget_tokens: 12000

task_sources:
  - type: github_issues
    repo: lucid-fdn/app
    query: "label:agent:ready state:open"
  - type: local
    path: .agentdesk/mandates

scope:
  writable:
    - src/**
    - tests/**
    - docs/**
  forbidden:
    - .env
    - infra/prod/**
    - billing/**

approval_rules:
  - when: touches_forbidden_path
    require: human
  - when: runs_destructive_command
    require: human
  - when: changes_auth_or_payments
    require: human
  - when: pushes_to_main
    require: human

success_criteria:
  - Tests pass or failure is explained.
  - Every modified file is listed.
  - A proof artifact is attached.
  - A reviewer handoff is created.

escalation:
  default_owner: "@quentin"
  channels:
    - type: github_comment
    - type: cli_prompt
```

### 2. Mandate format

`Taskpack` becomes the mandate object.

Rename in docs/product language:

```text
Taskpack = portable mandate
```

It should answer:

- what is the objective?
- where did this task come from?
- who owns the outcome?
- what scope is allowed?
- what context budget applies?
- what approvals are required?
- what evidence proves completion?
- where should the agent publish proof?

### 3. CLI

The CLI is the default human and agent interface.

Initial command shape:

```bash
agentdesk init
agentdesk sources list
agentdesk next
agentdesk mandate show <id>
agentdesk context compile <id> --role coder --budget 12000
agentdesk preflight <id> --path src/auth/login.ts
agentdesk preflight <id> --command "rm -rf dist"
agentdesk approval request <id> --reason "Need to edit auth policy"
agentdesk proof add <id> --kind test_report --path ./test-results.xml
agentdesk handoff create <id> --to reviewer --summary ./handoff.md
agentdesk close <id> --proof ./proof.json
agentdesk replay export <id>
agentdesk verify <id>
```

The CLI should work in three modes:

- file-only mode for local single-agent workflows
- localhost mode against a local daemon
- remote mode against a team-provisioned server

### 4. MCP server

The MCP server is the agent-native surface.

Tools:

- `get_next_mandate`
- `get_mandate`
- `compile_context`
- `check_preflight`
- `claim_scope`
- `request_approval`
- `publish_proof`
- `create_handoff`
- `close_mandate`
- `export_replay`

This is the most important v1 surface because it lets Codex, Claude Code, Cursor, OpenClaw, OpenFang, and custom agents use the project without custom glue.

### 5. Optional server

The server remains useful, but it is not the first adoption requirement.

Use it for:

- shared team state
- durable approvals
- multi-agent lock/claim state
- audit/replay storage
- policy decisions across machines
- team-level task ingestion

Default adoption path:

```text
file/spec -> CLI/MCP -> local daemon -> self-hosted server -> optional hosted cloud
```

## Task Sources

Do not build a kanban. Ingest work from where humans already manage work.

V1 sources:

- local `.agentdesk/mandates/*.yaml`
- GitHub Issues
- CLI ad hoc tasks
- CI failures via file/stdin

Implemented GitHub ingestion command:

```bash
guild agentdesk next --source github --repo lucid-fdn/app --query "label:agent:ready state:open"
```

The adapter maps GitHub Issues into local mandates with deterministic IDs, source references, sanitized labels, priority labels, role labels, and scope labels.

V1.5 sources:

- Linear
- Jira
- Slack messages
- Pull request review comments
- TODO comments

Later sources:

- Paperclip tasks
- OpenFang/OpenClaw tasks
- Notion
- email
- incident systems

## Source-To-Mandate Mapping

### GitHub issue

```text
title -> mandate title
body -> objective and context hints
labels -> priority, role, approval profile
assignees -> proposed DRI
comments -> additional context
linked PRs -> artifacts
```

Recommended labels:

- `agent:ready`
- `agent:blocked`
- `agent:review`
- `agent:codex`
- `priority:p1`
- `approval:required`
- `scope:auth`
- `scope:docs`

### CI failure

```text
workflow name -> mandate title
failed job -> objective
logs -> context artifact
repo/branch/sha -> source refs
success criteria -> green rerun or explained failure
```

### Local file

```yaml
id: mandate_fix_auth_tests
title: Fix failing auth tests
objective: Diagnose and fix failing auth tests.
source:
  type: local
priority: p1
role_hint: coder
scope:
  writable:
    - src/auth/**
    - tests/auth/**
success_criteria:
  - Auth tests pass.
  - Changes are summarized.
```

## Policy And Permission Model

The policy engine should be simple and predictable.

Inputs:

- workspace constitution
- mandate
- actor identity
- requested action
- file path or command
- tool name
- risk tags

Output:

```json
{
  "decision": "allow",
  "reason": "Path is within writable scope.",
  "approval_required": false,
  "matched_rules": ["scope.writable.src"]
}
```

Decision enum:

- `allow`
- `deny`
- `needs_approval`
- `needs_handoff`

Preflight should support at least:

- path read
- path write
- command run
- external network call
- secret access
- git push
- dependency install
- production environment access

## Context Compiler

The context compiler is not generic memory.

It builds a bounded context pack for one mandate, role, and token budget.

Inputs:

- mandate
- workspace constitution
- source task content
- relevant files
- proof requirements
- previous artifacts
- selected commons entries
- role
- token budget

Output:

```json
{
  "mandate_id": "mandate_fix_auth_tests",
  "role": "coder",
  "budget_tokens": 12000,
  "must_read": [
    "tests/auth/login.test.ts",
    "src/auth/login.ts"
  ],
  "may_read": [
    "docs/auth.md"
  ],
  "forbidden": [
    ".env",
    "infra/prod/**"
  ],
  "summary": "Fix auth test failures without changing production secrets.",
  "proof_required": [
    "test_report",
    "changed_files",
    "handoff_summary"
  ]
}
```

Rules:

- references before hydration
- role-specific output
- no transcript dump by default
- include omitted-context reasons when possible
- keep private scratchpads out of institutional records unless explicitly published

## Proof Model

Every mandate should close with proof.

Proof artifact kinds:

- `test_report`
- `diff`
- `changed_files`
- `screenshot`
- `log_excerpt`
- `benchmark_result`
- `security_review`
- `handoff_summary`
- `decision_record`
- `human_approval`

Close criteria:

- required proof artifacts exist
- policy blockers are resolved
- handoff created if reviewer required
- replay bundle validates

## Handoff Model

Handoffs should be small packets, not transcripts.

Handoff fields:

- mandate id
- from actor
- to actor or role
- state summary
- decisions made
- open questions
- relevant artifacts
- next action requested
- blocked reasons

Example:

```yaml
mandate_id: mandate_fix_auth_tests
from: codex
to: reviewer
summary: Auth tests now pass locally after updating token expiry fixture.
decisions:
  - Kept production auth behavior unchanged.
artifacts:
  - proof/test-results.xml
  - proof/changed-files.json
open_questions:
  - Should fixture factory move to tests/helpers?
next_action: Review diff and approve close.
```

## Reuse From Current Repo

Keep and refocus:

- `Taskpack` schema as mandate schema
- `Artifact` schema as proof record
- `Approval Request` schema as approval object
- `Replay Bundle` schema as evidence export
- MCP adapter as the agent-native surface
- CLI validation and replay export
- OpenAPI as optional server contract
- file-backed local mode
- Postgres mode for team/shared state
- conformance profiles

De-emphasize:

- UI-first experience plane
- org-chart/institution dashboard
- promotion/commons as v1 headline
- broad "agent civilization" copy
- server-first onboarding

Move to later:

- commons browser
- promotion gates
- full governance UI
- marketplace
- decentralized public memory

## Architecture

```text
Human tools
  GitHub / Linear / Jira / Slack / local files
        |
        v
Task source adapters
        |
        v
Mandate registry
        |
        +--> CLI
        +--> MCP server
        +--> SDK
        |
        v
Agent runtime
  Codex / Claude Code / OpenFang / OpenClaw / LangGraph / CI
        |
        v
Preflight / context / approval / proof / handoff
        |
        v
Replay bundle and proof artifacts
```

Storage modes:

- file mode: `.agentdesk/`
- local daemon: HTTP over localhost with file storage
- team server: HTTP with Postgres/object storage
- later public archive: decentralized commons snapshots

## V0 Execution Plan

Goal:

```text
Make one coding agent run safer and more accountable in a repo without any hosted server.
```

Deliverables:

1. `agentdesk.yaml` schema and example.
2. Local mandate file schema.
3. `agentdesk init`.
4. `agentdesk next` from local files.
5. `agentdesk preflight` for path write and command risk.
6. `agentdesk context compile` with file refs and token budget.
7. `agentdesk proof add`.
8. `agentdesk close`.
9. `agentdesk replay export`.
10. MCP server exposing the same primitives.

Demo:

```bash
agentdesk init
agentdesk mandate create "Fix failing auth tests"
agentdesk next
agentdesk context compile mandate_123 --role coder
agentdesk preflight mandate_123 --path src/auth/login.ts --action write
agentdesk proof add mandate_123 --kind test_report --path test-results.xml
agentdesk close mandate_123
agentdesk replay export mandate_123
```

Success:

- no server required
- no dashboard required
- one agent can consume the mandate through MCP
- proof bundle validates

## V1 Execution Plan

Goal:

```text
Become the neutral preflight/proof layer for agent work in GitHub repos.
```

Deliverables:

1. GitHub Issues source adapter.
2. GitHub Actions integration.
3. `agentdesk verify` CI check.
4. PR comment summary with mandate, scope, approvals, and proof.
5. Approval flow via GitHub comments or CLI.
6. Scope claim/lock file to avoid agent collisions.
7. Adapter conformance profile.
8. Codex/Claude Code/OpenFang MCP setup examples.

GitHub check output:

```text
Agent Work Contract: passed
Mandate: Fix failing auth tests (#184)
DRI: codex
Scope: src/auth/**, tests/auth/**
Approvals: none required
Proof: test_report, changed_files, handoff_summary
Replay: attached
```

Implemented CI reporting command:

```bash
guild agentdesk verify --id <mandate-id> --github-report --replay-file .agentdesk/replay/replay.json
```

When running in GitHub Actions, the command writes to `GITHUB_STEP_SUMMARY` and posts a PR comment when `GITHUB_TOKEN`, `GITHUB_REPOSITORY`, and PR metadata are available.

## V1.5 Execution Plan

Goal:

```text
Make the work contract portable across task systems and runtimes.
```

Deliverables:

1. Linear source adapter.
2. Jira source adapter.
3. Slack request source adapter.
4. OpenFang adapter.
5. OpenClaw adapter.
6. LangGraph adapter refresh around mandate/preflight/proof.
7. Optional shared server for teams.
8. Postgres-backed approval and proof registry.

## V2 Execution Plan

Goal:

```text
Institutional learning emerges from proof and replay.
```

Deliverables:

1. Promotion candidates from repeated proof patterns.
2. Replay suites for mandate classes.
3. Commons entries for accepted learnings.
4. Policy suggestions from recurring approvals.
5. Public/decentralized commons snapshots.

This is where the old institution narrative returns, but as earned product value rather than launch positioning.

## Competitive Wedge

Paperclip:

- owns human-managed AI company/control plane
- likely dashboard/org/budget first
- we integrate as a work-contract/proof layer

OpenFang/OpenClaw:

- own runtime/Agent OS execution
- may add internal task records
- we stay neutral across runtimes and task sources

GitHub/Copilot/Coder:

- own GitHub-native coding delegation
- we stay portable across local agents, CI, OpenFang, Claude Code, Codex, and non-GitHub sources

Memory projects:

- own persistent context
- we own preflight, permission, proof, and bounded context packs

## Why This Can Get Stars

The repo can give developers an immediate aha moment:

```text
My agent finally knows what it is allowed to do.
```

Star-worthy artifacts:

- simple `agentdesk.yaml`
- useful CLI in five minutes
- MCP tools agents can call immediately
- GitHub issue-to-agent mandate demo
- PR check that proves agent work
- no hosted account required
- works with whichever agent is hot this month

## Naming Direction

The product name should be agent-first and concrete.

Better names for this wedge:

- `Mandate`
- `AgentDesk`
- `Workpack`
- `Taskpack`
- `Desk`
- `Brief`

Less ideal for this wedge:

- `Civitas`
- `Themis`
- `Guild`
- `Polity`

Reason:

The new wedge is immediate and operational. Mythic/institutional names can still appear in internals, but the launch name should communicate the job quickly.

Recommended default:

```text
Mandate
```

Reason:

- direct
- ownable as a primitive
- not another agent/org/swarm name
- maps to the core promise

Tagline:

```text
Mandate is the work contract for autonomous agents.
```

## North Star Metric

Number of agent runs that close with valid proof.

Supporting metrics:

- mandates created
- preflight checks performed
- approval requests resolved
- proof artifacts attached
- replay bundles exported
- GitHub issues completed through mandate flow
- runtimes connected through MCP

## Final Strategic Call

Do not throw away the repo.

Refocus it:

```text
from institution control plane
to agent-first work contract
```

The institution story becomes the long-term philosophy. The launch product becomes a sharp, immediate tool agents can actually use today.
