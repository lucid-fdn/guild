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
import {
  buildArtifact,
  buildDriBinding,
  buildTaskpack,
  type GuildControlPlaneClient
} from "@guild/adapter-core";

export type McpJsonSchema = {
  type: "object";
  properties: Record<string, unknown>;
  required?: string[];
  additionalProperties: false;
};

export type McpToolDefinition = {
  name: string;
  description: string;
  inputSchema: McpJsonSchema;
};

export type McpToolCall = {
  name: string;
  arguments?: Record<string, unknown>;
};

export type McpToolResult = {
  content: Array<{ type: "text"; text: string }>;
  isError?: boolean;
};

export type GuildMcpClient = GuildControlPlaneClient & {
  exportReplayBundle(taskpackId: string): Promise<ReplayBundle>;
  getNextMandate?(): Promise<Taskpack>;
  claimMandate?(input: ClaimMandateArgs): Promise<unknown>;
  compileContext?(input: CompileContextArgs): Promise<ContextPack>;
  checkPreflight?(input: CheckPreflightArgs): Promise<PreflightDecision>;
  requestApproval?(input: RequestApprovalArgs): Promise<ApprovalRequest>;
  createHandoff?(input: CreateHandoffArgs): Promise<Artifact>;
  verifyMandate?(input: VerifyMandateArgs): Promise<unknown>;
  closeMandate?(input: CloseMandateArgs): Promise<unknown>;
};

type CreateTaskpackArgs = {
  taskpack_id: string;
  title: string;
  objective: string;
  requested_by: ActorRef;
  created_at?: string;
  institution_id?: string;
  priority?: Taskpack["priority"];
  task_type?: Taskpack["task_type"];
  role_hint?: Taskpack["role_hint"];
  labels?: string[];
};

type AssignDriArgs = {
  dri_binding_id: string;
  taskpack_id: string;
  owner: ActorRef;
  created_at?: string;
  reviewers?: ActorRef[];
  specialists?: ActorRef[];
  approvers?: ActorRef[];
  status?: DriBinding["status"];
};

type PublishArtifactArgs = {
  artifact_id: string;
  taskpack_id: string;
  title: string;
  producer: ActorRef;
  uri: string;
  created_at?: string;
  kind?: Artifact["kind"];
  summary?: string;
  mime_type?: string;
  labels?: string[];
  parent_artifact_ids?: string[];
  trace_id?: string;
  run_id?: string;
};

type ExportReplayArgs = {
  taskpack_id: string;
};

type GetNextMandateArgs = {
  role?: string;
};

export type ClaimMandateArgs = {
  taskpack_id: string;
  agent?: string;
  ttl_minutes?: number;
  force?: boolean;
};

export type CompileContextArgs = {
  taskpack_id: string;
  role: string;
  budget_tokens?: number;
};

export type CheckPreflightArgs = {
  taskpack_id: string;
  action: PreflightDecision["action"];
  path?: string;
  command?: string;
};

export type RequestApprovalArgs = {
  taskpack_id: string;
  reason: string;
  required_approvals?: number;
};

export type CreateHandoffArgs = {
  taskpack_id: string;
  to: string;
  summary: string;
};

export type VerifyMandateArgs = {
  taskpack_id: string;
};

export type CloseMandateArgs = {
  taskpack_id: string;
};

