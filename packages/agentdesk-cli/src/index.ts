#!/usr/bin/env node
import { existsSync, mkdirSync, writeFileSync } from "node:fs";
import { basename, dirname, join } from "node:path";
import { AgentDesk, defaultAgentDeskConfig, unique } from "@lucid-fdn/agentdesk-core";
import { createAgentReadyIssue, publishGitHubAgentWorkReport, syncGitHubIssueMandates } from "@lucid-fdn/agentdesk-github";
import { serveMcp } from "@lucid-fdn/agentdesk-mcp";
import type { Taskpack } from "@guild/client";

const usage = `Guild AgentDesk

Usage:
  guild-agentdesk init
  guild-agentdesk bootstrap github --repo owner/repo
  guild-agentdesk doctor [--id <uuid>] [--repo owner/repo]
  guild-agentdesk issue create "Fix docs typo" --repo owner/repo --scope "docs/**"
  guild-agentdesk mandate create "Fix failing auth tests"
  guild-agentdesk mandate show --id <uuid>
  guild-agentdesk next [--source local|github] [--repo owner/repo] [--include-claimed]
  guild-agentdesk claim --id <uuid> [--agent codex]
  guild-agentdesk preflight --id <uuid> --action write --path src/auth/login.ts
  guild-agentdesk context compile --id <uuid> --role coder [--budget 12000]
  guild-agentdesk approval request --id <uuid> --reason "Need approval"
  guild-agentdesk approval resolve --approval-id <uuid> --decision approved
  guild-agentdesk proof add --id <uuid> --kind test_report --path test-results.xml
  guild-agentdesk handoff create --id <uuid> --to reviewer --summary "Ready"
  guild-agentdesk verify --id <uuid> [--github-report]
  guild-agentdesk close --id <uuid>
  guild-agentdesk replay export --id <uuid> [--file replay.json]
  guild-agentdesk mcp serve
`;

export async function run(argv: string[], stdout: NodeJS.WritableStream = process.stdout): Promise<void> {
  if (argv[0] === "--") {
    argv = argv.slice(1);
  }
  const [command, subcommand, ...rest] = argv;
  switch (command) {
    case "init":
      return init(rest, stdout);
    case "bootstrap":
      if (subcommand !== "github") throw new Error("unknown bootstrap command");
      return bootstrapGitHub(rest, stdout);
    case "doctor":
      return json(stdout, AgentDesk.open().doctor(opts([subcommand, ...rest].filter(Boolean)).mappedIdRepo));
    case "issue":
      if (subcommand !== "create") throw new Error("unknown issue command");
      return issueCreate(rest, stdout);
    case "mandate":
      return mandate(subcommand, rest, stdout);
    case "next":
      return next([subcommand, ...rest].filter(Boolean), stdout);
    case "claim":
      return claim([subcommand, ...rest].filter(Boolean), stdout);
    case "preflight":
      return preflight([subcommand, ...rest].filter(Boolean), stdout);
    case "context":
      if (subcommand !== "compile") throw new Error("unknown context command");
      return contextCompile(rest, stdout);
    case "approval":
      return approval(subcommand, rest, stdout);
    case "proof":
      if (subcommand !== "add") throw new Error("unknown proof command");
      return proofAdd(rest, stdout);
    case "handoff":
      if (subcommand !== "create") throw new Error("unknown handoff command");
      return handoffCreate(rest, stdout);
    case "verify":
      return verify([subcommand, ...rest].filter(Boolean), stdout);
    case "close":
      return close([subcommand, ...rest].filter(Boolean), stdout);
    case "replay":
      if (subcommand !== "export") throw new Error("unknown replay command");
      return replayExport(rest, stdout);
    case "mcp":
      if (subcommand !== "serve") throw new Error("unknown mcp command");
      return serveMcp();
    case "help":
    case "-h":
    case "--help":
    case undefined:
      stdout.write(usage);
      return;
    default:
      throw new Error(`unknown command ${JSON.stringify(command)}`);
  }
}

async function init(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  new AgentDesk(process.cwd()).init({
    workspace: parsed.string("workspace"),
    mission: parsed.string("mission") || "Ship reliable product changes with accountable AI agents."
  });
  stdout.write("agentdesk-init-ok agentdesk.yaml\n");
}

