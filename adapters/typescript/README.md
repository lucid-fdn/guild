# TypeScript Adapter Core

`@guild/adapter-core` is the tiny neutral layer that framework-specific adapters
should build on.

It is not an orchestrator. It does three things:

- builds valid Guild objects from orchestrator-local concepts
- preserves DRI ownership as a first-class record
- submits taskpacks, ownership, and artifacts to the Guild control plane

Framework adapters should map their own runtime concepts into this package:

- LangGraph node/run -> `Taskpack`
- OpenAI Agents handoff -> `Taskpack`
- CrewAI task owner -> `DRI Binding`
- tool output or final answer -> `Artifact`

## Example

```ts
import { GuildClient } from "@guild/client";
import { buildArtifact, buildDriBinding, buildTaskpack, submitInstitutionalRun } from "@guild/adapter-core";

const client = new GuildClient("http://localhost:8080");
const now = new Date().toISOString();

const taskpack = buildTaskpack({
  taskpackId: crypto.randomUUID(),
  title: "Review checkout retry logic",
  objective: "Find one duplicate-charge risk and produce a mitigation plan.",
  requestedBy: {
    actor_id: crypto.randomUUID(),
    actor_type: "human",
    display_name: "Operator"
  },
  createdAt: now
});

const owner = {
  actor_id: crypto.randomUUID(),
  actor_type: "agent",
  display_name: "checkout-dri",
  orchestrator: "custom"
} as const;

await submitInstitutionalRun(client, {
  taskpack,
  driBinding: buildDriBinding({
    driBindingId: crypto.randomUUID(),
    taskpackId: taskpack.taskpack_id,
    owner,
    createdAt: now
  }),
  artifacts: [
    buildArtifact({
      artifactId: crypto.randomUUID(),
      taskpackId: taskpack.taskpack_id,
      title: "Retry-risk mitigation plan",
      producer: owner,
      uri: "s3://guild/example/retry-plan.md",
      createdAt: now
    })
  ]
});
```

## Verification

```bash
pnpm --dir adapters/typescript run check
```

The check runs TypeScript type-checking plus behavior tests for the object
builders and submission ordering.
