# TypeScript Adapter Core Example

This example shows the intended integration shape for orchestrators:

1. Keep your existing runtime.
2. Convert a local task into a `Taskpack`.
3. Convert the accountable owner into a `DRI Binding`.
4. Convert durable outputs into `Artifact`s.
5. Submit everything to Guild.
6. Export a replay bundle.

Run a local Guild server first:

```bash
make run-server
```

Then run the example with your preferred TypeScript runner, or use the source as
the mapping template for a real adapter:

```bash
pnpm --dir examples/typescript-adapter-core run check
GUILD_BASE_URL=http://localhost:8080 pnpm --dir examples/typescript-adapter-core run demo
```

The example intentionally does not depend on LangGraph, CrewAI, OpenAI Agents,
or any other orchestrator. Specific adapters should be thin wrappers around this
shape.
