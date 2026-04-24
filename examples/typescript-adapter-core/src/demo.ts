import { buildArtifact, buildDriBinding, buildTaskpack, submitInstitutionalRun } from "@guild/adapter-core";
import { GuildClient } from "@guild/client";

declare const process: {
  env: {
    GUILD_BASE_URL?: string;
  };
};

const baseUrl = process.env.GUILD_BASE_URL ?? "http://localhost:8080";
const createdAt = "2026-04-24T12:30:00Z";

const humanRequester = {
  actor_id: "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
  actor_type: "human",
  display_name: "Operator"
} as const;

const driAgent = {
  actor_id: "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
  actor_type: "agent",
  display_name: "reference-dri",
  orchestrator: "custom-reference"
} as const;

const taskpack = buildTaskpack({
  taskpackId: "d013e9c3-3fdc-4f72-a79f-3ca30d0fe111",
  title: "Review checkout retry logic",
  objective: "Find one duplicate-charge risk and produce a mitigation plan.",
  requestedBy: humanRequester,
  createdAt,
  taskType: "review",
  priority: "high",
  roleHint: "dri",
  acceptanceCriterionId: "risk-plan",
  acceptanceDescription: "Produce one mitigation plan artifact."
});

const driBinding = buildDriBinding({
  driBindingId: "9eb2d8f5-f756-402c-9872-6652f2418f53",
  taskpackId: taskpack.taskpack_id,
  owner: driAgent,
  createdAt,
  status: "accepted"
});

const artifact = buildArtifact({
  artifactId: "c4ce7f7b-6d4b-49c3-a6a4-632ce4317a9c",
  taskpackId: taskpack.taskpack_id,
  title: "Checkout retry mitigation plan",
  summary: "Make retry idempotency explicit before charging.",
  producer: driAgent,
  uri: "s3://guild/examples/checkout-retry-plan.md",
  createdAt,
  kind: "plan",
  labels: ["example", "adapter-core"]
});

async function main(): Promise<void> {
  const client = new GuildClient(baseUrl);
  await submitInstitutionalRun(client, {
    taskpack,
    driBinding,
    artifacts: [artifact]
  });
  const replay = await client.exportReplayBundle(taskpack.taskpack_id);
  console.log(JSON.stringify(replay, null, 2));
}

void main();
