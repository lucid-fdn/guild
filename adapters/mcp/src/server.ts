#!/usr/bin/env node
import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import type {
  ApprovalRequest,
  Artifact,
  ContextPack,
  DriBinding,
  PreflightDecision,
  ReplayBundle,
  Taskpack
} from "@guild/client";
import {
  createGuildMcpBridge,
  guildMcpTools,
  type CheckPreflightArgs,
  type ClaimMandateArgs,
  type CloseMandateArgs,
  type CompileContextArgs,
  type CreateHandoffArgs,
  type GuildMcpClient,
  type McpToolCall,
  type RequestApprovalArgs,
  type VerifyMandateArgs
} from "./index";

type JsonRpcRequest = {
  jsonrpc?: string;
  id?: string | number | null;
  method: string;
  params?: Record<string, unknown>;
};

type JsonRpcResponse = {
  jsonrpc: "2.0";
  id: string | number | null;
  result?: unknown;
  error?: {
    code: number;
    message: string;
  };
};

export class LocalAgentDeskClient implements GuildMcpClient {
  constructor(
    private readonly cwd = process.cwd(),
    private readonly command = process.env.GUILD_CLI || "guild"
  ) {}

  async getNextMandate(): Promise<Taskpack> {
    const args = ["agentdesk", "next"];
    if (process.env.GUILD_AGENTDESK_SOURCE) {
      args.push("--source", process.env.GUILD_AGENTDESK_SOURCE);
    }
    if (process.env.GITHUB_REPOSITORY) {
      args.push("--repo", process.env.GITHUB_REPOSITORY);
    }
    return this.runJSON<Taskpack>(args);
  }

  async claimMandate(input: ClaimMandateArgs): Promise<unknown> {
    const args = ["agentdesk", "claim", "--id", input.taskpack_id];
    if (input.agent) {
      args.push("--agent", input.agent);
    }
    if (input.ttl_minutes) {
      args.push("--ttl-minutes", String(input.ttl_minutes));
    }
    if (input.force) {
      args.push("--force");
    }
    return this.runJSON(args);
  }

  async compileContext(input: CompileContextArgs): Promise<ContextPack> {
    const args = ["agentdesk", "context", "compile", "--id", input.taskpack_id, "--role", input.role];
    if (input.budget_tokens) {
      args.push("--budget", String(input.budget_tokens));
    }
    return this.runJSON<ContextPack>(args);
  }

  async checkPreflight(input: CheckPreflightArgs): Promise<PreflightDecision> {
    const args = ["agentdesk", "preflight", "--id", input.taskpack_id, "--action", input.action];
    if (input.path) {
      args.push("--path", input.path);
    }
    if (input.command) {
      args.push("--command", input.command);
    }
    return this.runJSON<PreflightDecision>(args);
  }

  async requestApproval(input: RequestApprovalArgs): Promise<ApprovalRequest> {
    const output = this.runText([
      "agentdesk",
      "approval",
      "request",
      "--id",
      input.taskpack_id,
      "--reason",
      input.reason,
      "--required",
      String(input.required_approvals ?? 1)
    ]);
    const approvalID = output.trim().split(/\s+/)[1];
    return this.readJSON<ApprovalRequest>(path.join(".agentdesk", "approvals", input.taskpack_id, `${approvalID}.json`));
  }

  async createHandoff(input: CreateHandoffArgs): Promise<Artifact> {
    const output = this.runText([
      "agentdesk",
      "handoff",
      "create",
      "--id",
      input.taskpack_id,
      "--to",
      input.to,
      "--summary",
      input.summary
    ]);
    const artifactID = output.trim().split(/\s+/)[1];
    return this.readJSON<Artifact>(path.join(".agentdesk", "proof", input.taskpack_id, `${artifactID}.json`));
  }

  async verifyMandate(input: VerifyMandateArgs): Promise<unknown> {
    return this.runJSON(["agentdesk", "verify", "--id", input.taskpack_id]);
  }

  async closeMandate(input: CloseMandateArgs): Promise<unknown> {
    const output = this.runText(["agentdesk", "close", "--id", input.taskpack_id]);
    return { status: "closed", output: output.trim() };
  }

  async exportReplayBundle(taskpackId: string): Promise<ReplayBundle> {
    return this.runJSON<ReplayBundle>(["agentdesk", "replay", "export", "--id", taskpackId]);
  }

  async createTaskpack(taskpack: Taskpack): Promise<Taskpack> {
    this.writeJSON(path.join(".agentdesk", "mandates", `${taskpack.taskpack_id}.json`), taskpack);
    return taskpack;
  }

  async createDriBinding(binding: DriBinding): Promise<DriBinding> {
    this.writeJSON(path.join(".agentdesk", "dri-bindings", `${binding.dri_binding_id}.json`), binding);
    return binding;
  }

