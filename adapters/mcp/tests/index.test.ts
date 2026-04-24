import assert from "node:assert/strict";
import { test } from "node:test";
import type {
  ActorRef,
  ApprovalRequest,
  Artifact,
  ContextPack,
  DriBinding,
  PreflightDecision,
  ReplayBundle,
  Taskpack
} from "@guild/client";
import { createGuildMcpBridge, guildMcpTools } from "../src/index";

const human: ActorRef = {
  actor_id: "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
  actor_type: "human",
  display_name: "Operator"
};

const owner: ActorRef = {
  actor_id: "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
  actor_type: "agent",
  display_name: "dri-agent",
  orchestrator: "mcp-host"
};

test("exports MCP tool definitions with closed input schemas", () => {
  assert.deepEqual(
    guildMcpTools.map((tool) => tool.name),
    [
      "guild_get_next_mandate",
      "guild_claim_mandate",
      "guild_create_taskpack",
      "guild_assign_dri",
      "guild_publish_artifact",
      "guild_compile_context",
      "guild_check_preflight",
      "guild_request_approval",
      "guild_create_handoff",
      "guild_verify_mandate",
      "guild_close_mandate",
      "guild_export_replay_bundle"
    ]
  );
  assert.equal(guildMcpTools.every((tool) => tool.inputSchema.additionalProperties === false), true);
});

test("bridge creates taskpacks, DRI bindings, artifacts, and replay bundles", async () => {
  const calls: string[] = [];
  const bridge = createGuildMcpBridge({
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
    },
    async getNextMandate() {
      calls.push("next");
      return taskpackFixture("4e1fe00c-6303-453c-8cb6-2c34f84896e4");
    },
    async compileContext(input): Promise<ContextPack> {
      calls.push(`context:${input.taskpack_id}`);
      return {
        schema_version: "v1alpha1",
        mandate_id: input.taskpack_id,
        role: input.role,
        budget_tokens: input.budget_tokens ?? 12000,
        summary: "Bounded context for MCP agent."
      };
    },
    async checkPreflight(input): Promise<PreflightDecision> {
      calls.push(`preflight:${input.taskpack_id}`);
      return {
        schema_version: "v1alpha1",
        mandate_id: input.taskpack_id,
        action: input.action,
        path: input.path,
        command: input.command,
        decision: "allow",
        reason: "Path is within scope.",
        approval_required: false,
        matched_rules: ["scope.writable"]
      };
    },
    async requestApproval(input): Promise<ApprovalRequest> {
      calls.push(`approval:${input.taskpack_id}`);
      return {
        schema_version: "v1alpha1",
        approval_id: "34a86927-5341-46c6-93a2-6b9da928dce6",
        taskpack_id: input.taskpack_id,
        requested_by: owner,
        reason: input.reason,
        required_approvals: input.required_approvals ?? 1,
        status: "pending",
        created_at: "2026-04-24T12:03:00Z"
      };
    },
    async createHandoff(input): Promise<Artifact> {
      calls.push(`handoff:${input.taskpack_id}`);
      return {
        schema_version: "v1alpha1",
        artifact_id: "7b2076da-1c11-40f1-a56c-5058ec366499",
        taskpack_id: input.taskpack_id,
        kind: "handoff_summary",
        title: "Handoff",
        summary: input.summary,
        producer: owner,
        storage: {
          uri: "memory://guild/handoff.md",
          mime_type: "text/markdown"
        },
        version: 1,
        created_at: "2026-04-24T12:04:00Z"
      };
    },
    async verifyMandate(input) {
      calls.push(`verify:${input.taskpack_id}`);
      return { taskpack_id: input.taskpack_id, ready: true };
    },
    async closeMandate(input) {
      calls.push(`close:${input.taskpack_id}`);
      return { taskpack_id: input.taskpack_id, status: "closed" };
    },
    async exportReplayBundle(taskpackId: string): Promise<ReplayBundle> {
      calls.push(`replay:${taskpackId}`);
      return {
        schema_version: "v1alpha1",
        taskpack: taskpackFixture(taskpackId),
        dri_bindings: [],
        artifacts: [],
        promotion_records: []
      };
    }
  });

  const taskResult = await bridge.handle({
    name: "guild_create_taskpack",
    arguments: {
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      title: "Review MCP bridge",
      objective: "Produce a bounded bridge.",
      requested_by: human,
      created_at: "2026-04-24T12:00:00Z"
    }
  });
  const nextResult = await bridge.handle({
    name: "guild_get_next_mandate",
    arguments: {}
  });
  const driResult = await bridge.handle({
    name: "guild_assign_dri",
    arguments: {
      dri_binding_id: "19887415-bb68-438b-9599-0b07b5f13e97",
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      owner,
      created_at: "2026-04-24T12:01:00Z"
    }
  });
  const artifactResult = await bridge.handle({
    name: "guild_publish_artifact",
    arguments: {
      artifact_id: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      title: "Bridge report",
      producer: owner,
      uri: "memory://guild/mcp/report.md",
      created_at: "2026-04-24T12:02:00Z"
    }
  });
  const contextResult = await bridge.handle({
    name: "guild_compile_context",
    arguments: {
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      role: "coder",
      budget_tokens: 12000
    }
  });
  const preflightResult = await bridge.handle({
    name: "guild_check_preflight",
    arguments: {
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      action: "write",
      path: "src/auth/login.ts"
    }
  });
  const approvalResult = await bridge.handle({
    name: "guild_request_approval",
    arguments: {
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      reason: "Needs auth owner consent"
    }
  });
  const handoffResult = await bridge.handle({
    name: "guild_create_handoff",
    arguments: {
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      to: "reviewer",
      summary: "Ready for review"
    }
  });
  const verifyResult = await bridge.handle({
    name: "guild_verify_mandate",
    arguments: {
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4"
    }
  });
  const closeResult = await bridge.handle({
    name: "guild_close_mandate",
    arguments: {
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4"
    }
  });
  const replayResult = await bridge.handle({
    name: "guild_export_replay_bundle",
    arguments: {
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4"
    }
  });

  assert.equal(taskResult.isError, undefined);
  assert.equal(nextResult.isError, undefined);
  assert.equal(driResult.isError, undefined);
  assert.equal(artifactResult.isError, undefined);
  assert.equal(contextResult.isError, undefined);
  assert.equal(preflightResult.isError, undefined);
  assert.equal(approvalResult.isError, undefined);
  assert.equal(handoffResult.isError, undefined);
  assert.equal(verifyResult.isError, undefined);
  assert.equal(closeResult.isError, undefined);
  assert.equal(replayResult.isError, undefined);
  assert.deepEqual(calls, [
    "taskpack:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "next",
    "dri:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "artifact:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "context:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "preflight:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "approval:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "handoff:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "verify:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "close:4e1fe00c-6303-453c-8cb6-2c34f84896e4",
    "replay:4e1fe00c-6303-453c-8cb6-2c34f84896e4"
  ]);
});