async function bootstrapGitHub(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  const repo = required(parsed.string("repo") || process.env.GITHUB_REPOSITORY, "--repo is required");
  if (!repo.includes("/")) throw new Error(`--repo must be owner/repo, got ${JSON.stringify(repo)}`);
  const force = parsed.bool("force");
  const version = parsed.string("version") || "v0.1.0-alpha.4";
  const root = process.cwd();
  const config = defaultAgentDeskConfig(parsed.string("workspace") || basename(root), parsed.string("mission") || undefined);
  config.task_sources = [
    { type: "local", path: ".agentdesk/mandates" },
    { type: "github_issues", repo, query: "label:agent:ready state:open" }
  ];
  const files = [];
  files.push(writeBootstrapFile("agentdesk.yaml", yamlString(config), force));
  for (const dir of [".agentdesk/mandates", ".agentdesk/proof", ".agentdesk/replay", ".agentdesk/handoffs", ".agentdesk/approvals", ".agentdesk/claims", ".agentdesk/closed"]) {
    mkdirSync(join(root, dir), { recursive: true });
    files.push({ path: dir, status: "ready" });
  }
  files.push(writeBootstrapFile(".github/ISSUE_TEMPLATE/agent-ready.yml", agentReadyIssueTemplate, force));
  files.push(writeBootstrapFile(".github/workflows/agent-work-contract.yml", agentWorkContractWorkflow(version), force));
  json(stdout, {
    schema_version: "v1alpha1",
    ready: true,
    files,
    labels: ["agent:ready", "priority:p1", "priority:p2", "priority:p3"].map((name) => ({ name, status: "skipped: TypeScript alpha does not mutate labels during bootstrap yet" })),
    next_steps: [
      "Create or label a GitHub issue with agent:ready.",
      `Run \`GITHUB_TOKEN=$(gh auth token) npx guild-agentdesk next --source github --repo ${repo}\`.`,
      "Claim, attach proof, verify, and export replay."
    ]
  });
}

async function issueCreate(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  json(
    stdout,
    await createAgentReadyIssue({
      repo: required(parsed.string("repo") || process.env.GITHUB_REPOSITORY, "--repo is required"),
      title: parsed.positionals.join(" "),
      objective: parsed.string("objective"),
      scope: parsed.string("scope"),
      acceptance: parsed.string("acceptance"),
      priority: parsed.string("priority") || "priority:p2",
      labels: splitCSV(parsed.string("labels")),
      notes: parsed.string("notes")
    })
  );
}

async function mandate(subcommand: string | undefined, args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const desk = AgentDesk.open();
  const parsed = opts(args);
  if (subcommand === "create") {
    const mandate = desk.createMandate({
      title: parsed.positionals.join(" "),
      objective: parsed.string("objective"),
      priority: (parsed.string("priority") || "medium") as Taskpack["priority"],
      role: parsed.string("role") || "builder",
      writable: splitCSV(parsed.string("writable"))
    });
    stdout.write(`mandate-created ${mandate.taskpack_id}\n`);
    return;
  }
  if (subcommand === "show") {
    return json(stdout, desk.loadMandate(required(parsed.string("id"), "--id is required")));
  }
  throw new Error("unknown mandate command");
}

async function next(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  const desk = AgentDesk.open();
  if ((parsed.string("source") || "local") === "github") {
    await syncGitHubIssueMandates(desk, { repo: parsed.string("repo"), query: parsed.string("query") });
  }
  json(stdout, desk.nextMandate(parsed.bool("include-claimed")));
}

async function claim(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  json(
    stdout,
    AgentDesk.open().claimMandate({
      mandateId: required(parsed.string("id"), "--id is required"),
      agent: parsed.string("agent"),
      ttlMinutes: parsed.number("ttl-minutes"),
      force: parsed.bool("force")
    })
  );
}

async function preflight(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  json(
    stdout,
    AgentDesk.open().preflight({
      mandateId: required(parsed.string("id"), "--id is required"),
      action: parsed.string("action"),
      path: parsed.string("path"),
      command: parsed.string("command")
    })
  );
}

async function contextCompile(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  json(stdout, AgentDesk.open().compileContext({ mandateId: required(parsed.string("id"), "--id is required"), role: parsed.string("role") || "coder", budgetTokens: parsed.number("budget") }));
}

async function approval(subcommand: string | undefined, args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  const desk = AgentDesk.open();
  if (subcommand === "request") {
    const approval = desk.requestApproval({
      mandateId: required(parsed.string("id"), "--id is required"),
      reason: required(parsed.string("reason"), "--reason is required"),
      requiredApprovals: parsed.number("required")
    });
    stdout.write(`approval-requested ${approval.approval_id}\n`);
    return;
  }
  if (subcommand === "resolve") {
    const approval = desk.resolveApproval({
      approvalId: required(parsed.string("approval-id"), "--approval-id is required"),
      decision: required(parsed.string("decision"), "--decision is required") as "approved" | "rejected",
      reason: parsed.string("reason"),
      actor: parsed.string("actor")
    });
    stdout.write(`approval-${approval.status} ${approval.approval_id}\n`);
    return;
  }
  throw new Error("unknown approval command");
}

async function proofAdd(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  const artifact = AgentDesk.open().addProof({
    mandateId: required(parsed.string("id"), "--id is required"),
    kind: (parsed.string("kind") || "custom") as never,
    path: required(parsed.string("path"), "--path is required"),
    summary: parsed.string("summary")
  });
  stdout.write(`proof-added ${artifact.artifact_id}\n`);
}