  async createArtifact(artifact: Artifact): Promise<Artifact> {
    this.writeJSON(path.join(".agentdesk", "proof", artifact.taskpack_id, `${artifact.artifact_id}.json`), artifact);
    return artifact;
  }

  private runJSON<T>(args: string[]): T {
    return JSON.parse(this.runText(args)) as T;
  }

  private runText(args: string[]): string {
    const commandParts = this.command.split(/\s+/).filter(Boolean);
    const executable = commandParts[0] ?? "guild";
    const commandArgs = [...commandParts.slice(1), ...args];
    return execFileSync(executable, commandArgs, {
      cwd: this.cwd,
      encoding: "utf8",
      env: process.env,
      stdio: ["ignore", "pipe", "pipe"]
    });
  }

  private readJSON<T>(relativePath: string): T {
    return JSON.parse(readFileSync(path.join(this.cwd, relativePath), "utf8")) as T;
  }

  private writeJSON(relativePath: string, payload: unknown): void {
    const target = path.join(this.cwd, relativePath);
    ensureDir(path.dirname(target));
    writeFileSync(target, `${JSON.stringify(payload, null, 2)}\n`);
  }
}

export async function handleMcpRequest(request: JsonRpcRequest, client: GuildMcpClient = new LocalAgentDeskClient()): Promise<JsonRpcResponse | null> {
  if (request.method.startsWith("notifications/")) {
    return null;
  }
  try {
    switch (request.method) {
      case "initialize":
        return ok(request.id, {
          protocolVersion: "2024-11-05",
          capabilities: { tools: {} },
          serverInfo: { name: "guild-agentdesk-mcp", version: "0.1.0" }
        });
      case "tools/list":
        return ok(request.id, {
          tools: guildMcpTools.map((tool) => ({
            name: tool.name,
            description: tool.description,
            inputSchema: tool.inputSchema
          }))
        });
      case "tools/call": {
        const params = request.params ?? {};
        const name = params.name;
        if (typeof name !== "string") {
          throw new Error("tools/call requires params.name");
        }
        const bridge = createGuildMcpBridge(client);
        const result = await bridge.handle({
          name,
          arguments: (params.arguments ?? {}) as McpToolCall["arguments"]
        });
        return ok(request.id, result);
      }
      default:
        return fail(request.id, -32601, `unknown MCP method: ${request.method}`);
    }
  } catch (error) {
    return fail(request.id, -32000, error instanceof Error ? error.message : "MCP server error");
  }
}

export function encodeMcpMessage(payload: unknown): string {
  const body = JSON.stringify(payload);
  return `Content-Length: ${Buffer.byteLength(body, "utf8")}\r\n\r\n${body}`;
}

function ok(id: JsonRpcRequest["id"], result: unknown): JsonRpcResponse {
  return { jsonrpc: "2.0", id: id ?? null, result };
}

function fail(id: JsonRpcRequest["id"], code: number, message: string): JsonRpcResponse {
  return { jsonrpc: "2.0", id: id ?? null, error: { code, message } };
}

function ensureDir(dir: string): void {
  if (existsSync(dir)) {
    return;
  }
  mkdirSync(dir, { recursive: true });
}

async function runStdioServer(): Promise<void> {
  let buffer: Buffer<ArrayBufferLike> = Buffer.alloc(0);
  process.stdin.on("data", async (chunk: Buffer<ArrayBufferLike>) => {
    buffer = Buffer.concat([buffer, chunk]);
    while (true) {
      const parsed = readMcpFrame(buffer);
      if (!parsed) {
        break;
      }
      buffer = parsed.rest;
      const response = await handleMcpRequest(JSON.parse(parsed.body) as JsonRpcRequest);
      if (response) {
        process.stdout.write(encodeMcpMessage(response));
      }
    }
  });
}

function readMcpFrame(buffer: Buffer<ArrayBufferLike>): { body: string; rest: Buffer<ArrayBufferLike> } | null {
  const marker = Buffer.from("\r\n\r\n");
  const headerEnd = buffer.indexOf(marker);
  if (headerEnd === -1) {
    return null;
  }
  const header = buffer.subarray(0, headerEnd).toString("utf8");
  const match = header.match(/Content-Length:\s*(\d+)/i);
  if (!match) {
    throw new Error("MCP frame missing Content-Length");
  }
  const length = Number.parseInt(match[1], 10);
  const bodyStart = headerEnd + marker.length;
  const bodyEnd = bodyStart + length;
  if (buffer.length < bodyEnd) {
    return null;
  }
  return {
    body: buffer.subarray(bodyStart, bodyEnd).toString("utf8"),
    rest: buffer.subarray(bodyEnd)
  };
}

if (process.argv[1] && import.meta.url.endsWith(process.argv[1].replace(/\\/g, "/"))) {
  void runStdioServer();
}
