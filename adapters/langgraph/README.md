# LangGraph Adapter

[![Guild adapter-alpha](../../conformance/badges/guild-adapter-alpha.svg)](../../conformance/profiles/langgraph.v1alpha1.json)

Package: `@guild/adapter-langgraph`

This is the first real orchestrator adapter. It gives LangGraph graphs a node-shaped bridge that persists institutional records into Guild without making Guild the orchestrator.

The adapter exports:

- `buildLangGraphInstitutionalRun`
- `createGuildLangGraphNode`

Usage sketch:

```ts
import { StateGraph } from "@langchain/langgraph";
import { GuildClient } from "@guild/client";
import { createGuildLangGraphNode } from "@guild/adapter-langgraph";

const guildNode = createGuildLangGraphNode({
  client: new GuildClient("http://localhost:8080"),
  task: (state) => ({
    taskpackId: state.taskpackId,
    driBindingId: state.driBindingId,
    title: state.title,
    objective: state.objective,
    requestedBy: state.requestedBy,
    owner: state.owner
  }),
  artifacts: (state) => state.artifacts
});

const graph = new StateGraph(/* your state schema */)
  .addNode("guild_record", guildNode);
```

Run checks:

```bash
pnpm --dir adapters/langgraph run check
```

Guild remains orchestrator-agnostic: LangGraph executes the graph; Guild records ownership, artifacts, replay, and institutional memory.
