import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { Writable } from "node:stream";
import test from "node:test";
import assert from "node:assert/strict";
import { run } from "../src/index";

test("CLI supports the local mandate/proof/replay loop", async () => {
  const root = mkdtempSync(join(tmpdir(), "agentdesk-cli-"));
  const oldCwd = process.cwd();
  process.chdir(root);
  try {
    await run(["init"], sink());
    const createOut = capture();
    await run(["mandate", "create", "Fix docs", "--writable", "docs/**"], createOut);
    const mandateId = createOut.text.trim().split(/\s+/)[1];
    await run(["claim", "--id", mandateId, "--agent", "codex"], sink());
    writeFileSync("test-results.xml", "<testsuite />\n");
    writeFileSync("changed-files.json", "[\"docs/readme.md\"]\n");
    await run(["proof", "add", "--id", mandateId, "--kind", "test_report", "--path", "test-results.xml"], sink());
    await run(["proof", "add", "--id", mandateId, "--kind", "changed_files", "--path", "changed-files.json"], sink());
    await run(["handoff", "create", "--id", mandateId, "--to", "reviewer", "--summary", "Ready"], sink());
    const verifyOut = capture();
    await run(["verify", "--id", mandateId], verifyOut);
    assert.equal(JSON.parse(verifyOut.text).ready, true);
  } finally {
    process.chdir(oldCwd);
  }
});

function capture(): Writable & { text: string } {
  const chunks: Buffer[] = [];
  const writable = new Writable({
    write(chunk, _encoding, callback) {
      chunks.push(Buffer.from(chunk));
      callback();
    }
  }) as Writable & { text: string };
  Object.defineProperty(writable, "text", {
    get: () => Buffer.concat(chunks).toString("utf8")
  });
  return writable;
}

function sink(): Writable {
  return new Writable({
    write(_chunk, _encoding, callback) {
      callback();
    }
  });
}
