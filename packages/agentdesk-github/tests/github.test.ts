import test from "node:test";
import assert from "node:assert/strict";
import { AgentDesk } from "@lucid-fdn/agentdesk-core";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { renderAgentReadyIssueBody, taskpackFromGitHubIssue } from "../src/index";

test("GitHub issue body is structured for agents", () => {
  const body = renderAgentReadyIssueBody("Fix docs", "docs/**", "Attach proof", "Be careful");
  assert.match(body, /## Objective\nFix docs/);
  assert.match(body, /## Allowed scope\ndocs\/\*\*/);
  assert.match(body, /- Attach proof/);
});

test("GitHub issues become deterministic mandates", () => {
  const desk = new AgentDesk(mkdtempSync(join(tmpdir(), "agentdesk-github-")));
  desk.init({ workspace: "demo" });
  const config = desk.loadConfig();
  const mandate = taskpackFromGitHubIssue(config, "lucid-fdn/demo", {
    number: 7,
    title: "Fix docs",
    body: "Please update setup.",
    html_url: "https://github.com/lucid-fdn/demo/issues/7",
    labels: [{ name: "agent:ready" }, { name: "priority:p1" }],
    user: { login: "quentin" }
  });
  assert.equal(mandate.priority, "high");
  const again = taskpackFromGitHubIssue(config, "lucid-fdn/demo", {
    number: 7,
    title: "Fix docs",
    body: "Please update setup.",
    html_url: "https://github.com/lucid-fdn/demo/issues/7",
    labels: [{ name: "agent:ready" }, { name: "priority:p1" }],
    user: { login: "quentin" }
  });
  assert.equal(mandate.taskpack_id, again.taskpack_id);
});
