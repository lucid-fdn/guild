import type { Artifact, DriBinding, GuildStatus, PromotionRecord, ReplayBundle, Taskpack } from "@guild/client";

export type GuildExperience = {
  status: GuildStatus | null;
  taskpacks: Taskpack[];
  driBindings: DriBinding[];
  artifacts: Artifact[];
  promotionRecords: PromotionRecord[];
  governancePolicies: GovernancePolicy[];
  approvalRequests: ApprovalRequest[];
  promotionGates: PromotionGate[];
  commonsEntries: CommonsEntry[];
  source: "api" | "demo";
};

export type GovernancePolicy = {
  policy_id: string;
  institution_id: string;
  name: string;
  description?: string;
  rules: Array<{ rule_id: string; effect: string; condition: string; min_approvals?: number }>;
};

export type ApprovalRequest = {
  approval_id: string;
  taskpack_id: string;
  policy_id?: string;
  reason: string;
  required_approvals: number;
  status: string;
};

export type PromotionGate = {
  gate_id: string;
  institution_id: string;
  name: string;
  candidate_kinds: string[];
  min_replay_runs: number;
  requires_approval: boolean;
};

export type CommonsEntry = {
  commons_entry_id: string;
  institution_id: string;
  promotion_record_id: string;
  title: string;
  summary: string;
  scope: string;
  status: string;
};

