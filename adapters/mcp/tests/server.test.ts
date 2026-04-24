import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import { LocalAgentDeskClient, encodeMcpMessage, handleMcpRequest } from "../src/server";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

test("MCP server lists agentdesk tools", async () => {
  const response = await handleMcpRequest({
    jsonrpc: "2.0",
    id: 1,
    method: "tools/list"
  });

  assert.equal(response?.jsonrpc, "2.0");
  const result = response?.result as { tools: Array<{ name: string }> };
  assert.ok(result.tools.some((tool) => tool.name === "guild_get_next_mandate"));
  assert.ok(result.tools.some((tool) => tool.name === "guild_claim_mandate"));
  assert.ok(result.tools.some((tool) => tool.name === "guild_check_preflight"));
  assert.ok(result.tools.some((tool) => tool.name === "guild_close_mandate"));
});

test("LocalAgentDeskClient wraps local agentdesk workspace", async () => {
  const dir = mkdtempSync(path.join(os.tmpdir(), "guild-agentdesk-mcp-"));
  const previousCLI = process.env.GUILD_CLI;
  const guildCLI = path.join(dir, "guild");
  execFileSync("go", ["build", "-o", guildCLI, "./cli/cmd/guild"], {
    cwd: repoRoot,
    stdio: ["ignore", "pipe", "pipe"]
  });
  process.env.GUILD_CLI = guildCLI;
  try {
    const client = new LocalAgentDeskClient(dir, guildCLI);
    runGuild(dir, ["agentdesk", "init", "--workspace", "mcp-demo"]);
    runGuild(dir, ["agentdesk", "mandate", "create", "Fix MCP workflow", "--writable", "src/**,tests/**"]);

    const next = await client.getNextMandate();
    assert.equal(next.title, "Fix MCP workflow");

    const claimResponse = await handleMcpRequest(
      {
        jsonrpc: "2.0",
        id: 2,
        method: "tools/call",
        params: {
          name: "guild_claim_mandate",
          arguments: {
            taskpack_id: next.taskpack_id,
            agent: "mcp-agent",
            ttl_minutes: 30
          }
        }
      },
      client
    );
    const claimText = ((claimResponse?.result as { content: Array<{ text: string }> }).content[0].text);
    const claim = JSON.parse(claimText) as { mandate_id: string; agent: string };
    assert.equal(claim.mandate_id, next.taskpack_id);
    assert.equal(claim.agent, "mcp-agent");

    const contextResponse = await handleMcpRequest(
      {
        jsonrpc: "2.0",
        id: 3,
        method: "tools/call",
        params: {
          name: "guild_compile_context",
          arguments: {
            taskpack_id: next.taskpack_id,
            role: "coder"
          }
        }
      },
      client
    );
    const text = ((contextResponse?.result as { content: Array<{ text: string }> }).content[0].text);
    const context = JSON.parse(text) as { mandate_id: string };
    assert.equal(context.mandate_id, next.taskpack_id);

    writeFileSync(path.join(dir, "test-results.xml"), "<testsuite failures=\"0\"></testsuite>\n");
    runGuild(dir, ["agentdesk", "proof", "add", "--id", next.taskpack_id, "--kind", "test_report", "--path", "test-results.xml"]);
    writeFileSync(path.join(dir, "changed-files.json"), "[\"src/index.ts\"]\n");
    runGuild(dir, ["agentdesk", "proof", "add", "--id", next.taskpack_id, "--kind", "changed_files", "--path", "changed-files.json"]);
    await client.createHandoff({ taskpack_id: next.taskpack_id, to: "reviewer", summary: "Ready for review." });
    const verify = await client.verifyMandate({ taskpack_id: next.taskpack_id }) as { ready: boolean };
    assert.equal(verify.ready, true);
  } finally {
    if (previousCLI === undefined) {
      delete process.env.GUILD_CLI;
    } else {
      process.env.GUILD_CLI = previousCLI;
    }
    rmSync(dir, { recursive: true, force: true });
  }
});

test("encodeMcpMessage emits Content-Length frames", () => {
  const frame = encodeMcpMessage({ jsonrpc: "2.0", id: 1, result: { ok: true } });
  assert.match(frame, /^Content-Length: \d+\r\n\r\n/);
  assert.match(frame, /"ok":true/);
});

function runGuild(cwd: string, args: string[]): string {
  return execFileSync(process.env.GUILD_CLI ?? "guild", args, {
    cwd,
    encoding: "utf8"
  });
}
