import { Buffer } from "node:buffer";
import { AgentDesk, actor, defaultAcceptance, deterministicUUID, unique, type VerifyReport } from "@lucid-fdn/agentdesk-core";
import type { Artifact, Taskpack, WorkspaceConstitution } from "@guild/client";

export const defaultGitHubIssueQuery = "label:agent:ready state:open";

export type GitHubIssue = {
  number: number;
  title: string;
  body?: string;
  html_url: string;
  state?: string;
  created_at?: string;
  labels?: Array<{ name: string }>;
  user?: { login: string };
};

export type GitHubIssueCreateReport = {
  schema_version: "v1alpha1";
  repo: string;
  issue_number: number;
  title: string;
  url: string;
  labels: string[];
};

type GitHubIssueSearchResponse = {
  items: GitHubIssue[];
};

type GitHubIssueCreateRequest = {
  title: string;
  body: string;
  labels?: string[];
};

export async function syncGitHubIssueMandates(desk: AgentDesk, input: { repo?: string; query?: string } = {}): Promise<Taskpack[]> {
  const config = desk.loadConfig();
  const source = resolveGitHubTaskSource(config, input.repo, input.query);
  const issues = await fetchGitHubIssues(source.repo, source.query);
  const mandates = issues.map((issue) => taskpackFromGitHubIssue(config, source.repo, issue));
  for (const mandate of mandates) {
    desk.saveMandate(mandate);
  }
  return mandates;
}

export function resolveGitHubTaskSource(config: WorkspaceConstitution, repoOverride?: string, queryOverride?: string): { repo: string; query: string } {
  let repo = repoOverride?.trim() || "";
  let query = queryOverride?.trim() || "";
  for (const source of config.task_sources ?? []) {
    if (source.type !== "github_issues") {
      continue;
    }
    repo ||= source.repo ?? "";
    query ||= source.query ?? "";
    break;
  }
  repo ||= process.env.GITHUB_REPOSITORY ?? "";
  query ||= defaultGitHubIssueQuery;
  if (!repo) {
    throw new Error("GitHub repo is required; pass --repo, set GITHUB_REPOSITORY, or add a github_issues task_source");
  }
  if (!repo.includes("/")) {
    throw new Error(`GitHub repo must be owner/repo, got ${JSON.stringify(repo)}`);
  }
  return { repo, query };
}

export async function fetchGitHubIssues(repo: string, query: string): Promise<GitHubIssue[]> {
  let searchQuery = query;
  if (!searchQuery.includes("repo:")) {
    searchQuery = `repo:${repo} ${searchQuery}`;
  }
  if (!searchQuery.includes("type:") && !searchQuery.includes("is:")) {
    searchQuery += " type:issue";
  }
  const endpoint = `${githubAPIURL()}/search/issues?q=${encodeURIComponent(searchQuery)}&per_page=20`;
  const response = await fetch(endpoint, { headers: githubHeaders() });
  if (!response.ok) {
    throw new Error(`GitHub issue search failed with ${response.status}: ${await response.text()}`);
  }
  const payload = (await response.json()) as GitHubIssueSearchResponse;
  return payload.items;
}

export function taskpackFromGitHubIssue(config: WorkspaceConstitution, repo: string, issue: GitHubIssue): Taskpack {
  const labels = githubLabelNames(issue.labels ?? []);
  const requester = issue.user?.login || "github";
  return {
    schema_version: "v1alpha1",
    taskpack_id: deterministicUUID(`github-issue:${issue.html_url}`),
    title: issue.title,
    objective: githubIssueObjective(issue),
    task_type: githubTaskType(labels),
    priority: githubPriority(labels),
    requested_by: {
      ...actor("human", requester),
      orchestrator: "github",
      endpoint: `https://github.com/${repo}`
    },
    role_hint: githubRoleHint(labels),
    references: [issue.html_url],
    labels: unique(["agentdesk", "github", `issue-${issue.number}`, ...sanitizeLabels(labels)]),
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
      scopes: githubScopes(config, labels)
    },
    acceptance_criteria: defaultAcceptance(config.success_criteria) as Taskpack["acceptance_criteria"],
    created_at: issue.created_at || new Date().toISOString()
  };
}

