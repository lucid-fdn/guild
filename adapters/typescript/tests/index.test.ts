import assert from "node:assert/strict";
import { test } from "node:test";
import type { ActorRef, Artifact, DriBinding, Taskpack } from "@guild/client";
import { buildArtifact, buildDriBinding, buildTaskpack, submitInstitutionalRun } from "../src/index";

const createdAt = "2026-04-24T12:00:00Z";
const human: ActorRef = {
  actor_id: "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
  actor_type: "human",
  display_name: "Operator"
};
const owner: ActorRef = {
  actor_id: "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
  actor_type: "agent",
  display_name: "dri-agent",
  orchestrator: "custom"
};

test("buildTaskpack emits a bounded DRI-ready default taskpack", () => {
  const taskpack = buildTaskpack({
    taskpackId: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    title: "Review retry logic",
    objective: "Find one retry risk.",
    requestedBy: human,
    createdAt
  });

  assert.equal(taskpack.schema_version, "v1alpha1");
  assert.equal(taskpack.task_type, "analysis");
  assert.equal(taskpack.priority, "medium");
  assert.equal(taskpack.context_budget.context_strategy, "artifact_refs_first");
  assert.equal(taskpack.context_budget.max_input_tokens, 4000);
  assert.equal(taskpack.permissions.approval_mode, "ask");
  assert.equal(taskpack.acceptance_criteria[0].criterion_id, "deliverable");
});

test("buildDriBinding preserves the single accountable owner", () => {
  const binding = buildDriBinding({
    driBindingId: "19887415-bb68-438b-9599-0b07b5f13e97",
    taskpackId: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    owner,
    createdAt
  });

  assert.equal(binding.schema_version, "v1alpha1");
  assert.equal(binding.status, "assigned");
  assert.equal(binding.owner.actor_id, owner.actor_id);
});

test("buildArtifact emits provenance and storage defaults", () => {
  const artifact = buildArtifact({
    artifactId: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
    taskpackId: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    title: "Retry review",
    producer: owner,
    uri: "s3://guild/example/retry-review.md",
    createdAt,
    traceId: "trace-1",
    runId: "9644f432-eeb5-443f-b7ea-f0aab64ad679"
  });

  assert.equal(artifact.schema_version, "v1alpha1");
  assert.equal(artifact.kind, "report");
  assert.equal(artifact.storage.mime_type, "text/markdown");
  assert.deepEqual(artifact.lineage?.source_taskpack_ids, ["4e1fe00c-6303-453c-8cb6-2c34f84896e4"]);
  assert.equal(artifact.version, 1);
});

test("submitInstitutionalRun submits taskpack, DRI, then artifacts", async () => {
  const calls: string[] = [];
  const taskpack = buildTaskpack({
    taskpackId: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    title: "Review retry logic",
    objective: "Find one retry risk.",
    requestedBy: human,
    createdAt
  });
  const binding = buildDriBinding({
    driBindingId: "19887415-bb68-438b-9599-0b07b5f13e97",
    taskpackId: taskpack.taskpack_id,
    owner,
    createdAt
  });
  const artifact = buildArtifact({
    artifactId: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
    taskpackId: taskpack.taskpack_id,
    title: "Retry review",
    producer: owner,
    uri: "s3://guild/example/retry-review.md",
    createdAt
  });

  await submitInstitutionalRun(
    {
      async createTaskpack(payload: Taskpack) {
        calls.push(`taskpack:${payload.taskpack_id}`);
        return payload;
      },
      async createDriBinding(payload: DriBinding) {
        calls.push(`dri:${payload.taskpack_id}`);
        return payload;
      },
      async createArtifact(payload: Artifact) {
        calls.push(`artifact:${payload.taskpack_id}`);
        return payload;
      }
    },
    {
      taskpack,
      driBinding: binding,
      artifacts: [artifact]
    }
  );

  assert.deepEqual(calls, [
    "taskpack:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "dri:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "artifact:4e1fe00c-6303-453c-8cb6-2c34f84896e4"
  ]);
});
