import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import assert from "node:assert/strict";
import { AgentDesk } from "../src/index";

test("local AgentDesk loop creates, claims, verifies, and exports replay", () => {
  const root = mkdtempSync(join(tmpdir(), "agentdesk-core-"));
  const desk = new AgentDesk(root);
  desk.init({ workspace: "demo" });
  const mandate = desk.createMandate({ title: "Fix docs", writable: ["docs/**"] });

  assert.equal(desk.nextMandate().taskpack_id, mandate.taskpack_id);
  const claim = desk.claimMandate({ mandateId: mandate.taskpack_id, agent: "codex" });
  assert.equal(claim.agent, "codex");

  const context = desk.compileContext({ mandateId: mandate.taskpack_id, role: "coder" });
  assert.deepEqual(context.may_write, ["docs/**"]);

  writeFileSync(join(root, "test-results.xml"), "<testsuite />\n");
  writeFileSync(join(root, "changed-files.json"), "[\"docs/readme.md\"]\n");
  desk.addProof({ mandateId: mandate.taskpack_id, kind: "test_report", path: "test-results.xml" });
  desk.addProof({ mandateId: mandate.taskpack_id, kind: "changed_files", path: "changed-files.json" });
  desk.createHandoff({ mandateId: mandate.taskpack_id, to: "reviewer", summary: "Ready" });

  const report = desk.verify(mandate.taskpack_id);
  assert.equal(report.ready, true);
  assert.equal(desk.exportReplay(mandate.taskpack_id).artifacts.length, 3);
});

test("loads existing Go-bootstrap workspace config with version key", () => {
  const root = mkdtempSync(join(tmpdir(), "agentdesk-core-legacy-"));
  writeFileSync(
    join(root, "agentdesk.yaml"),
    `version: v1alpha1
workspace: legacy
mission: Existing Go bootstrap config.
defaults:
  max_runtime_minutes: 45
  max_cost_usd: 5
  context_budget_tokens: 12000
task_sources:
  - type: local
    path: .agentdesk/mandates
scope:
  writable:
    - docs/**
  forbidden:
    - .env
success_criteria:
  - A proof artifact is attached.
`
  );
  const config = new AgentDesk(root).loadConfig();
  assert.equal(config.schema_version, "v1alpha1");
  assert.equal(config.workspace, "legacy");
});
