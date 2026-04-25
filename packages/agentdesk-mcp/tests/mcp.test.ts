import test from "node:test";
import assert from "node:assert/strict";
import { handleMcpRequest } from "../src/index";

test("MCP lists AgentDesk tools", async () => {
  const response = await handleMcpRequest({ id: 1, method: "tools/list" });
  assert.equal(response?.jsonrpc, "2.0");
  const result = response?.result as { tools: Array<{ name: string }> };
  assert.ok(result.tools.some((tool) => tool.name === "guild_get_next_mandate"));
});