export async function createAgentReadyIssue(input: {
  repo: string;
  title: string;
  objective?: string;
  scope?: string;
  acceptance?: string;
  priority?: string;
  labels?: string[];
  notes?: string;
}): Promise<GitHubIssueCreateReport> {
  const repo = resolveGitHubRepo(input.repo);
  const title = input.title.trim();
  if (!title) {
    throw new Error("issue title is required");
  }
  const labels = unique(["agent:ready", input.priority ?? "priority:p2", ...(input.labels ?? [])]);
  const issue = await createGitHubIssue(repo, {
    title,
    body: renderAgentReadyIssueBody(input.objective || title, input.scope || "docs/**", input.acceptance || defaultIssueAcceptance(), input.notes),
    labels
  });
  return {
    schema_version: "v1alpha1",
    repo,
    issue_number: issue.number,
    title: issue.title,
    url: issue.html_url,
    labels
  };
}

export async function createGitHubIssue(repo: string, payload: GitHubIssueCreateRequest): Promise<GitHubIssue> {
  const response = await fetch(`${githubAPIURL()}/repos/${repo}/issues`, {
    method: "POST",
    headers: { ...githubHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });
  if (response.status !== 201) {
    throw new Error(`GitHub issue create failed with ${response.status}: ${await response.text()}`);
  }
  return (await response.json()) as GitHubIssue;
}

export async function publishGitHubAgentWorkReport(mandate: Taskpack, report: VerifyReport, replayRef = ""): Promise<string> {
  const body = renderAgentWorkReport(mandate, report, replayRef);
  if (process.env.GITHUB_STEP_SUMMARY) {
    const { appendFileSync } = await import("node:fs");
    appendFileSync(process.env.GITHUB_STEP_SUMMARY, `${body}\n`);
  }
  const repo = process.env.GITHUB_REPOSITORY;
  const prNumber = await githubPullRequestNumber();
  if (process.env.GITHUB_TOKEN && repo && prNumber) {
    await postGitHubIssueComment(repo, prNumber, body);
  }
  return body;
}

export function renderAgentWorkReport(mandate: Taskpack, report: VerifyReport, replayRef = ""): string {
  const status = report.ready ? "passed" : "failed";
  const approvals = report.pending_approval_count > 0 ? `${report.pending_approval_count} pending` : "resolved";
  const proof = report.present_proof_kinds.length ? report.present_proof_kinds.join(", ") : "none";
  const lines = [
    `### Agent Work Contract: ${status}`,
    "",
    `- Mandate: ${mandate.title} (\`${mandate.taskpack_id}\`)`,
    `- Proof: ${proof}`,
    `- Approvals: ${approvals}`,
    `- Replay: ${replayRef ? "attached" : "not attached"}`
  ];
  if (replayRef) {
    lines.push(`- Replay ref: \`${replayRef}\``);
  }
  if (report.open_issues?.length) {
    lines.push("", "Open issues:", ...report.open_issues.map((issue) => `- ${issue}`));
  }
  return lines.join("\n");
}

export function renderAgentReadyIssueBody(objective: string, scope: string, acceptance: string, notes?: string): string {
  const sections = [`## Objective\n${objective.trim()}`, `## Allowed scope\n${scope.trim()}`, `## Acceptance criteria\n${normalizeIssueList(acceptance)}`];
  if (notes?.trim()) {
    sections.push(`## Notes for the agent\n${notes.trim()}`);
  }
  return `${sections.join("\n\n")}\n`;
}

export function resolveGitHubRepo(repoOverride?: string): string {
  const repo = repoOverride?.trim() || process.env.GITHUB_REPOSITORY || "";
  if (!repo) {
    throw new Error("GitHub repo is required; pass --repo or set GITHUB_REPOSITORY");
  }
  if (!repo.includes("/")) {
    throw new Error(`GitHub repo must be owner/repo, got ${JSON.stringify(repo)}`);
  }
  return repo;
}

async function postGitHubIssueComment(repo: string, number: string, body: string): Promise<void> {
  const response = await fetch(`${githubAPIURL()}/repos/${repo}/issues/${number}/comments`, {
    method: "POST",
    headers: { ...githubHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify({ body })
  });
  if (response.status !== 201) {
    throw new Error(`GitHub comment failed with ${response.status}: ${await response.text()}`);
  }
}

async function githubPullRequestNumber(): Promise<string> {
  if (process.env.GITHUB_PR_NUMBER) {
    return process.env.GITHUB_PR_NUMBER;
  }
  if (!process.env.GITHUB_EVENT_PATH) {
    return "";
  }
  const { readFileSync } = await import("node:fs");
  const payload = JSON.parse(readFileSync(process.env.GITHUB_EVENT_PATH, "utf8")) as {
    number?: number;
    pull_request?: { number?: number };
    issue?: { number?: number };
  };
  return String(payload.pull_request?.number ?? payload.issue?.number ?? payload.number ?? "");
}

function githubAPIURL(): string {
  return (process.env.GITHUB_API_URL || "https://api.github.com").replace(/\/+$/, "");
}

function githubHeaders(): HeadersInit {
  const headers: Record<string, string> = {
    Accept: "application/vnd.github+json",
    "User-Agent": "guild-agentdesk",
    "X-GitHub-Api-Version": "2022-11-28"
  };
  if (process.env.GITHUB_TOKEN) {
    headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`;
  }
  return headers;
}

function githubIssueObjective(issue: GitHubIssue): string {
  const body = issue.body?.trim();
  if (!body) {
    return issue.title;
  }
  return `${issue.title}\n\nSource issue:\n${body.length > 2000 ? `${body.slice(0, 2000)}...` : body}`;
}

function githubLabelNames(labels: Array<{ name: string }>): string[] {
  return labels.map((label) => label.name).filter(Boolean);
}

function githubPriority(labels: string[]): Taskpack["priority"] {
  for (const label of labels.map((value) => value.toLowerCase())) {
    if (["priority:p0", "priority:critical", "p0", "critical"].includes(label)) return "critical";
    if (["priority:p1", "priority:high", "p1", "high"].includes(label)) return "high";
    if (["priority:p3", "priority:low", "p3", "low"].includes(label)) return "low";
  }
  return "medium";
}

function githubTaskType(labels: string[]): Taskpack["task_type"] {
  for (const label of labels.map((value) => value.toLowerCase())) {
    if (["type:review", "task:review"].includes(label)) return "review";
    if (["type:research", "task:research"].includes(label)) return "research";
    if (["type:triage", "task:triage"].includes(label)) return "triage";
    if (["type:ops", "type:operations", "task:operations"].includes(label)) return "operations";
  }
  return "implementation";
}

function githubRoleHint(labels: string[]): Taskpack["role_hint"] {
  for (const label of labels.map((value) => value.toLowerCase())) {
    if (["role:reviewer", "agent:reviewer"].includes(label)) return "reviewer";
    if (["role:explorer", "agent:explorer"].includes(label)) return "explorer";
    if (["role:skeptic", "agent:skeptic"].includes(label)) return "skeptic";
    if (["role:specialist", "agent:specialist"].includes(label)) return "specialist";
  }
  return "builder";
}

function githubScopes(config: WorkspaceConstitution, labels: string[]): string[] {
  const scopes = [...(config.scope.writable ?? [])];
  for (const label of labels) {
    if (label.toLowerCase().startsWith("scope:")) {
      const scope = label.slice("scope:".length).trim();
      if (scope) scopes.push(`${scope}/**`);
    }
  }
  return unique(scopes);
}

function sanitizeLabels(labels: string[]): string[] {
  return labels
    .map((label) =>
      label
        .toLowerCase()
        .trim()
        .replaceAll(":", "-")
        .replace(/[^a-z0-9._/-]+/g, "-")
        .replace(/^[-._/]+|[-._/]+$/g, "")
    )
    .filter(Boolean);
}

function normalizeIssueList(value: string): string {
  const lines = value
    .trim()
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => (line.startsWith("- ") ? line : `- ${line}`));
  return lines.length ? lines.join("\n") : "- Complete the mandate.";
}

function defaultIssueAcceptance(): string {
  return ["Tests pass or failure is explained.", "Every modified file is listed.", "A proof artifact is attached.", "A reviewer handoff is created."].join("\n");
}

export function base64JSON(value: unknown): string {
  return Buffer.from(JSON.stringify(value)).toString("base64");
}
