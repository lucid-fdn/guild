import { stdin as defaultStdin, stdout as defaultStdout } from "node:process";
import { AgentDesk } from "@lucid-fdn/agentdesk-core";
import { syncGitHubIssueMandates } from "@lucid-fdn/agentdesk-github";

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

export type McpTool = {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
};

export const agentDeskMcpTools: McpTool[] = [
  tool("guild_get_next_mandate", "Return the next open mandate from local files or GitHub Issues.", {
    source: { type: "string", enum: ["local", "github"] },
    repo: { type: "string" },
    query: { type: "string" },
    include_claimed: { type: "boolean" }
  }),
  tool("guild_claim_mandate", "Create a local lease so multiple agents do not pick the same mandate.", {
    taskpack_id: { type: "string", format: "uuid" },
    agent: { type: "string" },
    ttl_minutes: { type: "integer", minimum: 1 },
    force: { type: "boolean" }
  }, ["taskpack_id"]),
  tool("guild_compile_context", "Compile bounded role-specific context for one mandate.", {
    taskpack_id: { type: "string", format: "uuid" },
    role: { type: "string" },
    budget_tokens: { type: "integer", minimum: 256 }
  }, ["taskpack_id"]),
  tool("guild_check_preflight", "Check whether an agent action is allowed or needs approval.", {
    taskpack_id: { type: "string", format: "uuid" },
    action: { type: "string" },
    path: { type: "string" },
    command: { type: "string" }
  }, ["taskpack_id"]),
  tool("guild_publish_artifact", "Attach a proof artifact to a mandate.", {
    taskpack_id: { type: "string", format: "uuid" },
    kind: { type: "string" },
    path: { type: "string" },
    summary: { type: "string" }
  }, ["taskpack_id", "path"]),
  tool("guild_create_handoff", "Create a handoff proof artifact.", {
    taskpack_id: { type: "string", format: "uuid" },
    to: { type: "string" },
    summary: { type: "string" }
  }, ["taskpack_id", "to", "summary"]),
  tool("guild_verify_mandate", "Verify proof, approvals, and readiness.", {
    taskpack_id: { type: "string", format: "uuid" }
  }, ["taskpack_id"]),
  tool("guild_export_replay_bundle", "Export a replay bundle for one mandate.", {
    taskpack_id: { type: "string", format: "uuid" }
  }, ["taskpack_id"])
];

export async function handleMcpRequest(request: JsonRpcRequest, cwd = process.cwd()): Promise<JsonRpcResponse | null> {
  if (request.method.startsWith("notifications/")) {
    return null;
  }
  try {
    switch (request.method) {
      case "initialize":
        return ok(request.id, {
          protocolVersion: "2024-11-05",
          capabilities: { tools: {} },
          serverInfo: { name: "guild-agentdesk-mcp", version: "0.1.0-alpha.0" }
        });
      case "tools/list":
        return ok(request.id, { tools: agentDeskMcpTools });
      case "tools/call": {
        const params = request.params ?? {};
        if (typeof params.name !== "string") {
          throw new Error("tools/call requires params.name");
        }
        const result = await callTool(params.name, (params.arguments ?? {}) as Record<string, unknown>, cwd);
        return ok(request.id, result);
      }
      default:
        return fail(request.id, -32601, `unknown MCP method: ${request.method}`);
    }
  } catch (error) {
    return fail(request.id, -32000, error instanceof Error ? error.message : String(error));
  }
}

