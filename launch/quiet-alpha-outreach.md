# Quiet Alpha Outreach

Goal: ask 5-10 agent-framework and devtool builders for sharp feedback before broad launch.

## Targets

- LangGraph/LangChain builders who care about production agent workflows.
- CrewAI builders who hear orchestration pain from users.
- MCP client maintainers for Claude/Codex-style local tool setup.
- OpenFang/OpenClaw-style agent OS builders who need cleaner task intake.
- A2A/agent interoperability builders.
- Vercel AI SDK and AI devtool builders.
- Cline/Continue/aider-style coding-agent maintainers.
- GitHub Actions/devex maintainers interested in AI proof artifacts.
- Devtools founders building review, CI, or compliance products.
- OSS maintainers experimenting with agent-labeled issues.

## Short DM

We are quietly alpha-testing Guild: a local-first work contract for autonomous agents.

The idea is simple: every agent run starts with a mandate and ends with proof. GitHub Issues labeled `agent:ready` become claimable mandates; agents attach tests, changed files, handoff summaries, and replay; CI verifies the Agent Work Contract on PRs.

Repo: https://github.com/lucid-fdn/guild
Demo PR: https://github.com/lucid-fdn/guild/pull/9

Would love blunt feedback on whether this solves a real agent-workflow pain for your users.

## Longer Email

Subject: Quiet alpha feedback on agent work contracts

Hi,

We are testing Guild before a broader launch. It is not another orchestrator. It is a local-first work contract layer for agents:

- GitHub issue labeled `agent:ready` becomes a mandate.
- One agent claims the mandate.
- Scope, context, guardrails, and approval rules are explicit.
- The run ends with proof artifacts.
- CI posts an Agent Work Contract report to the PR.
- Replay preserves the audit trail.

Install:

```bash
go install github.com/lucid-fdn/guild/cli/cmd/guild@latest
guild agentdesk init
guild mcp serve
```

The real demo PR is here: https://github.com/lucid-fdn/guild/pull/9

I would value your toughest feedback: is this a missing primitive for agent frameworks, or should it be shaped differently?
