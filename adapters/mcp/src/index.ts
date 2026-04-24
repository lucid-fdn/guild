import type { ActorRef, Artifact, DriBinding, ReplayBundle, Taskpack } from "@guild/client";
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

export const guildMcpTools: McpToolDefinition[] = [
  {
    name: "guild_create_taskpack",
    description: "Create a bounded Guild Taskpack from an MCP host or tool runtime.",
    inputSchema: objectSchema(
      {
        taskpack_id: { type: "string", format: "uuid" },
        title: { type: "string" },
        objective: { type: "string" },
        requested_by: { type: "object" },
        created_at: { type: "string", format: "date-time" },
        institution_id: { type: "string", format: "uuid" },
        priority: { type: "string", enum: ["low", "medium", "high", "urgent"] },
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
    description: "Publish a durable artifact reference from an MCP tool result.",
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
          enum: ["report", "code_patch", "review", "plan", "dataset", "decision_log", "benchmark_result", "skill_candidate", "custom"]
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
    name: "guild_export_replay_bundle",
    description: "Export a portable replay bundle for a Taskpack and its institutional records.",
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
        case "guild_create_taskpack":
          return jsonResult(await this.createTaskpack(readArgs<CreateTaskpackArgs>(call)));
        case "guild_assign_dri":
          return jsonResult(await this.assignDri(readArgs<AssignDriArgs>(call)));
        case "guild_publish_artifact":
          return jsonResult(await this.publishArtifact(readArgs<PublishArtifactArgs>(call)));
        case "guild_export_replay_bundle":
          return jsonResult(await this.exportReplay(readArgs<ExportReplayArgs>(call)));
        default:
          return errorResult(`unknown Guild MCP tool: ${call.name}`);
      }
    } catch (error) {
      return errorResult(error instanceof Error ? error.message : "Guild MCP tool failed");
    }
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
