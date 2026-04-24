# Guild

## The institution layer for AI teams

Agents do not fail only because they lack intelligence.
They fail because their teams lack institutions.

Guild adds the missing operating system above orchestrators:

- one accountable DRI per task
- bounded handoffs through Taskpacks
- durable artifacts instead of chat history
- human approval for risky decisions
- replay suites for evaluation
- promotion gates before institutional learning
- a commons registry for proven team knowledge

Bring your own orchestrator.
Guild works above LangGraph, MCP hosts, A2A transports, custom runtimes, and future agent frameworks.

## Why now

The next leap in agents is not just bigger models.
It is social intelligence: ownership, review, governance, memory, and learning that survives one run.

Human teams scale through institutions.
AI teams will too.

## The demo

1. A user opens a payment webhook audit.
2. Guild creates one Taskpack and assigns one DRI.
3. The DRI produces a review artifact.
4. Guild exports a recursive replay bundle.
5. An evaluator runs a benchmark suite.
6. A promotion candidate is created.
7. A human approval request gates the learning.
8. The approved pattern enters the commons.

## Positioning

MCP gives agents tools.
A2A gives agents interoperability.
LangGraph gives agents execution graphs.

Guild gives agents institutions.

## CTA

Run the full simulation:

```bash
make simulation
```

Run the release gate:

```bash
make release-check
```
