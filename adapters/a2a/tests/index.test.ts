import assert from "node:assert/strict";
import { test } from "node:test";
import type { ActorRef, Artifact, DriBinding, Taskpack } from "@guild/client";
import {
  actorFromAgentCard,
  buildArtifactFromA2AResult,
  buildDriBindingFromA2ATask,
  buildTaskpackFromA2ATask,
  submitA2ATask,
  toA2AArtifactMessage,
  type A2AAgentCard,
  type A2ATaskEnvelope
} from "../src/index";

const requester: ActorRef = {
  actor_id: "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
  actor_type: "human",
  display_name: "Operator"
};

const dri: A2AAgentCard = {
  id: "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
  name: "a2a-dri-agent",
  url: "https://agents.example/dri",
  skills: ["review", "integration"]
};

const task: A2ATaskEnvelope = {
  taskId: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
  title: "Coordinate external agent",
  objective: "Capture the A2A task as a Guild institution record.",
  requester,
  dri,
  createdAt: "2026-04-24T12:00:00Z",
  contextUri: "s3://guild/context.md"
};

test("maps A2A agent cards to Guild actor references", () => {
  const actor = actorFromAgentCard(dri);
  assert.equal(actor.actor_id, dri.id);
  assert.equal(actor.display_name, dri.name);
  assert.equal(actor.orchestrator, dri.url);
});

test("builds Taskpack and DRI binding from A2A task envelopes", () => {
  const taskpack = buildTaskpackFromA2ATask(task);
  const binding = buildDriBindingFromA2ATask(task, "19887415-bb68-438b-9599-0b07b5f13e97");

  assert.equal(taskpack.schema_version, "v1alpha1");
  assert.equal(taskpack.taskpack_id, task.taskId);
  assert.equal(taskpack.task_type, "operations");
  assert.equal(taskpack.role_hint, "dri");
  assert.deepEqual(taskpack.labels, ["skill:review", "skill:integration"]);
  assert.equal(binding.owner.actor_id, dri.id);
  assert.equal(binding.status, "assigned");
});

test("builds artifacts and A2A artifact reference messages", () => {
  const artifact = buildArtifactFromA2AResult({
    artifactId: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
    taskId: task.taskId,
    title: "External agent report",
    producer: dri,
    uri: "s3://guild/a2a/report.md",
    createdAt: "2026-04-24T12:05:00Z",
    summary: "Captured external work."
  });
  const message = toA2AArtifactMessage(artifact);

  assert.equal(artifact.producer.actor_id, dri.id);
  assert.equal(artifact.storage.uri, "s3://guild/a2a/report.md");
  assert.equal(message.parts[0].kind, "guild_artifact_ref");
  assert.equal(message.parts[0].artifact_id, artifact.artifact_id);
});

test("submits an A2A task through the neutral Guild control-plane interface", async () => {
  const calls: string[] = [];

  await submitA2ATask(
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
    task,
    {
      driBindingId: "19887415-bb68-438b-9599-0b07b5f13e97",
      results: [
        {
          artifactId: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
          taskId: task.taskId,
          title: "External agent report",
          producer: dri,
          uri: "s3://guild/a2a/report.md",
          createdAt: "2026-04-24T12:05:00Z"
        }
      ]
    }
  );

  assert.deepEqual(calls, [
    "taskpack:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "dri:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "artifact:4e1fe00c-6303-453c-8cb6-2c34f84896e4"
  ]);
});
