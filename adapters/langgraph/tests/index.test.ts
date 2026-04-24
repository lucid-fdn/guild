import assert from "node:assert/strict";
import { test } from "node:test";
import type { ActorRef, Artifact, DriBinding, Taskpack } from "@guild/client";
import { buildLangGraphInstitutionalRun, createGuildLangGraphNode, type LangGraphGuildState } from "../src/index";

const requestedBy: ActorRef = {
  actor_id: "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
  actor_type: "human",
  display_name: "Operator"
};

const owner: ActorRef = {
  actor_id: "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
  actor_type: "agent",
  display_name: "langgraph-dri"
};

test("buildLangGraphInstitutionalRun creates LangGraph-labeled Guild records", () => {
  const run = buildLangGraphInstitutionalRun(
    {
      taskpackId: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      driBindingId: "19887415-bb68-438b-9599-0b07b5f13e97",
      title: "Coordinate LangGraph node",
      objective: "Persist a LangGraph node handoff into Guild.",
      requestedBy,
      owner,
      createdAt: "2026-04-24T12:00:00Z"
    },
    [
      {
        artifactId: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
        title: "LangGraph report",
        uri: "s3://guild/langgraph/report.md",
        summary: "Node result persisted as an artifact."
      }
    ]
  );

  assert.equal(run.taskpack.task_type, "operations");
  assert.equal(run.taskpack.role_hint, "dri");
  assert.deepEqual(run.taskpack.labels, ["orchestrator:langgraph"]);
  assert.equal(run.driBinding.owner.orchestrator, "langgraph");
  assert.equal(run.artifacts[0].labels?.[0], "orchestrator:langgraph");
});

test("createGuildLangGraphNode submits records and returns a LangGraph state patch", async () => {
  const calls: string[] = [];
  const node = createGuildLangGraphNode<LangGraphGuildState>({
    client: {
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
    task: () => ({
      taskpackId: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      driBindingId: "19887415-bb68-438b-9599-0b07b5f13e97",
      title: "Coordinate LangGraph node",
      objective: "Persist a LangGraph node handoff into Guild.",
      requestedBy,
      owner,
      createdAt: "2026-04-24T12:00:00Z"
    }),
    artifacts: [
      {
        artifactId: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
        title: "LangGraph report",
        uri: "s3://guild/langgraph/report.md"
      }
    ]
  });

  const patch = await node({});

  assert.deepEqual(calls, [
    "taskpack:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "dri:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "artifact:4e1fe00c-6303-453c-8cb6-2c34f84896e4"
  ]);
  assert.equal(patch.guild.submitted, true);
  assert.equal(patch.guild.artifacts.length, 1);
});
