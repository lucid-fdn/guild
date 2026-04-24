# Examples

Runnable examples:

- `one-task-one-dri-commons`: the canonical launch story for one bounded task, one DRI, durable artifacts, replay, human approval, promotion gates, and commons learning
- `typescript-adapter-core`: a minimal TypeScript integration shape that submits a Guild institutional run

Planned examples:

- `langgraph-review-loop`
- `openai-agents-dri-task`
- `a2a-remote-specialist`
- `mcp-tool-audit`
- `promotion-gate-demo`

Each example should:

- create a valid `Taskpack`
- register a `DRI Binding`
- emit at least one `Artifact`
- optionally create a `Promotion Record`

Also included:

- `api-demo.http` for quick local endpoint exploration with REST clients like VS Code REST Client or Bruno

The first adapter implementation is `adapters/typescript`, a neutral package
that framework-specific examples can build on without making Guild an
orchestrator.

Run the canonical launch simulation:

```bash
examples/one-task-one-dri-commons/run.sh
```