const baseUrl = (process.env.GUILD_API_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");

export async function getGuildExperience(): Promise<GuildExperience> {
  const [status, taskpacks, driBindings, artifacts, promotionRecords, governancePolicies, approvalRequests, promotionGates, commonsEntries] = await Promise.all([
    getJson<GuildStatus>("/api/v1/status"),
    getCollection<Taskpack>("/api/v1/taskpacks"),
    getCollection<DriBinding>("/api/v1/dri-bindings"),
    getCollection<Artifact>("/api/v1/artifacts"),
    getCollection<PromotionRecord>("/api/v1/promotion-records"),
    getCollection<GovernancePolicy>("/api/v1/governance-policies"),
    getCollection<ApprovalRequest>("/api/v1/approval-requests"),
    getCollection<PromotionGate>("/api/v1/promotion-gates"),
    getCollection<CommonsEntry>("/api/v1/commons-entries")
  ]);

  if (!status && taskpacks.length === 0 && driBindings.length === 0 && artifacts.length === 0 && promotionRecords.length === 0) {
    return demoExperience;
  }

  return {
    status,
    taskpacks,
    driBindings,
    artifacts,
    promotionRecords,
    governancePolicies,
    approvalRequests,
    promotionGates,
    commonsEntries,
    source: "api"
  };
}

export async function getReplayBundle(taskpackId: string): Promise<ReplayBundle | null> {
  return getJson<ReplayBundle>(`/api/v1/replay/taskpacks/${encodeURIComponent(taskpackId)}`);
}

export function recordsForTask(experience: GuildExperience, taskpackId: string) {
  return {
    taskpack: experience.taskpacks.find((taskpack) => taskpack.taskpack_id === taskpackId) ?? null,
    driBindings: experience.driBindings.filter((binding) => binding.taskpack_id === taskpackId),
    artifacts: experience.artifacts.filter((artifact) => artifact.taskpack_id === taskpackId),
    promotionRecords: experience.promotionRecords.filter((record) =>
      experience.artifacts.some(
        (artifact) =>
          artifact.taskpack_id === taskpackId && artifact.artifact_id === record.candidate_ref.artifact_id
      )
    )
  };
}

async function getCollection<T>(path: string): Promise<T[]> {
  const payload = await getJson<{ items?: T[] }>(path);
  return payload?.items ?? [];
}

async function getJson<T>(path: string): Promise<T | null> {
  try {
    const response = await fetch(`${baseUrl}${path}`, { cache: "no-store" });
    if (!response.ok) {
      return null;
    }
    return (await response.json()) as T;
  } catch {
    return null;
  }
}

const requester = {
  actor_id: "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
  actor_type: "human" as const,
  display_name: "Quentin"
};

const owner = {
  actor_id: "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
  actor_type: "agent" as const,
  display_name: "payments-dri",
  orchestrator: "langgraph",
  model: "anthropic:claude-sonnet-4"
};

const reviewer = {
  actor_id: "4ea6b08c-e80b-477f-b403-f3e4a5a71a2d",
  actor_type: "agent" as const,
  display_name: "skeptic-reviewer",
  orchestrator: "openai-agents",
  model: "openai:gpt-5.4"
};

export const demoExperience: GuildExperience = {
  source: "demo",
  status: {
    name: "guild",
    version: "0.1.0",
    mode: "demo",
    ui_origin: "http://localhost:3000",
    services: [
      { name: "control-plane", status: "offline-demo" },
      { name: "experience-plane", status: "ready" }
    ]
  },
  taskpacks: [
    {
      schema_version: "v1alpha1",
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      institution_id: "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c",
      title: "Audit the payment webhook retry path",
      objective: "Find retry edge cases, propose fixes, and produce an implementation plan.",
      problem_statement: "Recent incidents suggest duplicate webhook handling still leaks side effects during timeout retries.",
      task_type: "analysis",
      priority: "high",
      requested_by: requester,
      role_hint: "dri",
      context_budget: {
        max_input_tokens: 12000,
        max_output_tokens: 2500,
        context_strategy: "artifact_refs_first",
        max_artifacts_inline: 2
      },
      permissions: {
        allow_network: false,
        allow_shell: false,
        allow_external_write: false,
        approval_mode: "ask",
        scopes: ["repo:read"]
      },
      acceptance_criteria: [
        {
          criterion_id: "root-cause",
          description: "Identify at least one likely duplicate side-effect path with file-level evidence.",
          required: true,
          verification_hint: "Code references and artifact links required."
        }
      ],
      created_at: "2026-04-24T10:00:00Z",
      trace_id: "a22fb6f0ff0bd4f8a6ef9173fbe80df4"
    }
  ],
  driBindings: [
    {
      schema_version: "v1alpha1",
      dri_binding_id: "19887415-bb68-438b-9599-0b07b5f13e97",
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      owner,
      reviewers: [reviewer],
      status: "assigned",
      created_at: "2026-04-24T10:01:00Z"
    }
  ],
  artifacts: [
    {
      schema_version: "v1alpha1",
      artifact_id: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      kind: "review",
      title: "Webhook retry audit findings",
      summary: "Identifies duplicate side effects in timeout retry handling and proposes an idempotency guard.",
      producer: owner,
      storage: {
        uri: "s3://guild/artifacts/tasks/4e1fe00c-6303-453c-8cb6-2c34f84896e4/findings.md",
        mime_type: "text/markdown"
      },
      version: 1,
      created_at: "2026-04-24T10:15:00Z"
    }
  ],
  governancePolicies: [
    {
      policy_id: "cd542e3e-41b0-46a0-b1f2-b27af9eb9fe4",
      institution_id: "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c",
      name: "Promotion requires evidence and human approval",
      description: "No learning enters the commons without replay evidence and one accountable human approver.",
      rules: [{ rule_id: "replay-evidence", effect: "require_approval", condition: "candidate_kind in ['skill', 'review_heuristic']", min_approvals: 1 }]
    }
  ],
  approvalRequests: [
    {
      approval_id: "8b63fd74-7e42-4c2a-aed1-d39a628d8988",
      taskpack_id: "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
      policy_id: "cd542e3e-41b0-46a0-b1f2-b27af9eb9fe4",
      reason: "Promote the retry-audit heuristic into the institution commons.",
      required_approvals: 1,
      status: "approved"
    }
  ],
  promotionGates: [
    {
      gate_id: "4fb32cb9-80e2-45de-97ff-7c1781a49149",
      institution_id: "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c",
      name: "Replay improvement gate",
      candidate_kinds: ["skill", "review_heuristic"],
      min_replay_runs: 1,
      requires_approval: true
    }
  ],
  commonsEntries: [
    {
      commons_entry_id: "dcb55b43-d39f-4d49-8b0f-54b4ef68f73e",
      institution_id: "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c",
      promotion_record_id: "b2ddb0dd-b29c-4a28-b1ba-e9a2f8ff23fb",
      title: "Check retries for duplicate side effects",
      summary: "When auditing webhook retry paths, inspect timeout recovery for non-idempotent side effects before acknowledging success.",
      scope: "institution",
      status: "active"
    }
  ],
  promotionRecords: [
    {
      schema_version: "v1alpha1",
      promotion_record_id: "b2ddb0dd-b29c-4a28-b1ba-e9a2f8ff23fb",
      institution_id: "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c",
      candidate_kind: "review_heuristic",
      candidate_ref: {
        artifact_id: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
        kind: "review",
        uri: "s3://guild/artifacts/tasks/4e1fe00c-6303-453c-8cb6-2c34f84896e4/findings.md",
        version: 1
      },
      decision: "accepted",
      decision_reason: "Catch-rate improvement outweighed the small latency cost and no policy regressions were observed.",
      decided_at: "2026-04-24T11:00:00Z"
    }
  ]
};