export async function callTool(name: string, args: Record<string, unknown>, cwd = process.cwd()): Promise<{ content: Array<{ type: "text"; text: string }>; isError?: boolean }> {
  try {
    const desk = AgentDesk.open(cwd);
    let result: unknown;
    switch (name) {
      case "guild_get_next_mandate":
        if (stringArg(args, "source") === "github") {
          await syncGitHubIssueMandates(desk, { repo: stringArg(args, "repo"), query: stringArg(args, "query") });
        }
        result = desk.nextMandate(Boolean(args.include_claimed));
        break;
      case "guild_claim_mandate":
        result = desk.claimMandate({ mandateId: requiredString(args, "taskpack_id"), agent: stringArg(args, "agent"), ttlMinutes: numberArg(args, "ttl_minutes"), force: Boolean(args.force) });
        break;
      case "guild_compile_context":
        result = desk.compileContext({ mandateId: requiredString(args, "taskpack_id"), role: stringArg(args, "role") || "coder", budgetTokens: numberArg(args, "budget_tokens") });
        break;
      case "guild_check_preflight":
        result = desk.preflight({ mandateId: requiredString(args, "taskpack_id"), action: stringArg(args, "action"), path: stringArg(args, "path"), command: stringArg(args, "command") });
        break;
      case "guild_publish_artifact":
        result = desk.addProof({ mandateId: requiredString(args, "taskpack_id"), kind: stringArg(args, "kind") as never, path: requiredString(args, "path"), summary: stringArg(args, "summary") });
        break;
      case "guild_create_handoff":
        result = desk.createHandoff({ mandateId: requiredString(args, "taskpack_id"), to: requiredString(args, "to"), summary: requiredString(args, "summary") });
        break;
      case "guild_verify_mandate":
        result = desk.verify(requiredString(args, "taskpack_id"));
        break;
      case "guild_export_replay_bundle":
        result = desk.exportReplay(requiredString(args, "taskpack_id"));
        break;
      default:
        throw new Error(`unknown Guild MCP tool: ${name}`);
    }
    return textResult(JSON.stringify(result, null, 2));
  } catch (error) {
    return textResult(error instanceof Error ? error.message : String(error), true);
  }
}

export async function serveMcp(input = defaultStdin, output = defaultStdout): Promise<void> {
  let buffer = Buffer.alloc(0);
  for await (const chunk of input) {
    buffer = Buffer.concat([buffer, Buffer.from(chunk)]);
    for (;;) {
      const parsed = readFrame(buffer);
      if (!parsed) break;
      buffer = Buffer.from(parsed.rest);
      const request = JSON.parse(parsed.body.toString("utf8")) as JsonRpcRequest;
      const response = await handleMcpRequest(request);
      if (response) {
        writeFrame(output, JSON.stringify(response));
      }
    }
  }
}

function readFrame(buffer: Buffer): { body: Buffer; rest: Buffer } | null {
  const headerEnd = buffer.indexOf("\r\n\r\n");
  if (headerEnd < 0) return null;
  const header = buffer.subarray(0, headerEnd).toString("utf8");
  const match = /Content-Length:\s*(\d+)/i.exec(header);
  if (!match) throw new Error("MCP frame missing Content-Length");
  const length = Number(match[1]);
  const bodyStart = headerEnd + 4;
  const bodyEnd = bodyStart + length;
  if (buffer.length < bodyEnd) return null;
  return { body: buffer.subarray(bodyStart, bodyEnd), rest: buffer.subarray(bodyEnd) };
}

function writeFrame(output: NodeJS.WritableStream, body: string): void {
  output.write(`Content-Length: ${Buffer.byteLength(body)}\r\n\r\n${body}`);
}

function tool(name: string, description: string, properties: Record<string, unknown>, required: string[] = []): McpTool {
  return {
    name,
    description,
    inputSchema: {
      type: "object",
      properties,
      required,
      additionalProperties: false
    }
  };
}

function ok(id: JsonRpcRequest["id"], result: unknown): JsonRpcResponse {
  return { jsonrpc: "2.0", id: id ?? null, result };
}

function fail(id: JsonRpcRequest["id"], code: number, message: string): JsonRpcResponse {
  return { jsonrpc: "2.0", id: id ?? null, error: { code, message } };
}

function textResult(text: string, isError = false): { content: Array<{ type: "text"; text: string }>; isError?: boolean } {
  return { content: [{ type: "text", text }], isError: isError || undefined };
}

function stringArg(args: Record<string, unknown>, key: string): string | undefined {
  return typeof args[key] === "string" ? args[key] : undefined;
}

function requiredString(args: Record<string, unknown>, key: string): string {
  const value = stringArg(args, key);
  if (!value) throw new Error(`${key} is required`);
  return value;
}

function numberArg(args: Record<string, unknown>, key: string): number | undefined {
  return typeof args[key] === "number" ? args[key] : undefined;
}
