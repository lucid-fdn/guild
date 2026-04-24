import type { Artifact, ActorRef, DriBinding, Taskpack } from "@guild/client";

type TaskpackInput = {
  taskpackId: string;
  title: string;
  objective: string;
  requestedBy: ActorRef;
  createdAt: string;
  institutionId?: string;
  parentTaskpackId?: string;
  taskType?: Taskpack["task_type"];
  priority?: Taskpack["priority"];
  roleHint?: Taskpack["role_hint"];
  labels?: string[];
  maxInputTokens?: number;
  maxOutputTokens?: number;
  allowNetwork?: boolean;
  allowShell?: boolean;
  allowExternalWrite?: boolean;
  approvalMode?: Taskpack["permissions"]["approval_mode"];
  acceptanceCriterionId?: string;
  acceptanceDescription?: string;
};

type DriBindingInput = {
  driBindingId: string;
  taskpackId: string;
  owner: ActorRef;
  createdAt: string;
  reviewers?: ActorRef[];
  specialists?: ActorRef[];
  approvers?: ActorRef[];
  status?: DriBinding["status"];
};

type ArtifactInput = {
  artifactId: string;
  taskpackId: string;
  title: string;
  producer: ActorRef;
  uri: string;
  createdAt: string;
  kind?: Artifact["kind"];
  summary?: string;
  mimeType?: string;
  version?: number;
  labels?: string[];
  parentArtifactIds?: string[];
  traceId?: string;
  runId?: string;
};

export type InstitutionalRun = {
  taskpack: Taskpack;
  driBinding: DriBinding;
  artifacts?: Artifact[];
};

export type GuildControlPlaneClient = {
  createTaskpack(taskpack: Taskpack): Promise<Taskpack>;
  createDriBinding(binding: DriBinding): Promise<DriBinding>;
  createArtifact(artifact: Artifact): Promise<Artifact>;
};

export function buildTaskpack(input: TaskpackInput): Taskpack {
  return {
    schema_version: "v1alpha1",
    taskpack_id: input.taskpackId,
    institution_id: input.institutionId,
    parent_taskpack_id: input.parentTaskpackId,
    title: input.title,
    objective: input.objective,
    task_type: input.taskType ?? "analysis",
    priority: input.priority ?? "medium",
    requested_by: input.requestedBy,
    role_hint: input.roleHint,
    labels: input.labels,
    context_budget: {
      max_input_tokens: input.maxInputTokens ?? 4000,
      max_output_tokens: input.maxOutputTokens ?? 1500,
      context_strategy: "artifact_refs_first"
    },
    permissions: {
      allow_network: input.allowNetwork ?? false,
      allow_shell: input.allowShell ?? false,
      allow_external_write: input.allowExternalWrite ?? false,
      approval_mode: input.approvalMode ?? "ask"
    },
    acceptance_criteria: [
      {
        criterion_id: input.acceptanceCriterionId ?? "deliverable",
        description: input.acceptanceDescription ?? "Produce the requested artifact.",
        required: true
      }
    ],
    created_at: input.createdAt
  };
}

export function buildDriBinding(input: DriBindingInput): DriBinding {
  return {
    schema_version: "v1alpha1",
    dri_binding_id: input.driBindingId,
    taskpack_id: input.taskpackId,
    owner: input.owner,
    reviewers: input.reviewers,
    specialists: input.specialists,
    approvers: input.approvers,
    status: input.status ?? "assigned",
    created_at: input.createdAt
  };
}

export function buildArtifact(input: ArtifactInput): Artifact {
  return {
    schema_version: "v1alpha1",
    artifact_id: input.artifactId,
    taskpack_id: input.taskpackId,
    parent_artifact_ids: input.parentArtifactIds,
    kind: input.kind ?? "report",
    title: input.title,
    summary: input.summary,
    producer: input.producer,
    storage: {
      uri: input.uri,
      mime_type: input.mimeType ?? "text/markdown"
    },
    lineage:
      input.traceId || input.runId
        ? {
            trace_id: input.traceId,
            run_id: input.runId,
            source_taskpack_ids: [input.taskpackId]
          }
        : undefined,
    labels: input.labels,
    version: input.version ?? 1,
    created_at: input.createdAt
  };
}

export async function submitInstitutionalRun(client: GuildControlPlaneClient, run: InstitutionalRun): Promise<void> {
  await client.createTaskpack(run.taskpack);
  await client.createDriBinding(run.driBinding);
  for (const artifact of run.artifacts ?? []) {
    await client.createArtifact(artifact);
  }
}