export const guildMcpTools: McpToolDefinition[] = [
  {
    name: "guild_get_next_mandate",
    description: "Return the next open mandate an agent can claim from the workspace or compatible API.",
    inputSchema: objectSchema(
      {
        role: { type: "string" }
      },
      []
    )
  },
  {
    name: "guild_claim_mandate",
    description: "Create a local lease so multiple agents do not pick the same mandate.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" },
        agent: { type: "string" },
        ttl_minutes: { type: "integer", minimum: 1 },
        force: { type: "boolean" }
      },
      ["taskpack_id"]
    )
  },
  {
    name: "guild_create_taskpack",
    description: "Create a bounded Guild Taskpack/mandate from an MCP host or tool runtime.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" },
        title: { type: "string" },
        objective: { type: "string" },
        requested_by: { type: "object" },
        created_at: { type: "string", format: "date-time" },
        institution_id: { type: "string", format: "uuid" },
        priority: { type: "string", enum: ["low", "medium", "high", "critical"] },
        task_type: {
          type: "string",
          enum: ["analysis", "implementation", "review", "research", "triage", "evaluation", "operations", "custom"]
        },
        role_hint: {
          type: "string",
          enum: ["dri", "explorer", "builder", "skeptic", "reviewer", "specialist", "approver"]
        },
        labels: { type: "array", items: { type: "string" } }
      },
      ["taskpack_id", "title", "objective", "requested_by"]
    )
  },
  {
    name: "guild_assign_dri",
    description: "Assign exactly one accountable DRI owner to a Guild Taskpack.",
    inputSchema: objectSchema(
      {
        dri_binding_id: { type: "string", format: "uuid" },
        taskpack_id: { type: "string", format: "uuid" },
        owner: { type: "object" },
        created_at: { type: "string", format: "date-time" },
        reviewers: { type: "array", items: { type: "object" } },
        specialists: { type: "array", items: { type: "object" } },
        approvers: { type: "array", items: { type: "object" } },
        status: { type: "string", enum: ["assigned", "accepted", "in_progress", "blocked", "completed", "canceled"] }
      },
      ["dri_binding_id", "taskpack_id", "owner"]
    )
  },
  {
    name: "guild_publish_artifact",
    description: "Publish a durable proof artifact reference from an MCP tool result.",
    inputSchema: objectSchema(
      {
        artifact_id: { type: "string", format: "uuid" },
        taskpack_id: { type: "string", format: "uuid" },
        title: { type: "string" },
        producer: { type: "object" },
        uri: { type: "string" },
        created_at: { type: "string", format: "date-time" },
        kind: {
          type: "string",
          enum: [
            "report",
            "code_patch",
            "review",
            "plan",
            "dataset",
            "decision_log",
            "benchmark_result",
            "skill_candidate",
            "test_report",
            "diff",
            "changed_files",
            "screenshot",
            "log_excerpt",
            "security_review",
            "handoff_summary",
            "human_approval",
            "custom"
          ]
        },
        summary: { type: "string" },
        mime_type: { type: "string" },
        labels: { type: "array", items: { type: "string" } },
        parent_artifact_ids: { type: "array", items: { type: "string", format: "uuid" } },
        trace_id: { type: "string" },
        run_id: { type: "string", format: "uuid" }
      },
      ["artifact_id", "taskpack_id", "title", "producer", "uri"]
    )
  },
  {
    name: "guild_compile_context",
    description: "Compile a role-specific bounded context pack for one mandate.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" },
        role: { type: "string" },
        budget_tokens: { type: "integer", minimum: 256 }
      },
      ["taskpack_id", "role"]
    )
  },
  {
    name: "guild_check_preflight",
    description: "Check whether an agent action is allowed, denied, or needs approval before execution.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" },
        action: {
          type: "string",
          enum: ["read", "write", "run", "network", "secret", "git_push", "dependency_install", "prod_access", "custom"]
        },
        path: { type: "string" },
        command: { type: "string" }
      },
      ["taskpack_id", "action"]
    )
  },
  {
    name: "guild_request_approval",
    description: "Request human or reviewer approval when a mandate action crosses policy boundaries.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" },
        reason: { type: "string" },
        required_approvals: { type: "integer", minimum: 1 }
      },
      ["taskpack_id", "reason"]
    )
  },
  {
    name: "guild_create_handoff",
    description: "Create a structured handoff proof artifact for another agent or reviewer.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" },
        to: { type: "string" },
        summary: { type: "string" }
      },
      ["taskpack_id", "to", "summary"]
    )
  },
  {
    name: "guild_verify_mandate",
    description: "Verify proof, approvals, and handoff readiness before an agent claims completion.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" }
      },
      ["taskpack_id"]
    )
  },
  {
    name: "guild_close_mandate",
    description: "Close a mandate after required proof artifacts have been published.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" }
      },
      ["taskpack_id"]
    )
  },
  {
    name: "guild_export_replay_bundle",
    description: "Export a portable replay/proof bundle for a mandate and its records.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" }
      },
      ["taskpack_id"]
    )
  }
];

export class GuildMcpBridge {
  readonly tools = guildMcpTools;

  constructor(private readonly client: GuildMcpClient) {}

  async handle(call: McpToolCall): Promise<McpToolResult> {
    try {
      switch (call.name) {
        case "guild_get_next_mandate":
          return jsonResult(await this.getNextMandate(readArgs<GetNextMandateArgs>(call)));
        case "guild_claim_mandate":
          return jsonResult(await this.claimMandate(readArgs<ClaimMandateArgs>(call)));
        case "guild_create_taskpack":
          return jsonResult(await this.createTaskpack(readArgs<CreateTaskpackArgs>(call)));
        case "guild_assign_dri":
          return jsonResult(await this.assignDri(readArgs<AssignDriArgs>(call)));
        case "guild_publish_artifact":
          return jsonResult(await this.publishArtifact(readArgs<PublishArtifactArgs>(call)));
        case "guild_compile_context":
          return jsonResult(await this.compileContext(readArgs<CompileContextArgs>(call)));
        case "guild_check_preflight":
          return jsonResult(await this.checkPreflight(readArgs<CheckPreflightArgs>(call)));
        case "guild_request_approval":
          return jsonResult(await this.requestApproval(readArgs<RequestApprovalArgs>(call)));
        case "guild_create_handoff":
          return jsonResult(await this.createHandoff(readArgs<CreateHandoffArgs>(call)));
        case "guild_verify_mandate":
          return jsonResult(await this.verifyMandate(readArgs<VerifyMandateArgs>(call)));
        case "guild_close_mandate":
          return jsonResult(await this.closeMandate(readArgs<CloseMandateArgs>(call)));
        case "guild_export_replay_bundle":
          return jsonResult(await this.exportReplay(readArgs<ExportReplayArgs>(call)));
        default:
          return errorResult(`unknown Guild MCP tool: ${call.name}`);
      }
    } catch (error) {
      return errorResult(error instanceof Error ? error.message : "Guild MCP tool failed");
    }
  }

