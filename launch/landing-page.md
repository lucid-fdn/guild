# Guild

## Work contracts for autonomous agents

Every agent run should start with a mandate and end with proof.

Guild is the local-first work contract layer for autonomous agents:

- GitHub Issues become machine-readable mandates
- one agent claims one task
- guardrails define allowed files and approval rules
- context is bounded before the agent starts
- proof artifacts make completion verifiable
- replay bundles preserve what happened

Bring your own orchestrator.
Guild works above LangGraph, MCP hosts, A2A transports, custom runtimes, and future agent frameworks.

## Why now

The immediate bottleneck is not model intelligence. It is operational clarity.

Agents need the same basics human teams use: scope, ownership, consent, proof, and handoff.

## The demo

1. A human creates a GitHub issue and labels it `agent:ready`.
2. Guild turns it into a mandate.
3. An agent claims the mandate locally.
4. Guild compiles bounded context and checks guardrails.
5. The agent attaches proof: tests, changed files, and handoff summary.
6. CI verifies the Agent Work Contract and comments on the PR.
7. The replay bundle becomes the audit trail.

## Positioning

MCP gives agents tools. A2A gives agents interoperability. LangGraph gives agents execution graphs.

Guild gives agents the work contract.

## CTA

Install the alpha:

```bash
go install github.com/lucid-fdn/guild/cli/cmd/guild@v0.1.0-alpha.4
guild agentdesk init
guild agentdesk doctor
```

Connect an MCP host:

```bash
guild mcp serve
```
