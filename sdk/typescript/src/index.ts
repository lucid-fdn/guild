import type { Artifact, DriBinding, PromotionRecord, ReplayBundle, Taskpack } from "./spec";

export type {
  ActorRef,
  ApprovalRequest,
  Artifact,
  ArtifactRef,
  ContextPack,
  DriBinding,
  Labels,
  MetricDelta,
  PreflightDecision,
  PromotionRecord,
  ReplayBundle,
  Taskpack,
  WorkspaceConstitution
} from "./spec";

export type GuildStatus = {
  name: string;
  version: string;
  mode: string;
  ui_origin: string;
  services: Array<Record<string, unknown>>;
};

type CollectionResponse<T> = {
  items: T[];
};

export class GuildAPIError extends Error {
  constructor(
    readonly status: number,
    readonly message: string
  ) {
    super(`Guild API request failed with ${status}: ${message}`);
    this.name = "GuildAPIError";
  }
}

export class GuildClient {
  private readonly baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl.replace(/\/+$/, "");
  }

  async getStatus(): Promise<GuildStatus> {
    return this.get<GuildStatus>("/api/v1/status");
  }

  async listTaskpacks(): Promise<Taskpack[]> {
    return this.getCollection<Taskpack>("/api/v1/taskpacks");
  }

  async getTaskpack(taskpackId: string): Promise<Taskpack> {
    return this.get<Taskpack>(`/api/v1/taskpacks/${taskpackId}`);
  }

  async createTaskpack(taskpack: Taskpack): Promise<Taskpack> {
    return this.post("/api/v1/taskpacks", taskpack);
  }

  async listDriBindings(): Promise<DriBinding[]> {
    return this.getCollection<DriBinding>("/api/v1/dri-bindings");
  }

  async getDriBinding(driBindingId: string): Promise<DriBinding> {
    return this.get<DriBinding>(`/api/v1/dri-bindings/${driBindingId}`);
  }

  async createDriBinding(binding: DriBinding): Promise<DriBinding> {
    return this.post("/api/v1/dri-bindings", binding);
  }

  async listArtifacts(): Promise<Artifact[]> {
    return this.getCollection<Artifact>("/api/v1/artifacts");
  }

  async listArtifactsForTaskpack(taskpackId: string): Promise<Artifact[]> {
    return this.getCollection<Artifact>(`/api/v1/taskpacks/${taskpackId}/artifacts`);
  }

  async getArtifact(artifactId: string): Promise<Artifact> {
    return this.get<Artifact>(`/api/v1/artifacts/${artifactId}`);
  }

  async createArtifact(artifact: Artifact): Promise<Artifact> {
    return this.post("/api/v1/artifacts", artifact);
  }

  async listPromotionRecords(): Promise<PromotionRecord[]> {
    return this.getCollection<PromotionRecord>("/api/v1/promotion-records");
  }

  async getPromotionRecord(promotionRecordId: string): Promise<PromotionRecord> {
    return this.get<PromotionRecord>(`/api/v1/promotion-records/${promotionRecordId}`);
  }

  async createPromotionRecord(record: PromotionRecord): Promise<PromotionRecord> {
    return this.post("/api/v1/promotion-records", record);
  }

  async exportReplayBundle(taskpackId: string): Promise<ReplayBundle> {
    return this.get<ReplayBundle>(`/api/v1/replay/taskpacks/${taskpackId}`);
  }

  private async getCollection<T>(path: string): Promise<T[]> {
    const payload = await this.get<CollectionResponse<T>>(path);
    return payload.items;
  }

  private async get<T>(path: string): Promise<T> {
    return this.request<T>(path, { method: "GET" });
  }

  private async post<T>(path: string, payload: T): Promise<T> {
    return this.request<T>(path, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });
  }

  private async request<T>(path: string, init: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, init);
    if (!response.ok) {
      throw new GuildAPIError(response.status, await readErrorMessage(response));
    }
    return response.json() as Promise<T>;
  }
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const payload = (await response.json()) as { error?: unknown };
    if (typeof payload.error === "string" && payload.error.length > 0) {
      return payload.error;
    }
  } catch {
    // Fall back to the protocol status text when the body is not a Guild error envelope.
  }
  return response.statusText || "request failed";
}
