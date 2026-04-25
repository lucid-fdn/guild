import { createHash, randomUUID } from "node:crypto";
import { existsSync, mkdirSync, readFileSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { dirname, basename, extname, join, resolve, relative } from "node:path";
import { pathToFileURL } from "node:url";
import { parse, stringify } from "yaml";
import type {
  ActorRef,
  ApprovalRequest,
  Artifact,
  ContextPack,
  PreflightDecision,
  ReplayBundle,
  Taskpack,
  WorkspaceConstitution
} from "@guild/client";

type AcceptanceCriteria = Taskpack["acceptance_criteria"][number];

export type AgentDeskClaim = {
  schema_version: "v1alpha1";
  mandate_id: string;
  agent: string;
  claimed_at: string;
  expires_at: string;
};

export type VerifyReport = {
  schema_version: "v1alpha1";
  mandate_id: string;
  ready: boolean;
  proof_count: number;
  required_proof_kinds: string[];
  present_proof_kinds: string[];
  missing_proof_kinds?: string[];
  pending_approval_count: number;
  open_issues?: string[];
};

export type CreateMandateInput = {
  title: string;
  objective?: string;
  priority?: Taskpack["priority"];
  role?: string;
  writable?: string[];
};

export type ProofInput = {
  mandateId: string;
  kind?: Artifact["kind"];
  path: string;
  summary?: string;
};

export type PreflightInput = {
  mandateId: string;
  action?: string;
  path?: string;
  command?: string;
};

export type ContextInput = {
  mandateId: string;
  role?: string;
  budgetTokens?: number;
};

export type HandoffInput = {
  mandateId: string;
  to: string;
  summary: string;
};

const AGENTDESK_DIRS = [
  ".agentdesk/mandates",
  ".agentdesk/proof",
  ".agentdesk/replay",
  ".agentdesk/handoffs",
  ".agentdesk/approvals",
  ".agentdesk/claims",
  ".agentdesk/closed"
];

export function defaultAgentDeskConfig(workspace: string, mission = "Ship reliable product changes with accountable AI agents."): WorkspaceConstitution {
  return {
    schema_version: "v1alpha1",
    workspace,
    mission,
    defaults: {
      max_runtime_minutes: 45,
      max_cost_usd: 5,
      context_budget_tokens: 12000
    },
    task_sources: [{ type: "local", path: ".agentdesk/mandates" }],
    scope: {
      writable: ["src/**", "tests/**", "docs/**"],
      forbidden: [".env", "infra/prod/**", "billing/**"]
    },
    approval_rules: [
      { when: "touches_forbidden_path", require: "human" },
      { when: "runs_destructive_command", require: "human" },
      { when: "pushes_to_main", require: "human" }
    ],
    success_criteria: [
      "Tests pass or failure is explained.",
      "Every modified file is listed.",
      "A proof artifact is attached.",
      "A reviewer handoff is created."
    ],
    escalation: {
      default_owner: "@owner",
      channels: [{ type: "cli_prompt" }]
    }
  };
}

export class AgentDesk {
  readonly root: string;

  constructor(root = process.cwd()) {
    this.root = resolve(root);
  }

  static findRoot(start = process.cwd()): string {
    let current = resolve(start);
    for (;;) {
      if (existsSync(join(current, "agentdesk.yaml"))) {
        return current;
      }
      const parent = dirname(current);
      if (parent === current) {
        throw new Error("agentdesk.yaml not found; run `guild-agentdesk init`");
      }
      current = parent;
    }
  }

  static open(start = process.cwd()): AgentDesk {
    return new AgentDesk(AgentDesk.findRoot(start));
  }

  init(input: { workspace?: string; mission?: string } = {}): string {
    const configPath = join(this.root, "agentdesk.yaml");
    if (existsSync(configPath)) {
      throw new Error("agentdesk.yaml already exists");
    }
    const config = defaultAgentDeskConfig(input.workspace ?? basename(this.root), input.mission);
    this.writeText("agentdesk.yaml", stringify(config));
    for (const dir of AGENTDESK_DIRS) {
      mkdirSync(join(this.root, dir), { recursive: true });
    }
    return configPath;
  }

  ensureDirs(): void {
    for (const dir of AGENTDESK_DIRS) {
      mkdirSync(join(this.root, dir), { recursive: true });
    }
  }

  loadConfig(): WorkspaceConstitution {
    const config = normalizeWorkspaceConfig(parse(readFileSync(join(this.root, "agentdesk.yaml"), "utf8")) as WorkspaceConstitution & { version?: string });
    assertWorkspaceConfig(config);
    return config;
  }

  writeConfig(config: WorkspaceConstitution): void {
    assertWorkspaceConfig(config);
    this.writeText("agentdesk.yaml", stringify(config));
  }

  createMandate(input: CreateMandateInput): Taskpack {
    const title = input.title.trim();
    if (!title) {
      throw new Error("mandate title is required");
    }
    const config = this.loadConfig();
    const scopes = input.writable?.length ? input.writable : (config.scope.writable ?? []);
    const mandate: Taskpack = {
      schema_version: "v1alpha1",
      taskpack_id: randomUUID(),
      title,
      objective: input.objective?.trim() || title,
      task_type: "implementation",
      priority: input.priority ?? "medium",
      requested_by: actor("human", "agentdesk-cli"),
      role_hint: normalizeRoleHint(input.role ?? "builder") as Taskpack["role_hint"],
      labels: ["agentdesk", "open"],
      context_budget: {
        max_input_tokens: Math.max(config.defaults.context_budget_tokens, 256),
        max_output_tokens: 1024,
        context_strategy: "artifact_refs_first"
      },
      permissions: {
        allow_network: false,
        allow_shell: true,
        allow_external_write: false,
        approval_mode: "ask",
        scopes
      },
      acceptance_criteria: defaultAcceptance(config.success_criteria) as Taskpack["acceptance_criteria"],
      created_at: now()
    };
    assertTaskpack(mandate);
    this.writeJSON(mandatePath(mandate.taskpack_id), mandate);
    return mandate;
  }

  saveMandate(mandate: Taskpack): Taskpack {
    assertTaskpack(mandate);
    this.writeJSON(mandatePath(mandate.taskpack_id), mandate);
    return mandate;
  }

  loadMandate(id: string): Taskpack {
    assertUUID(id, "--id must be a UUID");
    const mandate = this.readJSON<Taskpack>(mandatePath(id));
    assertTaskpack(mandate);
    return mandate;
  }

  listOpenMandates(includeClaimed = false): Taskpack[] {
    const dir = join(this.root, ".agentdesk", "mandates");
    if (!existsSync(dir)) {
      return [];
    }
    const mandates = readdirSync(dir)
      .filter((file) => file.endsWith(".json"))
      .map((file) => this.readJSON<Taskpack>(join(".agentdesk", "mandates", file)))
      .filter((mandate) => {
        assertTaskpack(mandate);
        if (existsSync(join(this.root, ".agentdesk", "closed", `${mandate.taskpack_id}.json`))) {
          return false;
        }
        return includeClaimed || !this.loadClaim(mandate.taskpack_id).active;
      });
    return mandates.sort((a, b) => priorityRank(b.priority) - priorityRank(a.priority) || a.created_at.localeCompare(b.created_at));
  }

  nextMandate(includeClaimed = false): Taskpack {
    const [mandate] = this.listOpenMandates(includeClaimed);
    if (!mandate) {
      throw new Error("no open mandates found");
    }
    return mandate;
  }

  claimMandate(input: { mandateId: string; agent?: string; ttlMinutes?: number; force?: boolean }): AgentDeskClaim {
    const mandate = this.loadMandate(input.mandateId);
    const ttl = input.ttlMinutes ?? 120;
    if (ttl < 1) {
      throw new Error("--ttl-minutes must be >= 1");
    }
    const existing = this.loadClaim(mandate.taskpack_id);
    if (existing.active && !input.force) {
      throw new Error(`mandate already claimed by ${existing.claim?.agent ?? "unknown"} until ${existing.claim?.expires_at ?? "unknown"}`);
    }
    const claimedAt = new Date();
    const claim: AgentDeskClaim = {
      schema_version: "v1alpha1",
      mandate_id: mandate.taskpack_id,
      agent: input.agent || process.env.AGENT_NAME || "agentdesk-agent",
      claimed_at: claimedAt.toISOString(),
      expires_at: new Date(claimedAt.getTime() + ttl * 60_000).toISOString()
    };
    this.writeJSON(claimPath(mandate.taskpack_id), claim);
    return claim;
  }

  loadClaim(mandateId: string): { claim?: AgentDeskClaim; active: boolean } {
    const path = join(this.root, claimPath(mandateId));
    if (!existsSync(path)) {
      return { active: false };
    }
    const claim = this.readJSON<AgentDeskClaim>(claimPath(mandateId));
    if (claim.schema_version !== "v1alpha1" || !isUUID(claim.mandate_id)) {
      throw new Error(`claim ${mandateId} is invalid`);
    }
    const expiresAt = Date.parse(claim.expires_at);
    if (Number.isNaN(expiresAt)) {
      throw new Error(`claim ${mandateId} has invalid expires_at`);
    }
    return { claim, active: Date.now() < expiresAt };
  }

  compileContext(input: ContextInput): ContextPack {
    const config = this.loadConfig();
    const mandate = this.loadMandate(input.mandateId);
    const pack: ContextPack = {
      schema_version: "v1alpha1",
      mandate_id: mandate.taskpack_id,
      role: input.role ?? "coder",
      budget_tokens: input.budgetTokens ?? config.defaults.context_budget_tokens,
      must_read: refsToPaths(mandate.references ?? []),
      may_read: unique([...(config.scope.writable ?? []), ...(mandate.permissions.scopes ?? [])]),
      may_write: unique(mandate.permissions.scopes ?? []),
      forbidden: config.scope.forbidden ?? [],
      summary: mandate.objective,
      proof_required: proofKindsFromAcceptance(mandate.acceptance_criteria),
      omitted_reasons: ["Full transcript omitted; agentdesk emits bounded context packs from mandates, scope, and artifacts."]
    };
    return pack;
  }

  preflight(input: PreflightInput): PreflightDecision {
    const config = this.loadConfig();
    const mandate = this.loadMandate(input.mandateId);
    const action = (input.action || inferPreflightAction(input.path, input.command)) as PreflightDecision["action"];
    const base: PreflightDecision = {
      schema_version: "v1alpha1",
      mandate_id: mandate.taskpack_id,
      action,
      path: input.path,
      command: input.command,
      decision: "allow",
      reason: "Action is within mandate and workspace policy.",
      approval_required: false,
      matched_rules: ["default.allow"]
    };
    if (input.path) {
      const normalized = normalizePath(input.path);
      if (matchesAny(normalized, config.scope.forbidden ?? [])) {
        return needsApproval(base, "Path matches forbidden workspace scope.", "scope.forbidden");
      }
      if (action === "write" && !matchesAny(normalized, [...(config.scope.writable ?? []), ...(mandate.permissions.scopes ?? [])])) {
        return needsApproval(base, "Path is outside writable workspace and mandate scope.", "scope.write_outside_allowlist");
      }
    }
    if (input.command) {
      const command = input.command.trim().toLowerCase();
      if (isDestructiveCommand(command)) {
        return needsApproval({ ...base, action: "run" }, "Command looks destructive and requires approval.", "approval.runs_destructive_command");
      }
      if (command.startsWith("git push")) {
        return needsApproval({ ...base, action: "git_push" }, "Pushing requires approval.", "approval.pushes_to_main");
      }
      if (!mandate.permissions.allow_shell) {
        return needsApproval({ ...base, action: "run" }, "Mandate does not allow shell commands.", "permissions.allow_shell");
      }
    }
    if (action === "network" && !mandate.permissions.allow_network) {
      return needsApproval({ ...base, action: "network" }, "Mandate does not allow network access.", "permissions.allow_network");
    }
    return base;
  }

  requestApproval(input: { mandateId: string; reason: string; requiredApprovals?: number }): ApprovalRequest {
    const mandate = this.loadMandate(input.mandateId);
    if (!input.reason.trim()) {
      throw new Error("--reason is required");
    }
    const approval: ApprovalRequest = {
      schema_version: "v1alpha1",
      approval_id: randomUUID(),
      taskpack_id: mandate.taskpack_id,
      requested_by: actor("agent", "agentdesk-cli", "agentdesk"),
      reason: input.reason,
      required_approvals: input.requiredApprovals ?? 1,
      status: "pending",
      created_at: now()
    };
    this.writeJSON(approvalPath(mandate.taskpack_id, approval.approval_id), approval);
    return approval;
  }

  resolveApproval(input: { approvalId: string; decision: "approved" | "rejected"; reason?: string; actor?: string }): ApprovalRequest {
    assertUUID(input.approvalId, "--approval-id must be a UUID");
    const found = this.findApproval(input.approvalId);
    const decidedAt = now();
    const approval = found.approval;
    approval.approvals = [
      ...(approval.approvals ?? []),
      {
        actor: actor("human", input.actor ?? "human-reviewer"),
        decision: input.decision,
        reason: input.reason,
        decided_at: decidedAt
      }
    ];
    if (input.decision === "rejected") {
      approval.status = "rejected";
      approval.decided_at = decidedAt;
    } else if (approval.approvals.length >= approval.required_approvals) {
      approval.status = "approved";
      approval.decided_at = decidedAt;
    }
    this.writeJSON(relative(this.root, found.path), approval);
    return approval;
  }

  addProof(input: ProofInput): Artifact {
    const mandate = this.loadMandate(input.mandateId);
    const artifact = this.buildProofArtifact(mandate, input.kind ?? "custom", input.path, input.summary);
    this.writeJSON(proofPath(mandate.taskpack_id, artifact.artifact_id), artifact);
    return artifact;
  }

  createHandoff(input: HandoffInput): Artifact {
    const mandate = this.loadMandate(input.mandateId);
    if (!input.to.trim()) {
      throw new Error("--to is required");
    }
    if (!input.summary.trim()) {
      throw new Error("--summary is required");
    }
    const artifactId = randomUUID();
    const handoffPath = join(".agentdesk", "handoffs", mandate.taskpack_id, `${artifactId}.md`);
    this.writeText(handoffPath, `# Handoff\n\nMandate: ${mandate.taskpack_id}\nTo: ${input.to}\nCreated: ${now()}\n\n${input.summary.trim()}\n`);
    const artifact = this.buildProofArtifact(mandate, "handoff_summary", handoffPath, `Handoff to ${input.to}: ${input.summary.trim()}`);
    artifact.artifact_id = artifactId;
    this.writeJSON(proofPath(mandate.taskpack_id, artifact.artifact_id), artifact);
    return artifact;
  }

  loadProofArtifacts(mandateId: string): Artifact[] {
    const dir = join(this.root, ".agentdesk", "proof", mandateId);
    if (!existsSync(dir)) {
      return [];
    }
    return readdirSync(dir)
      .filter((file) => file.endsWith(".json"))
      .map((file) => this.readJSON<Artifact>(join(".agentdesk", "proof", mandateId, file)))
      .sort((a, b) => a.created_at.localeCompare(b.created_at));
  }

  loadApprovals(mandateId: string): ApprovalRequest[] {
    const dir = join(this.root, ".agentdesk", "approvals", mandateId);
    if (!existsSync(dir)) {
      return [];
    }
    return readdirSync(dir)
      .filter((file) => file.endsWith(".json"))
      .map((file) => this.readJSON<ApprovalRequest>(join(".agentdesk", "approvals", mandateId, file)))
      .sort((a, b) => a.created_at.localeCompare(b.created_at));
  }

  verify(mandateId: string): VerifyReport {
    const mandate = this.loadMandate(mandateId);
    const artifacts = this.loadProofArtifacts(mandate.taskpack_id);
    const approvals = this.loadApprovals(mandate.taskpack_id);
    const required = proofKindsFromAcceptance(mandate.acceptance_criteria);
    const present = orderProofKinds(unique(artifacts.map((artifact) => artifact.kind)));
    const missing = required.filter((kind) => !present.includes(kind));
    const pending = approvals.filter((approval) => approval.status === "pending").length;
    const issues = [];
    if (artifacts.length === 0) {
      issues.push("at least one proof artifact is required");
    }
    if (pending > 0) {
      issues.push("pending approvals must be resolved");
    }
    for (const kind of missing) {
      issues.push(`missing proof kind: ${kind}`);
    }
    return {
      schema_version: "v1alpha1",
      mandate_id: mandate.taskpack_id,
      ready: issues.length === 0,
      proof_count: artifacts.length,
      required_proof_kinds: required,
      present_proof_kinds: present,
      missing_proof_kinds: missing.length ? missing : undefined,
      pending_approval_count: pending,
      open_issues: issues.length ? issues : undefined
    };
  }

  closeMandate(mandateId: string): { mandate_id: string; proof_count: number; closed_at: string } {
    const mandate = this.loadMandate(mandateId);
    const report = this.verify(mandate.taskpack_id);
    if (!report.ready) {
      throw new Error(`cannot close mandate: ${(report.open_issues ?? []).join("; ")}`);
    }
    const record = {
      schema_version: "v1alpha1",
      mandate_id: mandate.taskpack_id,
      proof_count: report.proof_count,
      closed_at: now()
    };
    this.writeJSON(join(".agentdesk", "closed", `${mandate.taskpack_id}.json`), record);
    return record;
  }

  exportReplay(mandateId: string, file?: string): ReplayBundle {
    const mandate = this.loadMandate(mandateId);
    const bundle: ReplayBundle = {
      schema_version: "v1alpha1",
      root_taskpack_id: mandate.taskpack_id,
      taskpack: mandate,
      taskpacks: [mandate],
      dri_bindings: [],
      artifacts: this.loadProofArtifacts(mandate.taskpack_id),
      promotion_records: []
    };
    if (file) {
      this.writeJSON(file, bundle);
    }
    return bundle;
  }

  doctor(input: { mandateId?: string; repo?: string } = {}): Record<string, unknown> {
    const checks: Array<{ name: string; status: "ok" | "warn" | "fail"; detail?: string }> = [];
    checks.push({ name: "agentdesk.yaml", status: existsSync(join(this.root, "agentdesk.yaml")) ? "ok" : "fail" });
    for (const dir of AGENTDESK_DIRS) {
      checks.push({ name: dir, status: existsSync(join(this.root, dir)) ? "ok" : "warn" });
    }
    checks.push({ name: "GITHUB_TOKEN", status: process.env.GITHUB_TOKEN ? "ok" : "warn", detail: process.env.GITHUB_TOKEN ? undefined : "not set" });
    if (input.mandateId) {
      try {
        const report = this.verify(input.mandateId);
        checks.push({ name: "proof readiness", status: report.ready ? "ok" : "warn", detail: (report.open_issues ?? []).join("; ") });
      } catch (error) {
        checks.push({ name: "proof readiness", status: "fail", detail: errorMessage(error) });
      }
    }
    return {
      schema_version: "v1alpha1",
      ready: checks.every((check) => check.status !== "fail"),
      root: this.root,
      repo: input.repo,
      checks
    };
  }

  private buildProofArtifact(mandate: Taskpack, kind: Artifact["kind"], pathValue: string, summary?: string): Artifact {
    if (!pathValue.trim()) {
      throw new Error("--path is required");
    }
    const absolute = resolve(this.root, pathValue);
    const info = statSync(absolute);
    return {
      schema_version: "v1alpha1",
      artifact_id: randomUUID(),
      taskpack_id: mandate.taskpack_id,
      kind,
      title: basename(pathValue),
      summary: summary || "Proof artifact published from local agentdesk workflow.",
      producer: actor("agent", "agentdesk-cli", "agentdesk"),
      storage: {
        uri: pathToFileURL(absolute).toString(),
        mime_type: mimeTypeForPath(pathValue),
        sha256: sha256File(absolute),
        size_bytes: info.size
      },
      labels: ["agentdesk", "proof"],
      version: 1,
      created_at: now()
    };
  }

  private findApproval(approvalId: string): { approval: ApprovalRequest; path: string } {
    const approvalsRoot = join(this.root, ".agentdesk", "approvals");
    for (const mandateDir of existsSync(approvalsRoot) ? readdirSync(approvalsRoot) : []) {
      const path = join(approvalsRoot, mandateDir, `${approvalId}.json`);
      if (existsSync(path)) {
        return { approval: JSON.parse(readFileSync(path, "utf8")) as ApprovalRequest, path };
      }
    }
    throw new Error(`approval ${approvalId} not found`);
  }

  private readJSON<T>(path: string): T {
    return JSON.parse(readFileSync(join(this.root, path), "utf8")) as T;
  }

  private writeJSON(path: string, payload: unknown): void {
    const target = join(this.root, path);
    mkdirSync(dirname(target), { recursive: true });
    writeFileSync(target, `${JSON.stringify(payload, null, 2)}\n`);
  }

  private writeText(path: string, content: string): void {
    const target = join(this.root, path);
    mkdirSync(dirname(target), { recursive: true });
    writeFileSync(target, content.endsWith("\n") ? content : `${content}\n`);
  }
}

export function mandatePath(id: string): string {
  return join(".agentdesk", "mandates", `${id}.json`);
}

export function claimPath(id: string): string {
  return join(".agentdesk", "claims", `${id}.json`);
}

export function proofPath(mandateId: string, artifactId: string): string {
  return join(".agentdesk", "proof", mandateId, `${artifactId}.json`);
}

export function approvalPath(mandateId: string, approvalId: string): string {
  return join(".agentdesk", "approvals", mandateId, `${approvalId}.json`);
}

export function actor(actorType: ActorRef["actor_type"], displayName: string, orchestrator?: string): ActorRef {
  return {
    actor_id: randomUUID(),
    actor_type: actorType,
    display_name: displayName,
    orchestrator
  };
}

export function deterministicUUID(seed: string): string {
  const bytes = createHash("sha1").update(seed).digest().subarray(0, 16);
  bytes[6] = (bytes[6] & 0x0f) | 0x50;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = bytes.toString("hex");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`;
}

export function proofKindsFromAcceptance(criteria: AcceptanceCriteria[]): Artifact["kind"][] {
  const required: Artifact["kind"][] = ["test_report", "changed_files", "handoff_summary"];
  for (const criterion of criteria) {
    const lower = criterion.description.toLowerCase();
    if (lower.includes("screenshot")) {
      required.push("screenshot");
    }
    if (lower.includes("benchmark")) {
      required.push("benchmark_result");
    }
  }
  return orderProofKinds(unique(required));
}

export function defaultAcceptance(criteria?: string[]): AcceptanceCriteria[] {
  const values = criteria?.length ? criteria : ["A proof artifact is attached."];
  return values.map((description, index) => ({
    criterion_id: `criterion_${String(index + 1).padStart(2, "0")}`,
    description,
    required: true
  }));
}

export function unique<T>(values: T[]): T[] {
  return [...new Set(values.filter(Boolean))];
}

function assertWorkspaceConfig(config: WorkspaceConstitution): void {
  if (config.schema_version !== "v1alpha1" || !config.workspace || !config.defaults || !config.scope) {
    throw new Error("agentdesk.yaml is not a valid v1alpha1 workspace constitution");
  }
}

function normalizeWorkspaceConfig(config: WorkspaceConstitution & { version?: string }): WorkspaceConstitution {
  if (!config.schema_version && config.version) {
    config.schema_version = config.version as WorkspaceConstitution["schema_version"];
  }
  return config;
}

function assertTaskpack(taskpack: Taskpack): void {
  if (taskpack.schema_version !== "v1alpha1" || !isUUID(taskpack.taskpack_id) || !taskpack.title || !taskpack.objective) {
    throw new Error("invalid v1alpha1 mandate/taskpack");
  }
}

function assertUUID(value: string, message: string): void {
  if (!isUUID(value)) {
    throw new Error(message);
  }
}

function isUUID(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
}

function now(): string {
  return new Date().toISOString();
}

function normalizeRoleHint(role: string): string {
  switch (role.trim().toLowerCase()) {
    case "coder":
    case "implementer":
    case "developer":
      return "builder";
    case "critic":
      return "skeptic";
    default:
      return role;
  }
}

function priorityRank(priority: Taskpack["priority"]): number {
  return { critical: 4, high: 3, medium: 2, low: 1 }[priority] ?? 0;
}

function refsToPaths(refs: string[]): string[] {
  return refs.map((ref) => {
    try {
      const parsed = new URL(ref);
      return parsed.protocol === "file:" ? parsed.pathname : ref;
    } catch {
      return ref;
    }
  });
}

function inferPreflightAction(pathValue?: string, command?: string): string {
  if (command) {
    return "run";
  }
  if (pathValue) {
    return "read";
  }
  return "custom";
}

function needsApproval(base: PreflightDecision, reason: string, rule: string): PreflightDecision {
  return {
    ...base,
    decision: "needs_approval",
    reason,
    approval_required: true,
    matched_rules: [rule]
  };
}

function isDestructiveCommand(command: string): boolean {
  return ["rm -rf", "rm -fr", "git reset --hard", "git clean -fd", "chmod -r", "chown -r", "drop database", "terraform apply", "kubectl delete"].some((item) =>
    command.includes(item)
  );
}

function matchesAny(pathValue: string, patterns: string[]): boolean {
  return patterns.some((pattern) => matchesPattern(pathValue, pattern));
}

function matchesPattern(pathValue: string, pattern: string): boolean {
  const normalized = normalizePath(pattern);
  const target = normalizePath(pathValue);
  if (normalized.endsWith("/**")) {
    const prefix = normalized.slice(0, -3);
    return target === prefix || target.startsWith(`${prefix}/`);
  }
  if (normalized.endsWith("/*")) {
    const prefix = normalized.slice(0, -2);
    return target.startsWith(`${prefix}/`) && !target.slice(prefix.length + 1).includes("/");
  }
  return target === normalized;
}

function normalizePath(value: string): string {
  return value.replaceAll("\\", "/").replace(/^\.\//, "").replace(/\/+/g, "/");
}

function sha256File(path: string): string {
  return createHash("sha256").update(readFileSync(path)).digest("hex");
}

function mimeTypeForPath(path: string): string {
  switch (extname(path).toLowerCase()) {
    case ".json":
      return "application/json";
    case ".xml":
      return "application/xml";
    case ".md":
      return "text/markdown";
    case ".txt":
    case ".log":
      return "text/plain";
    case ".png":
      return "image/png";
    case ".jpg":
    case ".jpeg":
      return "image/jpeg";
    default:
      return "application/octet-stream";
  }
}

function orderProofKinds<T extends string>(kinds: T[]): T[] {
  const rank = new Map<string, number>([
    ["test_report", 0],
    ["changed_files", 1],
    ["handoff_summary", 2]
  ]);
  return [...kinds].sort((a, b) => (rank.get(a) ?? 99) - (rank.get(b) ?? 99) || a.localeCompare(b));
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