async function handoffCreate(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  const artifact = AgentDesk.open().createHandoff({
    mandateId: required(parsed.string("id"), "--id is required"),
    to: required(parsed.string("to"), "--to is required"),
    summary: required(parsed.string("summary"), "--summary is required")
  });
  stdout.write(`handoff-created ${artifact.artifact_id}\n`);
}

async function verify(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  const desk = AgentDesk.open();
  const mandate = desk.loadMandate(required(parsed.string("id"), "--id is required"));
  const report = desk.verify(mandate.taskpack_id);
  if (parsed.bool("github-report")) {
    await publishGitHubAgentWorkReport(mandate, report, parsed.string("replay-file") || "");
  }
  json(stdout, report);
  if (!report.ready) {
    throw new Error("mandate is not ready");
  }
}

async function close(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  const closed = AgentDesk.open().closeMandate(required(parsed.string("id"), "--id is required"));
  stdout.write(`mandate-closed ${closed.mandate_id} proof_count=${closed.proof_count}\n`);
}

async function replayExport(args: string[], stdout: NodeJS.WritableStream): Promise<void> {
  const parsed = opts(args);
  const bundle = AgentDesk.open().exportReplay(required(parsed.string("id"), "--id is required"), parsed.string("file"));
  if (!parsed.string("file")) {
    json(stdout, bundle);
  }
}

function opts(args: string[]): ParsedOptions {
  const values = new Map<string, string | boolean>();
  const positionals = [];
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (!arg.startsWith("--")) {
      positionals.push(arg);
      continue;
    }
    const [rawKey, inline] = arg.slice(2).split(/=(.*)/s, 2);
    if (inline !== undefined) {
      values.set(rawKey, inline);
      continue;
    }
    const next = args[i + 1];
    if (next && !next.startsWith("--")) {
      values.set(rawKey, next);
      i++;
    } else {
      values.set(rawKey, true);
    }
  }
  return {
    positionals,
    string: (key) => {
      const value = values.get(key);
      return typeof value === "string" ? value : undefined;
    },
    number: (key) => {
      const value = values.get(key);
      return typeof value === "string" ? Number(value) : undefined;
    },
    bool: (key) => values.get(key) === true,
    get mappedIdRepo() {
      return { mandateId: this.string("id"), repo: this.string("repo") };
    }
  };
}

type ParsedOptions = {
  positionals: string[];
  string(key: string): string | undefined;
  number(key: string): number | undefined;
  bool(key: string): boolean;
  mappedIdRepo: { mandateId?: string; repo?: string };
};

function writeBootstrapFile(path: string, content: string, force: boolean): { path: string; status: string } {
  if (existsSync(path) && !force) {
    return { path, status: "exists" };
  }
  mkdirSync(dirname(path), { recursive: true });
  writeFileSync(path, content.endsWith("\n") ? content : `${content}\n`);
  return { path, status: force ? "written" : "created" };
}

function yamlString(value: unknown): string {
  // Keep the CLI package dependency-light by using JSON-compatible YAML.
  return JSON.stringify(value, null, 2);
}

function json(stdout: NodeJS.WritableStream, value: unknown): void {
  stdout.write(`${JSON.stringify(value, null, 2)}\n`);
}

function required(value: string | undefined, message: string): string {
  if (!value?.trim()) throw new Error(message);
  return value;
}

function splitCSV(value?: string): string[] {
  return unique((value ?? "").split(",").map((item) => item.trim()).filter(Boolean));
}

const agentReadyIssueTemplate = `name: Agent-ready mandate
description: Create a task that an autonomous agent can claim through Guild AgentDesk.
title: "[agent] "
labels:
  - agent:ready
body:
  - type: textarea
    id: objective
    attributes:
      label: Objective
    validations:
      required: true
  - type: textarea
    id: scope
    attributes:
      label: Allowed scope
    validations:
      required: true
  - type: textarea
    id: acceptance
    attributes:
      label: Acceptance criteria
      value: |
        - Tests pass or failure is explained.
        - Every modified file is listed.
        - A proof artifact is attached.
        - A reviewer handoff is created.
`;

function agentWorkContractWorkflow(version: string): string {
  return `name: Agent Work Contract

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  verify-agent-work:
    runs-on: ubuntu-latest
    if: contains(github.event.pull_request.body || '', 'AgentDesk-Mandate:')
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
      - run: corepack enable
      - run: corepack pnpm install --frozen-lockfile
      - name: Verify Agent Work Contract
        env:
          GITHUB_TOKEN: \${{ secrets.GITHUB_TOKEN }}
          GITHUB_REPOSITORY: \${{ github.repository }}
        run: |
          MANDATE_ID="$(printf '%s' "\${{ github.event.pull_request.body }}" | sed -n 's/^AgentDesk-Mandate: //p' | head -n1)"
          corepack pnpm --dir packages/agentdesk-cli exec guild-agentdesk verify --id "$MANDATE_ID" --github-report
`;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  run(process.argv.slice(2)).catch((error: unknown) => {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = 1;
  });
}