  private async getNextMandate(_input: GetNextMandateArgs): Promise<Taskpack> {
    if (!this.client.getNextMandate) {
      throw new Error("guild_get_next_mandate requires a client with getNextMandate()");
    }
    return this.client.getNextMandate();
  }

  private async claimMandate(input: ClaimMandateArgs): Promise<unknown> {
    if (!this.client.claimMandate) {
      throw new Error("guild_claim_mandate requires a client with claimMandate()");
    }
    return this.client.claimMandate(input);
  }

  private async createTaskpack(input: CreateTaskpackArgs): Promise<Taskpack> {
    return this.client.createTaskpack(
      buildTaskpack({
        taskpackId: input.taskpack_id,
        title: input.title,
        objective: input.objective,
        requestedBy: input.requested_by,
        createdAt: input.created_at ?? new Date().toISOString(),
        institutionId: input.institution_id,
        priority: input.priority,
        taskType: input.task_type,
        roleHint: input.role_hint,
        labels: input.labels
      })
    );
  }

  private async assignDri(input: AssignDriArgs): Promise<DriBinding> {
    return this.client.createDriBinding(
      buildDriBinding({
        driBindingId: input.dri_binding_id,
        taskpackId: input.taskpack_id,
        owner: input.owner,
        reviewers: input.reviewers,
        specialists: input.specialists,
        approvers: input.approvers,
        status: input.status,
        createdAt: input.created_at ?? new Date().toISOString()
      })
    );
  }

  private async publishArtifact(input: PublishArtifactArgs): Promise<Artifact> {
    return this.client.createArtifact(
      buildArtifact({
        artifactId: input.artifact_id,
        taskpackId: input.taskpack_id,
        title: input.title,
        producer: input.producer,
        uri: input.uri,
        createdAt: input.created_at ?? new Date().toISOString(),
        kind: input.kind,
        summary: input.summary,
        mimeType: input.mime_type,
        labels: input.labels,
        parentArtifactIds: input.parent_artifact_ids,
        traceId: input.trace_id,
        runId: input.run_id
      })
    );
  }

  private async exportReplay(input: ExportReplayArgs): Promise<ReplayBundle> {
    return this.client.exportReplayBundle(input.taskpack_id);
  }

  private async compileContext(input: CompileContextArgs): Promise<ContextPack> {
    if (!this.client.compileContext) {
      throw new Error("guild_compile_context requires a client with compileContext()");
    }
    return this.client.compileContext(input);
  }

  private async checkPreflight(input: CheckPreflightArgs): Promise<PreflightDecision> {
    if (!this.client.checkPreflight) {
      throw new Error("guild_check_preflight requires a client with checkPreflight()");
    }
    return this.client.checkPreflight(input);
  }

  private async requestApproval(input: RequestApprovalArgs): Promise<ApprovalRequest> {
    if (!this.client.requestApproval) {
      throw new Error("guild_request_approval requires a client with requestApproval()");
    }
    return this.client.requestApproval(input);
  }

  private async createHandoff(input: CreateHandoffArgs): Promise<Artifact> {
    if (!this.client.createHandoff) {
      throw new Error("guild_create_handoff requires a client with createHandoff()");
    }
    return this.client.createHandoff(input);
  }

  private async verifyMandate(input: VerifyMandateArgs): Promise<unknown> {
    if (!this.client.verifyMandate) {
      throw new Error("guild_verify_mandate requires a client with verifyMandate()");
    }
    return this.client.verifyMandate(input);
  }

  private async closeMandate(input: CloseMandateArgs): Promise<unknown> {
    if (!this.client.closeMandate) {
      throw new Error("guild_close_mandate requires a client with closeMandate()");
    }
    return this.client.closeMandate(input);
  }
}

export function createGuildMcpBridge(client: GuildMcpClient): GuildMcpBridge {
  return new GuildMcpBridge(client);
}

function objectSchema(properties: Record<string, unknown>, required?: string[]): McpJsonSchema {
  return {
    type: "object",
    properties,
    required,
    additionalProperties: false
  };
}

function readArgs<T>(call: McpToolCall): T {
  if (!call.arguments || typeof call.arguments !== "object") {
    throw new Error(`${call.name} requires arguments`);
  }
  return call.arguments as T;
}

function jsonResult(payload: unknown): McpToolResult {
  return {
    content: [
      {
        type: "text",
        text: JSON.stringify(payload, null, 2)
      }
    ]
  };
}

function errorResult(message: string): McpToolResult {
  return {
    isError: true,
    content: [
      {
        type: "text",
        text: message
      }
    ]
  };
}
