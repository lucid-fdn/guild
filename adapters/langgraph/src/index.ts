import type { ActorRef, Artifact, DriBinding, Taskpack } from "@guild/client";
import {
  buildArtifact,
  buildDriBinding,
  buildTaskpack,
  submitInstitutionalRun,
  type GuildControlPlaneClient
} from "@guild/adapter-core";

export type LangGraphRunnableConfig = {
  configurable?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  runName?: string;
  tags?: string[];
};

export type LangGraphGuildState = {
  guild?: {
    taskpack?: Taskpack;
    driBinding?: DriBinding;
    artifacts?: Artifact[];
    submitted?: boolean;
  };
  [key: string]: unknown;
};

export type LangGraphTaskInput = {
  taskpackId: string;
  driBindingId: string;
  title: string;
  objective: string;
  requestedBy: ActorRef;
  owner: ActorRef;
  createdAt?: string;
  institutionId?: string;
  priority?: Taskpack["priority"];
  taskType?: Taskpack["task_type"];
  labels?: string[];
};

export type LangGraphArtifactInput = {
  artifactId: string;
  title: string;
  uri: string;
  producer?: ActorRef;
  createdAt?: string;
  kind?: Artifact["kind"];
  summary?: string;
  mimeType?: string;
};

export type LangGraphGuildNodeOptions<State extends LangGraphGuildState> = {
  client: GuildControlPlaneClient;
  task: LangGraphTaskInput | ((state: State, config?: LangGraphRunnableConfig) => LangGraphTaskInput | Promise<LangGraphTaskInput>);
  artifacts?: LangGraphArtifactInput[] | ((state: State, config?: LangGraphRunnableConfig) => LangGraphArtifactInput[] | Promise<LangGraphArtifactInput[]>);
  markSubmitted?: boolean;
};

export type LangGraphGuildNodeResult = {
  guild: {
    taskpack: Taskpack;
    driBinding: DriBinding;
    artifacts: Artifact[];
    submitted: boolean;
  };
};

export function buildLangGraphInstitutionalRun(
  taskInput: LangGraphTaskInput,
  artifactInputs: LangGraphArtifactInput[] = []
): LangGraphGuildNodeResult["guild"] {
  const createdAt = taskInput.createdAt ?? new Date().toISOString();
  const taskpack = buildTaskpack({
    taskpackId: taskInput.taskpackId,
    title: taskInput.title,
    objective: taskInput.objective,
    requestedBy: taskInput.requestedBy,
    createdAt,
    institutionId: taskInput.institutionId,
    priority: taskInput.priority,
    taskType: taskInput.taskType ?? "operations",
    roleHint: "dri",
    labels: ["orchestrator:langgraph", ...(taskInput.labels ?? [])]
  });
  const driBinding = buildDriBinding({
    driBindingId: taskInput.driBindingId,
    taskpackId: taskInput.taskpackId,
    owner: {
      ...taskInput.owner,
      orchestrator: taskInput.owner.orchestrator ?? "langgraph"
    },
    createdAt,
    status: "assigned"
  });
  const artifacts = artifactInputs.map((artifactInput) =>
    buildArtifact({
      artifactId: artifactInput.artifactId,
      taskpackId: taskInput.taskpackId,
      title: artifactInput.title,
      producer: artifactInput.producer ?? driBinding.owner,
      uri: artifactInput.uri,
      createdAt: artifactInput.createdAt ?? createdAt,
      kind: artifactInput.kind,
      summary: artifactInput.summary,
      mimeType: artifactInput.mimeType,
      labels: ["orchestrator:langgraph"]
    })
  );

  return {
    taskpack,
    driBinding,
    artifacts,
    submitted: false
  };
}

export function createGuildLangGraphNode<State extends LangGraphGuildState>(
  options: LangGraphGuildNodeOptions<State>
): (state: State, config?: LangGraphRunnableConfig) => Promise<LangGraphGuildNodeResult> {
  return async (state, config) => {
    const taskInput = await resolve(options.task, state, config);
    const artifactInputs = await resolve(options.artifacts ?? [], state, config);
    const guild = buildLangGraphInstitutionalRun(taskInput, artifactInputs);

    await submitInstitutionalRun(options.client, {
      taskpack: guild.taskpack,
      driBinding: guild.driBinding,
      artifacts: guild.artifacts
    });

    return {
      guild: {
        ...guild,
        submitted: options.markSubmitted ?? true
      }
    };
  };
}

async function resolve<State, Value>(
  value: Value | ((state: State, config?: LangGraphRunnableConfig) => Value | Promise<Value>),
  state: State,
  config?: LangGraphRunnableConfig
): Promise<Value> {
  if (typeof value === "function") {
    return (value as (state: State, config?: LangGraphRunnableConfig) => Value | Promise<Value>)(state, config);
  }
  return value;
}