function taskpackFixture(taskpackId: string): Taskpack {
  return {
    schema_version: "v1alpha1",
    taskpack_id: taskpackId,
    title: "Review MCP bridge",
    objective: "Produce a bounded bridge.",
    task_type: "analysis",
    priority: "medium",
    requested_by: human,
    context_budget: {
      max_input_tokens: 4000,
      max_output_tokens: 1500,
      context_strategy: "artifact_refs_first"
    },
    permissions: {
      allow_network: false,
      allow_shell: false,
      allow_external_write: false,
      approval_mode: "ask"
    },
    acceptance_criteria: [
      {
        criterion_id: "deliverable",
        description: "Produce the requested artifact.",
        required: true
      }
    ],
    created_at: "2026-04-24T12:00:00Z"
  };
}

test("bridge reports unknown tools without throwing", async () => {
  const bridge = createGuildMcpBridge({
    async createTaskpack(payload: Taskpack) {
      return payload;
    },
    async createDriBinding(payload: DriBinding) {
      return payload;
    },
    async createArtifact(payload: Artifact) {
      return payload;
    },
    async exportReplayBundle(): Promise<ReplayBundle> {
      throw new Error("not used");
    }
  });

  const result = await bridge.handle({ name: "guild_nope" });
  assert.equal(result.isError, true);
  assert.match(result.content[0].text, /unknown Guild MCP tool/);
});
