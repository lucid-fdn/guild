import type { ActorRef, Artifact, DriBinding, Taskpack } from "@guild/client";
import {
  buildArtifact,
  buildDriBinding,
  buildTaskpack,
  submitInstitutionalRun,
  type GuildControlPlaneClient
} from "@guild/adapter-core";

export type A2AAgentCard = {
  id: string;
  name: string;
  url?: string;
  orchestrator?: string;
  skills?: string[];
};

export type A2ATaskEnvelope = {
  taskId: string;
  title: string;
  objective: string;
  requester: ActorRef;
  dri: A2AAgentCard;
  createdAt: string;
  institutionId?: string;
  priority?: Taskpack["priority"];
  taskType?: Taskpack["task_type"];
  contextUri?: string;
  labels?: string[];
};

export type A2AResultEnvelope = {
  artifactId: string;
  taskId: string;
  title: string;
  producer: A2AAgentCard;
  uri: string;
  createdAt: string;
  kind?: Artifact["kind"];
  summary?: string;
  mimeType?: string;
  traceId?: string;
  runId?: string;
};

export type A2AArtifactMessage = {
  role: "agent";
  parts: Array<{
    kind: "guild_artifact_ref";
    artifact_id: string;
    taskpack_id: string;
    uri: string;
    title: string;
    summary?: string;
  }>;
};

export function actorFromAgentCard(card: A2AAgentCard): ActorRef {
  return {
    actor_id: card.id,
    actor_type: "agent",
    display_name: card.name,
    orchestrator: card.orchestrator ?? card.url ?? "a2a"
  };
}

export function buildTaskpackFromA2ATask(input: A2ATaskEnvelope): Taskpack {
  return buildTaskpack({
    taskpackId: input.taskId,
    title: input.title,
    objective: input.objective,
    requestedBy: input.requester,
    createdAt: input.createdAt,
    institutionId: input.institutionId,
    priority: input.priority,
    taskType: input.taskType ?? "operations",
    roleHint: "dri",
    labels: [...(input.labels ?? []), ...(input.dri.skills ?? []).map((skill) => `skill:${skill}`)],
    acceptanceDescription: input.contextUri
      ? `Complete the A2A task using context artifact ${input.contextUri}.`
      : "Complete the A2A task and publish durable artifacts."
  });
}

export function buildDriBindingFromA2ATask(input: A2ATaskEnvelope, driBindingId: string): DriBinding {
  return buildDriBinding({
    driBindingId,
    taskpackId: input.taskId,
    owner: actorFromAgentCard(input.dri),
    createdAt: input.createdAt,
    status: "assigned"
  });
}

export function buildArtifactFromA2AResult(input: A2AResultEnvelope): Artifact {
  return buildArtifact({
    artifactId: input.artifactId,
    taskpackId: input.taskId,
    title: input.title,
    producer: actorFromAgentCard(input.producer),
    uri: input.uri,
    createdAt: input.createdAt,
    kind: input.kind,
    summary: input.summary,
    mimeType: input.mimeType,
    traceId: input.traceId,
    runId: input.runId
  });
}

export function toA2AArtifactMessage(artifact: Artifact): A2AArtifactMessage {
  return {
    role: "agent",
    parts: [
      {
        kind: "guild_artifact_ref",
        artifact_id: artifact.artifact_id,
        taskpack_id: artifact.taskpack_id,
        uri: artifact.storage.uri,
        title: artifact.title,
        summary: artifact.summary
      }
    ]
  };
}

export async function submitA2ATask(
  client: GuildControlPlaneClient,
  task: A2ATaskEnvelope,
  options: {
    driBindingId: string;
    results?: A2AResultEnvelope[];
  }
): Promise<void> {
  await submitInstitutionalRun(client, {
    taskpack: buildTaskpackFromA2ATask(task),
    driBinding: buildDriBindingFromA2ATask(task, options.driBindingId),
    artifacts: options.results?.map(buildArtifactFromA2AResult)
  });
}
